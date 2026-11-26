package lock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMissingFile(t *testing.T) {
	current, exists, err := Read(filepath.Join(t.TempDir(), FileName))
	if err != nil || exists {
		t.Fatalf("Read() exists=%v err=%v, want a missing lock", exists, err)
	}
	if current.Files == nil {
		t.Fatal("Files is nil, want an empty map")
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", FileName)
	written := Lock{Version: "2.0.0", Source: "someone/config", Files: map[string]File{
		"CLAUDE.md": {SHA256: "abc"},
		"ccc.yml":   {SHA256: "def", Declined: "123"},
	}}
	if err := Write(path, written); err != nil {
		t.Fatal(err)
	}

	read, exists, err := Read(path)
	if err != nil || !exists {
		t.Fatalf("Read() exists=%v err=%v", exists, err)
	}
	if read.Version != "2.0.0" || read.Schema != schema || read.Source != "someone/config" || read.Files["ccc.yml"].Declined != "123" {
		t.Errorf("Read() = %+v", read)
	}

	data, _ := os.ReadFile(path)
	if !strings.HasSuffix(string(data), "}\n") {
		t.Error("lock file should end with a newline")
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Errorf("found %d files next to the lock, want no temp file left", len(entries))
	}
}

func TestReadRefusesNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	if err := os.WriteFile(path, []byte(`{"schema": 99, "files": {}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(path); err == nil || !strings.Contains(err.Error(), "newer ccc") {
		t.Fatalf("Read() err = %v, want a newer-schema error", err)
	}
}

func TestReadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(path); err == nil {
		t.Fatal("Read() should fail on invalid JSON")
	}
}
