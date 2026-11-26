// Package install copies a config repo's files and keeps them in sync with the lock file.
package install

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Jerome1337/ccc/src/lock"
)

type Mode int

const (
	Project Mode = iota
	Global
)

const (
	projectConfigDir = ".claude"
	globalSourceDir  = "global"
	projectSourceDir = "project"
)

// Entry maps a file of the repo to its path relative to the target folder.
type Entry struct {
	Source string
	Dest   string
}

// Manifest maps a config repo's global/ or project/ folder onto the target, file for file, in a stable order.
func Manifest(content fs.FS, mode Mode) ([]Entry, error) {
	folder := SourceDir(mode)
	if _, err := fs.Stat(content, folder); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("the source has no %s/ folder", folder)
	}

	var entries []Entry
	err := fs.WalkDir(content, folder, func(source string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		entries = append(entries, Entry{Source: source, Dest: strings.TrimPrefix(source, folder+"/")})

		return nil
	})
	// An empty folder would make every file of the previous source look dropped, and delete it.
	if err == nil && len(entries) == 0 {
		return nil, fmt.Errorf("the source's %s/ folder is empty", folder)
	}

	return entries, err
}

// SourceDir is the folder of a config repo that holds the files for a mode.
func SourceDir(mode Mode) string {
	if mode == Global {
		return globalSourceDir
	}

	return projectSourceDir
}

// GlobalDir follows Claude Code: CLAUDE_CONFIG_DIR when set, ~/.claude otherwise.
func GlobalDir() (string, error) {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()

	return filepath.Join(home, ".claude"), err
}

// LockedSource is the config repo the last run installed from. It's empty before the first run,
// and in a lock written by a ccc that still had a built-in config.
func LockedSource(target string, mode Mode) (string, error) {
	current, _, err := lock.Read(lockPath(target, mode))

	return current.Source, err
}

// IsManaged reports whether ccc installed files in the target before, from a repo or its old built-in config.
func IsManaged(target string, mode Mode) bool {
	_, exists, err := lock.Read(lockPath(target, mode))

	return err == nil && exists
}

func lockPath(target string, mode Mode) string {
	if mode == Global {
		return filepath.Join(target, lock.FileName)
	}

	return filepath.Join(target, projectConfigDir, lock.FileName)
}

// configRoot is the folder ccc never deletes, even when every file in it is gone.
func configRoot(target string, mode Mode) string {
	if mode == Global {
		return target
	}

	return filepath.Join(target, projectConfigDir)
}

func toOSPath(target, dest string) string {
	return filepath.Join(target, filepath.FromSlash(path.Clean(dest)))
}
