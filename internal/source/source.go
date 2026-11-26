// Package source fetches a config repo: a local folder, a GitHub owner/repo, or any URL git can clone.
package source

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// gitHubShorthand is owner/repo. An owner can't start with a dot, so ./config stays a local folder.
var gitHubShorthand = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9_.-]+$`)

// Resolve makes a local folder absolute, so the lock still finds it from another directory.
func Resolve(from string) string {
	if !isLocal(from) {
		return from
	}
	absolute, err := filepath.Abs(from)
	if err != nil {
		return from
	}

	return absolute
}

// Open returns the repo's files and a function that deletes the clone once they're copied.
// The files are read through an os.Root, so a symlink in the repo can't expose a file outside it.
func Open(from string) (fs.FS, func(), error) {
	dir, cleanup, err := fetch(from)
	if err != nil {
		return nil, nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		cleanup()

		return nil, nil, err
	}

	return root.FS(), func() { root.Close(); cleanup() }, nil
}

func fetch(from string) (string, func(), error) {
	if isLocal(from) {
		return from, func() {}, nil
	}
	git, err := exec.LookPath("git")
	if err != nil {
		return "", nil, fmt.Errorf("git not found, install it to fetch %s", from)
	}
	dir, err := os.MkdirTemp("", "ccc-source-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(dir) }

	// -- keeps a source that starts with a dash from being read as a git option.
	output, err := exec.Command(git, "clone", "--quiet", "--depth", "1", "--", cloneURL(from), dir).CombinedOutput()
	if err != nil {
		cleanup()

		return "", nil, fmt.Errorf("git clone %s: %s", from, strings.TrimSpace(string(output)))
	}

	return dir, cleanup, nil
}

func cloneURL(from string) string {
	if gitHubShorthand.MatchString(from) {
		return "https://github.com/" + from + ".git"
	}

	return from
}

// isLocal never takes owner/repo as a folder, so a folder with that name can't stand in for the GitHub repo.
func isLocal(from string) bool {
	if gitHubShorthand.MatchString(from) {
		return false
	}
	info, err := os.Stat(from)

	return err == nil && info.IsDir()
}
