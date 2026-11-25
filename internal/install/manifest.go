// Package install copies the embedded config files and keeps them in sync with the lock file.
package install

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/Jerome1337/ccc/internal/lock"
)

type Mode int

const (
	Project Mode = iota
	Global
)

const (
	projectConfigDir = ".claude"
	rulesDir         = "rules"
	skillsDir        = "skills"
)

var (
	projectFiles = []string{"CLAUDE.md", "AGENTS.md", "RTK.md", "CODING_STYLE.md"}
	// globalFiles leaves out AGENTS.md and the stack-specific rules: they'd load in every project.
	globalFiles = []string{"CLAUDE.md", "RTK.md", "CODING_STYLE.md", "ccc.yml"}
)

// Entry maps an embedded file to its path relative to the target folder.
type Entry struct {
	Source string
	Dest   string
}

// Manifest lists what this version ships for a mode, in a stable order.
func Manifest(content fs.FS, mode Mode) ([]Entry, error) {
	files, prefix := projectFiles, projectConfigDir+"/"
	if mode == Global {
		files, prefix = globalFiles, ""
	}

	var entries []Entry
	for _, file := range files {
		entries = append(entries, Entry{Source: file, Dest: file})
	}

	folders := []string{skillsDir}
	if mode == Project {
		folders = []string{rulesDir, skillsDir}
	}
	for _, folder := range folders {
		nested, err := walk(content, folder, prefix)
		if err != nil {
			return nil, err
		}
		entries = append(entries, nested...)
	}

	return entries, nil
}

func walk(content fs.FS, folder, prefix string) ([]Entry, error) {
	var entries []Entry
	err := fs.WalkDir(content, folder, func(source string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		entries = append(entries, Entry{Source: source, Dest: prefix + source})

		return nil
	})

	return entries, err
}

// GlobalDir follows Claude Code: CLAUDE_CONFIG_DIR when set, ~/.claude otherwise.
func GlobalDir() (string, error) {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()

	return filepath.Join(home, ".claude"), err
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
