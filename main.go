// Command ccc installs and updates a reusable Claude Code config, in a project or in ~/.claude.
package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Jerome1337/ccc/internal/claude"
	"github.com/Jerome1337/ccc/internal/config"
	"github.com/Jerome1337/ccc/internal/install"
	"github.com/Jerome1337/ccc/internal/selfupdate"
	"github.com/Jerome1337/ccc/internal/term"
)

//go:embed CLAUDE.md AGENTS.md RTK.md CODING_STYLE.md ccc.yml rules/*.md skills
var content embed.FS

// version is set at build time with -ldflags "-X main.version=…". `ccc version` prints it alone
// on one line: the self-update smoke test of older binaries depends on that output.
var version = "dev"

const (
	configFile  = "ccc.yml"
	exitUsage   = 2
	exitFailure = 1
	usage       = `Usage: ccc [init|update|version] [--global] [--force]

  init      Copy the config into the current project (the default)
  update    Update ccc itself, then the config, MCP servers and plugins
  version   Print the version
  --global  Use ~/.claude (or $CLAUDE_CONFIG_DIR) and install MCP servers and plugins
  --force   Overwrite files you edited, without asking
`
)

var commands = map[string]string{
	"init": "init", "update": "update",
	"version": "version", "--version": "version",
	"help": "help", "--help": "help", "-h": "help",
}

type options struct {
	command  string
	isGlobal bool
	isForced bool
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccc: %v\n\n%s", err, usage)
		os.Exit(exitUsage)
	}
	if err := run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "\nccc: %v\n", err)
		os.Exit(exitFailure)
	}
}

func parseArgs(args []string) (options, error) {
	opts := options{command: "init"}
	hasCommand := false
	for _, arg := range args {
		switch arg {
		case "--global", "-g":
			opts.isGlobal = true
		case "--force", "-f":
			opts.isForced = true
		default:
			command, isCommand := commands[arg]
			if !isCommand || hasCommand {
				return opts, fmt.Errorf("unexpected argument %q", arg)
			}
			opts.command, hasCommand = command, true
		}
	}

	return opts, nil
}

func run(opts options) error {
	switch opts.command {
	case "version":
		fmt.Println(version)

		return nil
	case "help":
		fmt.Print(usage)

		return nil
	}

	fmt.Printf("\n%s %s\n\n", term.Bold("ccc"), term.Dim(version))
	if opts.command == "update" {
		selfUpdate()
	}

	target, mode, err := resolveTarget(opts.isGlobal)
	if err != nil {
		return err
	}
	if err := syncFiles(opts, target, mode); err != nil {
		return err
	}
	if opts.isGlobal {
		installExtensions(filepath.Join(target, configFile))
	}

	return nil
}

func selfUpdate() {
	if previous := os.Getenv(selfupdate.UpdatedEnv); previous != "" {
		fmt.Printf("  %s  ccc %s → %s\n", term.Green("update"), previous, version)

		return
	}
	if updater, isEnabled := selfupdate.New(version, os.Stdout); isEnabled {
		updater.Run()
	}
}

func resolveTarget(isGlobal bool) (string, install.Mode, error) {
	if isGlobal {
		dir, err := install.GlobalDir()

		return dir, install.Global, err
	}
	dir, err := os.Getwd()

	return dir, install.Project, err
}

func syncFiles(opts options, target string, mode install.Mode) error {
	fmt.Printf("%s %s\n\n", term.Bold("Files"), term.Dim(target))
	result, err := install.Sync(install.Options{
		Content:  content,
		Target:   target,
		Mode:     mode,
		IsUpdate: opts.command == "update",
		IsForced: opts.isForced,
		Prompter: term.NewPrompter(os.Stdin, os.Stdout, term.IsInteractive()),
		Out:      os.Stdout,
		Version:  version,
	})
	fmt.Printf("\n  %s\n\n", term.Bold(fmt.Sprintf("%d changed, %d kept", result.Changed, result.Skipped)))

	return err
}

// installExtensions never fails the command: a missing claude CLI or a broken ccc.yml only skips this step.
func installExtensions(configPath string) {
	fmt.Printf("%s %s\n\n", term.Bold("MCP servers and plugins"), term.Dim(configPath))
	summary, err := readAndInstall(configPath)
	if err != nil {
		fmt.Printf("  %s  %s\n\n", term.Yellow("skip  "), term.Dim(err.Error()))

		return
	}

	fmt.Printf("\n  %s\n", term.Bold(fmt.Sprintf("%d installed, %d already there", summary.Installed, summary.Present)))
	if summary.Installed > 0 {
		fmt.Printf("  %s\n", term.Dim("In Claude Code, log in to MCP servers with /mcp and set up plugins with /plugin."))
	}
	fmt.Println()
}

func readAndInstall(configPath string) (claude.Summary, error) {
	data, err := os.ReadFile(configPath)
	if errors.Is(err, fs.ErrNotExist) {
		return claude.Summary{}, errors.New("no " + configFile + " yet")
	}
	if err != nil {
		return claude.Summary{}, err
	}
	current, err := config.Parse(data)
	if err != nil {
		return claude.Summary{}, fmt.Errorf("%s: %w", configFile, err)
	}
	cli, err := claude.Find(os.Stdout)
	if err != nil {
		return claude.Summary{}, errors.New("claude not found, install Claude Code first")
	}

	installer := claude.Installer{Runner: cli, LookupEnv: os.LookupEnv, Out: os.Stdout}

	return installer.Install(current)
}
