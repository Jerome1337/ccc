package source

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCloneURL(t *testing.T) {
	cases := map[string]string{
		"someone/claude-config":                 "https://github.com/someone/claude-config.git",
		"https://gitlab.com/someone/config.git": "https://gitlab.com/someone/config.git",
		"git@github.com:someone/config.git":     "git@github.com:someone/config.git",
		"--upload-pack=touch /tmp/x":            "--upload-pack=touch /tmp/x",
		"./claude-config":                       "./claude-config",
		"../claude-config":                      "../claude-config",
	}
	for from, want := range cases {
		if got := cloneURL(from); got != want {
			t.Errorf("cloneURL(%q) = %q, want %q", from, got, want)
		}
	}
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFrom(t *testing.T, from, name string) string {
	t.Helper()
	files, closeSource, err := Open(from)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSource()
	data, err := fs.ReadFile(files, name)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func TestOpenLocalFolder(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "global", "CLAUDE.md"), "mine\n")

	if got := readFrom(t, Resolve(repo), "global/CLAUDE.md"); got != "mine\n" {
		t.Errorf("read %q, want the folder's file", got)
	}
	if !filepath.IsAbs(Resolve(".")) {
		t.Error("Resolve(.) should be absolute")
	}
	if got := Resolve("someone/config"); got != "someone/config" {
		t.Errorf("Resolve kept %q, want a missing folder left alone", got)
	}
}

func TestShorthandIsNeverAFolder(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, filepath.Join("acme", "config", "global", "CLAUDE.md"), "planted\n")

	if got := Resolve("acme/config"); got != "acme/config" {
		t.Errorf("Resolve(acme/config) = %q, want the GitHub shorthand kept", got)
	}
	if got := readFrom(t, Resolve("./acme/config"), "global/CLAUDE.md"); got != "planted\n" {
		t.Errorf("read %q, want ./acme/config read as a folder", got)
	}
}

func TestOpenRefusesSymlinkOutsideRepo(t *testing.T) {
	repo, outside := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(outside, "id_rsa"), "secret\n")
	writeFile(t, filepath.Join(repo, "project", "CLAUDE.md"), "ok\n")
	if err := os.Symlink(filepath.Join(outside, "id_rsa"), filepath.Join(repo, "project", "AGENTS.md")); err != nil {
		t.Fatal(err)
	}

	files, closeSource, err := Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSource()
	if data, err := fs.ReadFile(files, "project/AGENTS.md"); err == nil {
		t.Fatalf("read %q through a symlink that leaves the repo", data)
	}
}

func TestOpenClonesGitRepo(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "global", "CLAUDE.md"), "cloned\n")
	for _, args := range [][]string{
		{"init", "--quiet"},
		{"add", "."},
		{"-c", "user.name=ccc", "-c", "user.email=ccc@example.com", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "init"},
	} {
		if output, err := exec.Command(git, append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, output)
		}
	}

	if got := readFrom(t, "file://"+repo, "global/CLAUDE.md"); got != "cloned\n" {
		t.Errorf("read %q, want the committed file", got)
	}
	if _, _, err := Open("file://" + filepath.Join(repo, "missing")); err == nil {
		t.Error("Open() of a missing repo should fail")
	}
}
