package install

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Jerome1337/ccc/internal/lock"
)

type fakePrompter struct {
	canAsk    bool
	answers   []bool
	questions []string
}

func (prompter *fakePrompter) CanAsk() bool { return prompter.canAsk }

func (prompter *fakePrompter) Confirm(question string) bool {
	prompter.questions = append(prompter.questions, question)
	if len(prompter.answers) == 0 {
		return false
	}
	answer := prompter.answers[0]
	prompter.answers = prompter.answers[1:]

	return answer
}

func content(files map[string]string) fstest.MapFS {
	mapFS := fstest.MapFS{}
	for name, data := range files {
		mapFS[name] = &fstest.MapFile{Data: []byte(data)}
	}

	return mapFS
}

func version1() fstest.MapFS {
	return content(map[string]string{
		"CLAUDE.md": "claude v1\n", "AGENTS.md": "agents v1\n", "RTK.md": "rtk v1\n",
		"CODING_STYLE.md": "style v1\n", "ccc.yml": "plugins: []\n",
		"rules/a.md": "rule a v1\n", "rules/b.md": "rule b v1\n",
		"skills/review/SKILL.md": "skill v1\n",
	})
}

func syncWith(t *testing.T, options Options) (Result, string) {
	t.Helper()
	out := &bytes.Buffer{}
	options.Out = out
	if options.Prompter == nil {
		options.Prompter = &fakePrompter{}
	}
	result, err := Sync(options)
	if err != nil {
		t.Fatalf("Sync() error = %v\n%s", err, out.String())
	}

	return result, out.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func TestManifest(t *testing.T) {
	cases := []struct {
		mode Mode
		want []string
	}{
		{Global, []string{"CLAUDE.md", "RTK.md", "CODING_STYLE.md", "ccc.yml", "skills/review/SKILL.md"}},
		{Project, []string{"CLAUDE.md", "AGENTS.md", "RTK.md", "CODING_STYLE.md", ".claude/rules/a.md", ".claude/rules/b.md", ".claude/skills/review/SKILL.md"}},
	}
	for _, testCase := range cases {
		entries, err := Manifest(version1(), testCase.mode)
		if err != nil {
			t.Fatal(err)
		}
		var dests []string
		for _, entry := range entries {
			dests = append(dests, entry.Dest)
		}
		if !slices.Equal(dests, testCase.want) {
			t.Errorf("Manifest(%v) = %v\nwant %v", testCase.mode, dests, testCase.want)
		}
	}
}

func TestInitThenRerun(t *testing.T) {
	target := t.TempDir()
	result, _ := syncWith(t, Options{Content: version1(), Target: target, Mode: Project, Version: "2.0.0"})
	if result.Changed != 7 {
		t.Errorf("first init changed %d files, want 7", result.Changed)
	}

	result, out := syncWith(t, Options{Content: version1(), Target: target, Mode: Project})
	if result.Changed != 0 || result.Skipped != 0 || strings.Count(out, "ok") != 7 {
		t.Errorf("rerun = %+v, want everything up to date\n%s", result, out)
	}

	current, exists, err := lock.Read(filepath.Join(target, ".claude", lock.FileName))
	if err != nil || !exists || len(current.Files) != 7 {
		t.Fatalf("lock = %+v exists=%v err=%v", current, exists, err)
	}
}

func TestUpdateKeepsEditsAndRemovesDroppedFiles(t *testing.T) {
	target := t.TempDir()
	syncWith(t, Options{Content: version1(), Target: target, Mode: Project})
	if err := os.WriteFile(filepath.Join(target, "CLAUDE.md"), []byte("my own claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	version2 := version1()
	version2["RTK.md"] = &fstest.MapFile{Data: []byte("rtk v2\n")}
	version2["CLAUDE.md"] = &fstest.MapFile{Data: []byte("claude v2\n")}
	delete(version2, "rules/b.md")

	prompter := &fakePrompter{canAsk: true, answers: []bool{false}}
	result, out := syncWith(t, Options{Content: version2, Target: target, Mode: Project, IsUpdate: true, Prompter: prompter})

	if readFile(t, filepath.Join(target, "RTK.md")) != "rtk v2\n" {
		t.Error("unedited RTK.md should update silently")
	}
	if readFile(t, filepath.Join(target, "CLAUDE.md")) != "my own claude\n" {
		t.Error("edited CLAUDE.md was overwritten after answering no")
	}
	if len(prompter.questions) != 1 || !strings.Contains(out, "- my own claude") || !strings.Contains(out, "+ claude v2") {
		t.Errorf("want one prompt with a diff, got questions=%v\n%s", prompter.questions, out)
	}
	if _, err := os.Stat(filepath.Join(target, ".claude", "rules", "b.md")); !os.IsNotExist(err) {
		t.Error("rules/b.md is no longer shipped and unedited, it should be deleted")
	}
	if result.Changed != 2 || result.Skipped != 1 {
		t.Errorf("result = %+v, want 2 changed (RTK.md, b.md) and 1 skipped", result)
	}

	// The same shipped version isn't asked about again.
	prompter = &fakePrompter{canAsk: true}
	_, out = syncWith(t, Options{Content: version2, Target: target, Mode: Project, IsUpdate: true, Prompter: prompter})
	if len(prompter.questions) != 0 || !strings.Contains(out, "keeping yours") {
		t.Errorf("declined version asked again: %v\n%s", prompter.questions, out)
	}

	// --force overwrites it.
	syncWith(t, Options{Content: version2, Target: target, Mode: Project, IsUpdate: true, IsForced: true})
	if readFile(t, filepath.Join(target, "CLAUDE.md")) != "claude v2\n" {
		t.Error("--force should overwrite the edited CLAUDE.md")
	}
}

func TestUpdateWithoutTerminalSkipsEdits(t *testing.T) {
	target := t.TempDir()
	syncWith(t, Options{Content: version1(), Target: target, Mode: Global})
	if err := os.WriteFile(filepath.Join(target, "ccc.yml"), []byte("plugins: [mine]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	version2 := version1()
	version2["ccc.yml"] = &fstest.MapFile{Data: []byte("plugins: [new]\n")}

	_, out := syncWith(t, Options{Content: version2, Target: target, Mode: Global, IsUpdate: true})
	if readFile(t, filepath.Join(target, "ccc.yml")) != "plugins: [mine]\n" || !strings.Contains(out, "in a terminal") {
		t.Errorf("edited ccc.yml should be skipped with a hint\n%s", out)
	}
}

func TestEditedFileNoLongerShippedIsKept(t *testing.T) {
	target := t.TempDir()
	syncWith(t, Options{Content: version1(), Target: target, Mode: Project})
	path := filepath.Join(target, ".claude", "rules", "a.md")
	if err := os.WriteFile(path, []byte("my rule\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	version2 := version1()
	delete(version2, "rules/a.md")

	_, out := syncWith(t, Options{Content: version2, Target: target, Mode: Project, IsUpdate: true})
	if readFile(t, path) != "my rule\n" || !strings.Contains(out, "no longer managed") {
		t.Errorf("edited dropped file should be kept\n%s", out)
	}
	current, _, _ := lock.Read(filepath.Join(target, ".claude", lock.FileName))
	if _, isTracked := current.Files[".claude/rules/a.md"]; isTracked {
		t.Error("a released file should leave the lock")
	}
}

// TestMigrationFromVersion1 uses the real repo files: rules/*.md and the skill are unchanged since 1.x,
// so their hashes are in legacyHashes.
func TestMigrationFromVersion1(t *testing.T) {
	repo := os.DirFS("../..")
	target := t.TempDir()
	for source, dest := range map[string]string{
		"rules/architecture.md":     "rules/architecture.md",
		"rules/patterns.md":         "rules/patterns.md",
		"skills/pr-review/SKILL.md": "skills/pr-review.md",
	} {
		data, err := os.ReadFile(filepath.Join("../..", source))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(target, dest)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, dest), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mine := filepath.Join(target, "rules", "commands.md")
	if err := os.WriteFile(mine, []byte("edited by hand\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	prompter := &fakePrompter{canAsk: true}
	_, out := syncWith(t, Options{Content: repo, Target: target, Mode: Global, IsUpdate: true, Prompter: prompter})

	if len(prompter.questions) != 0 {
		t.Errorf("migration asked %v, want no prompt\n%s", prompter.questions, out)
	}
	for _, gone := range []string{"rules/architecture.md", "rules/patterns.md", "skills/pr-review.md"} {
		if _, err := os.Stat(filepath.Join(target, gone)); !os.IsNotExist(err) {
			t.Errorf("%s should be deleted", gone)
		}
	}
	if readFile(t, mine) != "edited by hand\n" {
		t.Error("an edited 1.x rule must be kept")
	}
	if _, err := os.Stat(filepath.Join(target, "skills", "pr-review", "SKILL.md")); err != nil {
		t.Error("the skill should be installed in its folder")
	}
}

func TestLockPathsOutsideTargetAreIgnored(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "claude")
	outside := filepath.Join(root, "precious.txt")
	if err := os.WriteFile(outside, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(lock.Lock{Schema: 1, Files: map[string]lock.File{"../precious.txt": {SHA256: hash([]byte("keep me\n"))}}})
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, lock.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}

	syncWith(t, Options{Content: version1(), Target: target, Mode: Global, IsUpdate: true})
	if readFile(t, outside) != "keep me\n" {
		t.Fatal("a lock entry outside the target folder deleted a file")
	}
}
