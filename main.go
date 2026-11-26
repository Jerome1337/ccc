// Command ccc installs and updates a Claude Code config from a git repo, in a project or in ~/.claude.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jerome1337/ccc/src/claude"
	"github.com/Jerome1337/ccc/src/config"
	"github.com/Jerome1337/ccc/src/export"
	"github.com/Jerome1337/ccc/src/install"
	"github.com/Jerome1337/ccc/src/selfupdate"
	"github.com/Jerome1337/ccc/src/source"
	"github.com/Jerome1337/ccc/src/term"
)

// version is set at build time with -ldflags "-X main.version=…". `ccc version` prints it alone
// on one line: the self-update smoke test of older binaries depends on that output.
var version = "dev"

const (
	configFile = "ccc.yml"
	// defaultRepoFolder, in the home folder, is where ccc offers to export a config that has no repo yet.
	defaultRepoFolder = "claude-config"
	exitUsage         = 2
	exitFailure       = 1
	usage             = `Usage: ccc [init|update|export <folder>|version] [--global] [--force] [--from <repo>]

  init      Install the config in the current project (the default)
  update    Update ccc itself, then the config, MCP servers and plugins
  export    Copy the config you already have into <folder>, laid out as a config repo
  version   Print the version
  --global  Use ~/.claude (or $CLAUDE_CONFIG_DIR) and install MCP servers and plugins
  --force   Overwrite files you edited, without asking
  --from    The config repo: owner/repo on GitHub, a git URL or a local folder.
            ccc remembers it, so later updates come from the same repo.
            Without one, ccc offers to export the config on this machine
`
)

var commands = map[string]string{
	"init": "init", "update": "update", "export": "export",
	"version": "version", "--version": "version",
	"help": "help", "--help": "help", "-h": "help",
}

type options struct {
	command  string
	from     string
	folder   string
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
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--global" || arg == "-g":
			opts.isGlobal = true
		case arg == "--force" || arg == "-f":
			opts.isForced = true
		case arg == "--from":
			index++
			if index == len(args) || args[index] == "" {
				return opts, errors.New("--from needs a config repo")
			}
			opts.from = args[index]
		case opts.command == "export" && opts.folder == "" && !strings.HasPrefix(arg, "-"):
			opts.folder = arg
		default:
			command, isCommand := commands[arg]
			if !isCommand || hasCommand {
				return opts, fmt.Errorf("unexpected argument %q", arg)
			}
			opts.command, hasCommand = command, true
		}
	}

	return opts, checkExport(opts)
}

func checkExport(opts options) error {
	if opts.command != "export" {
		return nil
	}
	if opts.folder == "" {
		return errors.New("export needs a folder to copy the config into")
	}
	if opts.from != "" {
		return errors.New("export copies the config on this machine, it takes no --from")
	}

	return nil
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
	if opts.command == "export" {
		return exportConfig(opts, target, mode)
	}

	from, err := chooseSource(opts.from, target, mode)
	if err != nil {
		return err
	}
	if from == "" {
		return offerExport(opts, target, mode)
	}
	files, closeSource, err := openSource(from, mode)
	if err != nil {
		return err
	}
	defer closeSource()
	sync := install.Options{Content: files, Source: from, Target: target, Mode: mode}
	if err := syncFiles(opts, sync); err != nil {
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

// chooseSource is --from, else the repo the lock remembers. It's empty when there's no config repo yet.
func chooseSource(from, target string, mode install.Mode) (string, error) {
	if from != "" {
		return source.Resolve(from), nil
	}

	return install.LockedSource(target, mode)
}

func openSource(from string, mode install.Mode) (fs.FS, func(), error) {
	fmt.Printf("%s %s\n\n", term.Bold("Config"), term.Dim(from))
	if mode == install.Project && filepath.IsAbs(from) {
		fmt.Printf("  %s  %s\n\n", term.Yellow("note  "), term.Dim("a local folder only works on this machine, use a git URL before you commit .claude/ccc-lock.json"))
	}

	return source.Open(from)
}

// offerExport starts a config repo from the config already on this machine, since there's none to install.
func offerExport(opts options, target string, mode install.Mode) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	opts.folder = filepath.Join(home, defaultRepoFolder)
	exported := filepath.Join(opts.folder, install.SourceDir(mode))
	if info, err := os.Stat(exported); err == nil && info.IsDir() {
		fmt.Printf("%s\n\n  %s\n", term.Bold("No config repo yet"), term.Dim(exported+" is ready."))
		printNextSteps(opts.folder, mode)

		return nil
	}

	fmt.Printf("%s\n\n", term.Bold("No config repo yet"))
	if install.IsManaged(target, mode) {
		fmt.Printf("  %s\n\n", term.Dim("These files came from the config older ccc versions had built in. Pass --from <owner/repo> to switch to your repo."))
	}
	prompter := term.NewPrompter(os.Stdin, os.Stdout, term.IsInteractive())
	if !prompter.CanAsk() {
		return errors.New("no config repo yet, pass --from <owner/repo>, or run ccc in a terminal to export the config on this machine")
	}
	if !prompter.Confirm(fmt.Sprintf("  Export the config on this machine to %s?", opts.folder)) {
		fmt.Printf("\n  %s\n\n", term.Dim("Pass --from <owner/repo> to install a config repo, or run ccc export <folder> to make one."))

		return nil
	}
	fmt.Println()

	return exportConfig(opts, target, mode)
}

func syncFiles(opts options, sync install.Options) error {
	fmt.Printf("%s %s\n\n", term.Bold("Files"), term.Dim(sync.Target))
	sync.IsUpdate = opts.command == "update"
	sync.IsForced = opts.isForced
	sync.Prompter = term.NewPrompter(os.Stdin, os.Stdout, term.IsInteractive())
	sync.Out = os.Stdout
	sync.Version = version
	result, err := install.Sync(sync)
	fmt.Printf("\n  %s\n\n", term.Bold(fmt.Sprintf("%d changed, %d kept", result.Changed, result.Skipped)))

	return err
}

func exportConfig(opts options, from string, mode install.Mode) error {
	globalDir, err := install.GlobalDir()
	if err != nil {
		return err
	}
	if mode == install.Project && filepath.Join(from, ".claude") == globalDir {
		return errors.New("this is your home folder, run ccc export --global to export ~/.claude")
	}
	to, err := filepath.Abs(opts.folder)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s → %s\n\n", term.Bold("Export"), term.Dim(from), term.Dim(to))
	copied, err := export.Run(export.Options{From: from, To: to, Mode: mode, IsForced: opts.isForced, Out: os.Stdout})
	if err != nil {
		return err
	}

	fmt.Printf("\n  %s\n", term.Bold(fmt.Sprintf("%d copied", copied)))
	fmt.Printf("  %s\n", term.Dim("settings.json and hooks stay on this machine: settings.json can hold API keys."))
	printNextSteps(to, mode)

	return nil
}

// printNextSteps tells the user how to push the exported folder and install it from there.
func printNextSteps(folder string, mode install.Mode) {
	name := filepath.Base(folder)
	command := "ccc init --from <you>/" + name
	if mode == install.Global {
		command = "ccc update --global --from <you>/" + name
	}

	fmt.Printf("\n  Push it to a repo of your own:\n\n")
	fmt.Printf("    cd %s && git init && git add . && git commit -m \"feat: claude code config\"\n", folder)
	fmt.Printf("    gh repo create %s --private --source . --push\n\n", name)
	fmt.Printf("  Then install it, here and on your other machines:\n\n    %s\n\n", command)
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
