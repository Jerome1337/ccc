package main

import (
	"slices"
	"testing"

	"github.com/Jerome1337/ccc/internal/install"
)

// TestEmbeddedManifest guards the go:embed patterns: a file left out of them would silently stop shipping.
func TestEmbeddedManifest(t *testing.T) {
	cases := []struct {
		mode install.Mode
		want []string
	}{
		{install.Global, []string{
			"CLAUDE.md", "RTK.md", "CODING_STYLE.md", "ccc.yml",
			"skills/pr-review/SKILL.md", "skills/pr-review/references/grading.md", "skills/pr-review/references/summary.md",
		}},
		{install.Project, []string{
			"CLAUDE.md", "AGENTS.md", "RTK.md", "CODING_STYLE.md",
			".claude/rules/architecture.md", ".claude/rules/code-style.md", ".claude/rules/commands.md", ".claude/rules/patterns.md",
			".claude/skills/pr-review/SKILL.md", ".claude/skills/pr-review/references/grading.md",
			".claude/skills/pr-review/references/summary.md",
		}},
	}
	for _, testCase := range cases {
		entries, err := install.Manifest(content, testCase.mode)
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

func TestParseArgs(t *testing.T) {
	cases := []struct {
		args    []string
		want    options
		isValid bool
	}{
		{nil, options{command: "init"}, true},
		{[]string{"update", "--global"}, options{command: "update", isGlobal: true}, true},
		{[]string{"-g", "-f", "update"}, options{command: "update", isGlobal: true, isForced: true}, true},
		{[]string{"--version"}, options{command: "version"}, true},
		{[]string{"-h"}, options{command: "help"}, true},
		{[]string{"update", "init"}, options{}, false},
		{[]string{"--globl"}, options{}, false},
	}
	for _, testCase := range cases {
		got, err := parseArgs(testCase.args)
		if (err == nil) != testCase.isValid {
			t.Errorf("parseArgs(%v) err = %v", testCase.args, err)
			continue
		}
		if testCase.isValid && got != testCase.want {
			t.Errorf("parseArgs(%v) = %+v, want %+v", testCase.args, got, testCase.want)
		}
	}
}
