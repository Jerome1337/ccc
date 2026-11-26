package term

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// /dev/null is a character device but not a terminal: `ccc update </dev/null` must not prompt.
func TestIsTerminalRejectsDevNullAndFiles(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()
	if isTerminal(devNull) {
		t.Error("/dev/null reported as a terminal")
	}

	file, err := os.Create(filepath.Join(t.TempDir(), "file"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if isTerminal(file) {
		t.Error("a regular file reported as a terminal")
	}
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"Yes\n", true},
		{" y \n", true},
		{"n\n", false},
		{"\n", false},
		{"", false},
		{"yep\n", false},
	}
	for _, testCase := range cases {
		prompter := NewPrompter(strings.NewReader(testCase.input), &bytes.Buffer{}, true)
		if got := prompter.Confirm("overwrite?"); got != testCase.want {
			t.Errorf("Confirm(%q) = %v, want %v", testCase.input, got, testCase.want)
		}
	}
}

func TestConfirmReadsOneLinePerQuestion(t *testing.T) {
	prompter := NewPrompter(strings.NewReader("y\nn\ny\n"), &bytes.Buffer{}, true)
	answers := []bool{prompter.Confirm("a"), prompter.Confirm("b"), prompter.Confirm("c")}
	want := []bool{true, false, true}
	for index := range want {
		if answers[index] != want[index] {
			t.Fatalf("answers = %v, want %v", answers, want)
		}
	}
}
