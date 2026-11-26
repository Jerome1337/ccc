package export

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Jerome1337/ccc/src/install"
	"github.com/Jerome1337/ccc/src/source"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, data := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func listFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		files = append(files, filepath.ToSlash(relative))

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return files
}

func TestExportGlobal(t *testing.T) {
	home, linked, repo := t.TempDir(), t.TempDir(), t.TempDir()
	writeFiles(t, home, map[string]string{
		"CLAUDE.md": "@RTK.md and @docs/style.md, mail me@example.com, @missing.md, @../outside.md\n",
		"RTK.md":    "rtk\n", "docs/style.md": "see @nested.md\n", "docs/nested.md": "nested\n",
		"ccc.yml": "plugins: []\n", "agents/reviewer.md": "agent\n", "skills/review/SKILL.md": "skill\n",
		"skills/review/.git/HEAD": "ref\n", "skills/.DS_Store": "junk\n",
		"settings.json": `{"env": {"KEY": "secret"}}`, "history.jsonl": "{}\n", "projects/x/session.jsonl": "{}\n",
		"unimported.md": "notes\n", "skills/synced/uuid/pdf/SKILL.md": "synced from claude.ai\n",
	})
	writeFiles(t, linked, map[string]string{"SKILL.md": "linked skill\n"})
	if err := os.Symlink(linked, filepath.Join(home, "skills", "linked")); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	copied, err := Run(Options{From: home, To: repo, Mode: install.Global, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"global/CLAUDE.md", "global/RTK.md", "global/agents/reviewer.md", "global/ccc.yml",
		"global/docs/nested.md", "global/docs/style.md", "global/skills/review/SKILL.md",
	}
	if got := listFiles(t, repo); !slices.Equal(got, want) || copied != len(want) {
		t.Errorf("exported %d files %v\nwant %v", copied, got, want)
	}
	for _, skipped := range []string{"skills/linked", "skills/synced"} {
		if !strings.Contains(out.String(), skipped+" (installed by another tool)") {
			t.Errorf("output should report %s as skipped\n%s", skipped, out.String())
		}
	}

	if _, err := Run(Options{From: home, To: repo, Mode: install.Global, Out: out}); err == nil {
		t.Error("a second export should refuse to write over global/")
	}
	if _, err := Run(Options{From: home, To: repo, Mode: install.Global, IsForced: true, Out: out}); err != nil {
		t.Errorf("--force export = %v", err)
	}
}

// TestExportProjectRoundTrip installs an exported project config with the install manifest, file for file.
func TestExportProjectRoundTrip(t *testing.T) {
	project, repo := t.TempDir(), t.TempDir()
	writeFiles(t, project, map[string]string{
		"CLAUDE.md": "project @AGENTS.md\n", "AGENTS.md": "agents\n", "README.md": "readme\n",
		".claude/rules/stack.md": "stack\n", ".claude/settings.local.json": "{}\n", "src/main.go": "package main\n",
	})
	if _, err := Run(Options{From: project, To: repo, Mode: install.Project, Out: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}

	files, closeSource, err := source.Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSource()
	entries, err := install.Manifest(files, install.Project)
	if err != nil {
		t.Fatal(err)
	}
	var dests []string
	for _, entry := range entries {
		dests = append(dests, entry.Dest)
	}
	if want := []string{".claude/rules/stack.md", "AGENTS.md", "CLAUDE.md"}; !slices.Equal(dests, want) {
		t.Errorf("round trip dests = %v, want %v", dests, want)
	}
}

// TestExportFollowsDotfilesLinks covers a dotfiles setup: CLAUDE.md and the whole skills/ folder are symlinks.
func TestExportFollowsDotfilesLinks(t *testing.T) {
	home, dotfiles, repo := t.TempDir(), t.TempDir(), t.TempDir()
	writeFiles(t, dotfiles, map[string]string{"CLAUDE.md": "mine\n", "skills/own/SKILL.md": "own skill\n"})
	for _, name := range []string{"CLAUDE.md", "skills"} {
		if err := os.Symlink(filepath.Join(dotfiles, name), filepath.Join(home, name)); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := Run(Options{From: home, To: repo, Mode: install.Global, Out: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	if got, want := listFiles(t, repo), []string{"global/CLAUDE.md", "global/skills/own/SKILL.md"}; !slices.Equal(got, want) {
		t.Errorf("exported %v, want %v", got, want)
	}
}

func TestExportNothing(t *testing.T) {
	_, err := Run(Options{From: t.TempDir(), To: t.TempDir(), Mode: install.Project, Out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "no Claude Code config") {
		t.Errorf("Run() err = %v, want nothing to export", err)
	}
}
