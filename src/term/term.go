// Package term handles colors, terminal detection and y/N prompts.
package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	codeBold   = "1"
	codeDim    = "2"
	codeRed    = "31"
	codeGreen  = "32"
	codeYellow = "33"
	codeCyan   = "36"
)

// hasColor is decided once: NO_COLOR (https://no-color.org) and redirected output both disable colors.
var hasColor = os.Getenv("NO_COLOR") == "" && isTerminal(os.Stdout)

func paint(code, text string) string {
	if !hasColor {
		return text
	}

	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func Bold(text string) string   { return paint(codeBold, text) }
func Dim(text string) string    { return paint(codeDim, text) }
func Red(text string) string    { return paint(codeRed, text) }
func Green(text string) string  { return paint(codeGreen, text) }
func Yellow(text string) string { return paint(codeYellow, text) }
func Cyan(text string) string   { return paint(codeCyan, text) }

// IsInteractive reports whether stdin is a terminal, so a prompt can be answered.
func IsInteractive() bool {
	return isTerminal(os.Stdin)
}

// Prompter asks y/N questions. A single reader is kept so piped answers are read one line at a time.
type Prompter struct {
	reader        *bufio.Reader
	out           io.Writer
	isInteractive bool
}

func NewPrompter(in io.Reader, out io.Writer, isInteractive bool) *Prompter {
	return &Prompter{reader: bufio.NewReader(in), out: out, isInteractive: isInteractive}
}

func (prompter *Prompter) CanAsk() bool {
	return prompter.isInteractive
}

// Confirm returns true only for an explicit yes. A read error or end of input counts as no.
func (prompter *Prompter) Confirm(question string) bool {
	fmt.Fprintf(prompter.out, "%s %s ", question, Dim("(y/N)"))
	answer, err := prompter.reader.ReadString('\n')
	if err != nil && answer == "" {
		fmt.Fprintln(prompter.out)

		return false
	}
	answer = strings.ToLower(strings.TrimSpace(answer))

	return answer == "y" || answer == "yes"
}
