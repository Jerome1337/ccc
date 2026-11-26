// Package lock stores what ccc installed, so an update can tell files you edited from files it wrote.
package lock

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	FileName = "ccc-lock.json"
	// schema 2 added the source. A ccc from before it refuses the lock, so it can't
	// replace the repo's files with the config it had built in.
	schema   = 2
	fileMode = 0o644
)

// File is the last version ccc wrote, and the shipped version you chose to keep yours over.
type File struct {
	SHA256   string `json:"sha256,omitempty"`
	Declined string `json:"declined,omitempty"`
}

type Lock struct {
	Schema  int    `json:"schema"`
	Version string `json:"version"`
	// Source is the config repo the files came from.
	Source string          `json:"source,omitempty"`
	Files  map[string]File `json:"files"`
}

// Read returns an empty lock and exists=false when there's no lock file yet.
func Read(path string) (Lock, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Lock{Schema: schema, Files: map[string]File{}}, false, nil
	}
	if err != nil {
		return Lock{}, false, err
	}

	var current Lock
	if err := json.Unmarshal(data, &current); err != nil {
		return Lock{}, true, fmt.Errorf("%s is not valid JSON: %w", path, err)
	}
	if current.Schema > schema {
		return Lock{}, true, fmt.Errorf("%s was written by a newer ccc, run ccc update first", path)
	}
	if current.Files == nil {
		current.Files = map[string]File{}
	}

	return current, true, nil
}

// Write saves the lock atomically: a crash leaves either the old file or the new one, never half of it.
func Write(path string, current Lock) error {
	current.Schema = schema
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".ccc-lock-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())

	if _, err := temp.Write(append(data, '\n')); err != nil {
		temp.Close()

		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(temp.Name(), fileMode); err != nil {
		return err
	}

	return os.Rename(temp.Name(), path)
}
