package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jerome1337/ccc/src/install"
)

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
		{[]string{"init", "--from", "someone/config"}, options{command: "init", from: "someone/config"}, true},
		{[]string{"--from"}, options{}, false},
		{[]string{"--from", ""}, options{}, false},
		{[]string{"export", "./repo", "-g"}, options{command: "export", folder: "./repo", isGlobal: true}, true},
		{[]string{"export"}, options{}, false},
		{[]string{"export", "./repo", "./other"}, options{}, false},
		{[]string{"export", "./repo", "--from", "someone/config"}, options{}, false},
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

// TestOfferExport runs without a terminal, as tests do: ccc can't ask, so it explains instead of exporting.
func TestOfferExport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := offerExport(options{command: "init"}, t.TempDir(), install.Project)
	if err == nil || !strings.Contains(err.Error(), "no config repo yet") {
		t.Fatalf("offerExport() err = %v, want the no-repo hint", err)
	}
	if _, err := os.Stat(filepath.Join(home, defaultRepoFolder)); !os.IsNotExist(err) {
		t.Error("nothing should be exported without a terminal to ask in")
	}

	if err := os.MkdirAll(filepath.Join(home, defaultRepoFolder, "project"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := offerExport(options{command: "init"}, t.TempDir(), install.Project); err != nil {
		t.Errorf("offerExport() with an export already there = %v, want the push steps", err)
	}
}
