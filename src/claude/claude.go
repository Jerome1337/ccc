// Package claude installs the MCP servers, marketplaces and plugins from ccc.yml with the claude CLI.
package claude

import (
	"io"
	"os"
	"os/exec"
)

// Runner is the claude CLI; tests replace it with a fake.
type Runner interface {
	// Output captures stdout, for --json listings.
	Output(args ...string) ([]byte, error)
	// Capture captures stdout and stderr together, to read why a command failed.
	Capture(args ...string) ([]byte, error)
	// Run shares the terminal, so claude shows its own progress and prompts.
	Run(args ...string) error
}

type CLI struct {
	path string
	out  io.Writer
}

// Find locates claude on the PATH.
func Find(out io.Writer) (CLI, error) {
	path, err := exec.LookPath("claude")

	return CLI{path: path, out: out}, err
}

func (cli CLI) Output(args ...string) ([]byte, error) {
	return exec.Command(cli.path, args...).Output()
}

func (cli CLI) Capture(args ...string) ([]byte, error) {
	return exec.Command(cli.path, args...).CombinedOutput()
}

func (cli CLI) Run(args ...string) error {
	command := exec.Command(cli.path, args...)
	command.Stdin = os.Stdin
	command.Stdout = cli.out
	command.Stderr = cli.out

	return command.Run()
}
