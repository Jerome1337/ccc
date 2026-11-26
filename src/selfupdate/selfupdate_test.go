package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const asset = "ccc_test_arch"

// release serves /releases/latest as a redirect and the assets of one tag, like GitHub does.
type release struct {
	tag       string
	binary    string
	checksums string
	hasBinary bool
}

func fakeBinary(version string) string {
	return "#!/bin/sh\necho " + version + "\n"
}

func newRelease(tag string) release {
	binary := fakeBinary(strings.TrimPrefix(tag, "v"))
	sum := sha256.Sum256([]byte(binary))
	checksums := fmt.Sprintf("%s  %s\n%s  ccc_other_arch\n", hex.EncodeToString(sum[:]), asset, strings.Repeat("0", 64))

	return release{tag: tag, binary: binary, checksums: checksums, hasBinary: true}
}

func (current release) serve(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/releases/latest":
			http.Redirect(writer, request, "/releases/tag/"+current.tag, http.StatusFound)
		case "/releases/download/" + current.tag + "/checksums.txt":
			io.WriteString(writer, current.checksums)
		case "/releases/download/" + current.tag + "/" + asset:
			if !current.hasBinary {
				http.NotFound(writer, request)

				return
			}
			io.WriteString(writer, current.binary)
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

func installedBinary(t *testing.T, version string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ccc")
	if err := os.WriteFile(path, []byte(fakeBinary(version)), 0o755); err != nil {
		t.Fatal(err)
	}

	return path
}

func updaterFor(server *httptest.Server, executable, current string) Updater {
	return Updater{BaseURL: server.URL, Current: current, Executable: executable, Asset: asset, Out: io.Discard}
}

func assertUntouched(t *testing.T, executable, version string) {
	t.Helper()
	data, err := os.ReadFile(executable)
	if err != nil || string(data) != fakeBinary(version) {
		t.Errorf("the installed binary changed: %q, %v", data, err)
	}
	entries, _ := os.ReadDir(filepath.Dir(executable))
	if len(entries) != 1 {
		t.Errorf("%d files next to the binary, want no temp file left", len(entries))
	}
}

func TestUpdateReplacesTheBinary(t *testing.T) {
	server := newRelease("v2.0.1").serve(t)
	executable := installedBinary(t, "2.0.0")

	latest, err := updaterFor(server, executable, "2.0.0").Update()
	if err != nil || latest != "2.0.1" {
		t.Fatalf("Update() = %q, %v", latest, err)
	}
	data, _ := os.ReadFile(executable)
	if string(data) != fakeBinary("2.0.1") {
		t.Errorf("binary = %q, want the 2.0.1 one", data)
	}
	info, _ := os.Stat(executable)
	if info.Mode().Perm() != binaryMode {
		t.Errorf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(binaryMode))
	}
}

func TestUpdateWhenUpToDate(t *testing.T) {
	server := newRelease("v2.0.0").serve(t)
	executable := installedBinary(t, "2.0.0")

	latest, err := updaterFor(server, executable, "2.0.0").Update()
	if err != nil || latest != "" {
		t.Fatalf("Update() = %q, %v, want nothing to do", latest, err)
	}
	assertUntouched(t, executable, "2.0.0")
}

func TestUpdateFailuresKeepTheOldBinary(t *testing.T) {
	cases := []struct {
		name   string
		modify func(*release)
		want   string
	}{
		{"checksum mismatch", func(current *release) { current.binary = fakeBinary("2.0.1") + "tampered\n" }, "checksum mismatch"},
		{"no binary for the platform", func(current *release) { current.checksums = "" }, "no binary"},
		{"asset missing", func(current *release) { current.hasBinary = false }, "no binary"},
		{"binary reports the wrong version", func(current *release) {
			current.binary = fakeBinary("9.9.9")
			sum := sha256.Sum256([]byte(current.binary))
			current.checksums = hex.EncodeToString(sum[:]) + "  " + asset + "\n"
		}, "reports version"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			current := newRelease("v2.0.1")
			testCase.modify(&current)
			server := current.serve(t)
			executable := installedBinary(t, "2.0.0")

			_, err := updaterFor(server, executable, "2.0.0").Update()
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("Update() err = %v, want %q", err, testCase.want)
			}
			assertUntouched(t, executable, "2.0.0")
		})
	}
}

func TestUpdateInReadOnlyFolder(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores folder permissions")
	}
	server := newRelease("v2.0.1").serve(t)
	executable := installedBinary(t, "2.0.0")
	folder := filepath.Dir(executable)
	if err := os.Chmod(folder, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(folder, 0o755) })

	_, err := updaterFor(server, executable, "2.0.0").Update()
	if err == nil || !strings.Contains(err.Error(), "reinstall with") {
		t.Fatalf("Update() err = %v, want the reinstall hint", err)
	}
	assertUntouched(t, executable, "2.0.0")
}

func TestUpdateWithoutRelease(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	if _, err := updaterFor(server, installedBinary(t, "2.0.0"), "2.0.0").Update(); err == nil {
		t.Fatal("Update() should fail when there's no release")
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		tag, current string
		want         bool
	}{
		{"v2.0.1", "2.0.0", true},
		{"v2.0.10", "2.0.9", true},
		{"v3.0.0", "2.9.9", true},
		{"v2.0.0", "2.0.0", false},
		{"v1.9.9", "2.0.0", false},
		{"v2.0.0", "dev", false},
		{"latest", "2.0.0", false},
		{"v2.0", "1.0.0", false},
	}
	for _, testCase := range cases {
		if got := isNewer(testCase.tag, testCase.current); got != testCase.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", testCase.tag, testCase.current, got, testCase.want)
		}
	}
}

func TestNewSkipsDevAndLoopGuard(t *testing.T) {
	if _, ok := New("dev", io.Discard); ok {
		t.Error("a dev build shouldn't update itself")
	}
	t.Setenv(UpdatedEnv, "2.0.0")
	if _, ok := New("2.0.1", io.Discard); ok {
		t.Error("a re-executed binary shouldn't update again")
	}
}

func TestNewRejectsPlainHTTPOverride(t *testing.T) {
	t.Setenv(baseURLEnv, "http://example.com")
	updater, ok := New("2.0.0", io.Discard)
	if !ok || updater.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want the default for a plain-http override", updater.BaseURL)
	}
}
