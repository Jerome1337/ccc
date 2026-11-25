// Package selfupdate replaces the ccc binary with the latest GitHub release, then re-runs it.
package selfupdate

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Jerome1337/ccc/internal/term"
)

const (
	DefaultBaseURL = "https://github.com/Jerome1337/ccc"
	InstallCommand = "curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh"

	// UpdatedEnv carries the previous version into the re-executed binary, and stops a second update.
	UpdatedEnv  = "CCC_SELF_UPDATED"
	disableEnv  = "CCC_NO_SELF_UPDATE"
	baseURLEnv  = "CCC_RELEASES_URL"
	devVersion  = "dev"
	localPrefix = "http://127.0.0.1"

	checkTimeout    = 10 * time.Second
	downloadTimeout = 2 * time.Minute
	smokeTimeout    = 10 * time.Second
	maxBinarySize   = 100 << 20
	binaryMode      = 0o755
)

var errNoBinary = errors.New("no binary for this platform in the latest release")

type Updater struct {
	BaseURL    string
	Current    string
	Executable string
	Asset      string
	Out        io.Writer
}

// New reads the environment; it returns ok=false when this binary shouldn't update itself.
func New(current string, out io.Writer) (Updater, bool) {
	if current == devVersion || os.Getenv(disableEnv) == "1" || os.Getenv(UpdatedEnv) != "" {
		return Updater{}, false
	}

	baseURL := DefaultBaseURL
	override := os.Getenv(baseURLEnv)
	if strings.HasPrefix(override, "https://") || strings.HasPrefix(override, localPrefix) {
		baseURL = strings.TrimSuffix(override, "/")
	}
	executable, err := os.Executable()
	if err == nil {
		executable, err = filepath.EvalSymlinks(executable)
	}
	if err != nil {
		return Updater{}, false
	}

	asset := fmt.Sprintf("ccc_%s_%s", runtime.GOOS, runtime.GOARCH)

	return Updater{BaseURL: baseURL, Current: current, Executable: executable, Asset: asset, Out: out}, true
}

// Run updates and re-executes the binary. It only returns when there's nothing to do or the update
// failed, after printing why, so the config sync can go on with the current binary.
func (updater Updater) Run() {
	latest, err := updater.Update()
	if err != nil {
		updater.print(term.Yellow("skip  "), "ccc self-update", err.Error())

		return
	}
	if latest == "" {
		updater.print(term.Dim("ok    "), "ccc "+updater.Current, "latest")

		return
	}

	args := append([]string{updater.Executable}, os.Args[1:]...)
	env := append(os.Environ(), UpdatedEnv+"="+updater.Current)
	err = syscall.Exec(updater.Executable, args, env)
	updater.print(term.Yellow("note  "), "ccc "+latest+" installed, run the command again", err.Error())
}

func (updater Updater) print(label, subject, detail string) {
	fmt.Fprintf(updater.Out, "  %s  %s %s\n", label, subject, term.Dim("("+detail+")"))
}

// Update installs a newer release over the executable and returns its version, or "" when up to date.
func (updater Updater) Update() (string, error) {
	tag, err := updater.latestTag()
	if err != nil {
		return "", err
	}
	if !isNewer(tag, updater.Current) {
		return "", nil
	}

	expected, err := updater.checksum(tag)
	if err != nil {
		return "", err
	}
	staged, err := updater.stage(tag, expected)
	if err != nil {
		return "", err
	}
	if err := os.Rename(staged, updater.Executable); err != nil {
		os.Remove(staged)

		return "", err
	}

	return strings.TrimPrefix(tag, "v"), nil
}

// latestTag reads the redirect of /releases/latest instead of the API: no rate limit, no JSON.
func (updater Updater) latestTag() (string, error) {
	client := &http.Client{
		Timeout: checkTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := client.Get(updater.BaseURL + "/releases/latest")
	if err != nil {
		return "", fmt.Errorf("can't reach GitHub: %w", err)
	}
	defer response.Body.Close()

	location := response.Header.Get("Location")
	if response.StatusCode/100 != 3 || location == "" {
		return "", fmt.Errorf("no release found (%s)", response.Status)
	}

	return path.Base(location), nil
}

func (updater Updater) checksum(tag string) (string, error) {
	body, err := updater.download(tag, "checksums.txt")
	if err != nil {
		return "", err
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == updater.Asset {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errNoBinary
}

// stage downloads next to the executable, so the final rename stays on one filesystem and is atomic.
func (updater Updater) stage(tag, expected string) (string, error) {
	body, err := updater.download(tag, updater.Asset)
	if err != nil {
		return "", err
	}
	defer body.Close()

	staged, err := os.CreateTemp(filepath.Dir(updater.Executable), ".ccc-update-*")
	if err != nil {
		return "", fmt.Errorf("can't write to %s, reinstall with: %s", filepath.Dir(updater.Executable), InstallCommand)
	}
	isStaged := false
	defer func() {
		if !isStaged {
			os.Remove(staged.Name())
		}
	}()

	digest := sha256.New()
	_, err = io.Copy(io.MultiWriter(staged, digest), body)
	err = errors.Join(err, staged.Close())
	if err != nil {
		return "", err
	}
	if hex.EncodeToString(digest.Sum(nil)) != expected {
		return "", fmt.Errorf("checksum mismatch for %s", updater.Asset)
	}
	if err := os.Chmod(staged.Name(), binaryMode); err != nil {
		return "", err
	}
	if err := smokeTest(staged.Name(), strings.TrimPrefix(tag, "v")); err != nil {
		return "", err
	}
	isStaged = true

	return staged.Name(), nil
}

func (updater Updater) download(tag, asset string) (io.ReadCloser, error) {
	client := &http.Client{Timeout: downloadTimeout}
	response, err := client.Get(updater.BaseURL + "/releases/download/" + tag + "/" + asset)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusNotFound {
		response.Body.Close()

		return nil, errNoBinary
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()

		return nil, fmt.Errorf("download of %s failed (%s)", asset, response.Status)
	}

	return struct {
		io.Reader
		io.Closer
	}{io.LimitReader(response.Body, maxBinarySize), response.Body}, nil
}

// smokeTest runs the new binary before it replaces the old one: a wrong architecture, a truncated
// download or a binary macOS refuses to run fails here instead of breaking ccc.
func smokeTest(binary, want string) error {
	ctx, cancel := context.WithTimeout(context.Background(), smokeTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, binary, "version").Output()
	if err != nil {
		return fmt.Errorf("the new binary doesn't run: %w", err)
	}
	if got := string(bytes.TrimSpace(output)); got != want {
		return fmt.Errorf("the new binary reports version %q, want %q", got, want)
	}

	return nil
}

// isNewer compares vX.Y.Z tags; anything that doesn't parse is never newer.
func isNewer(tag, current string) bool {
	latest, isLatestValid := parseVersion(tag)
	running, isRunningValid := parseVersion(current)
	if !isLatestValid || !isRunningValid {
		return false
	}
	for index := range latest {
		if latest[index] != running[index] {
			return latest[index] > running[index]
		}
	}

	return false
}

func parseVersion(version string) ([3]int, bool) {
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}

	var numbers [3]int
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return [3]int{}, false
		}
		numbers[index] = number
	}

	return numbers, true
}
