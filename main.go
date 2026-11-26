// Command ccc installs and updates a reusable Claude Code config, in a project or in ~/.claude.
package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jerome1337/ccc/internal/claude"
	"github.com/Jerome1337/ccc/internal/config"
	"github.com/Jerome1337/ccc/internal/export"
	"github.com/Jerome1337/ccc/internal/install"
	"github.com/Jerome1337/ccc/internal/selfupdate"
	"github.com/Jerome1337/ccc/internal/source"
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
	usage       = `Usage: ccc [init|update|export <folder>|version] [--global] [--force] [--from <repo>]

  init      Copy the config into the current project (the default)
  update    Update ccc itself, then the config, MCP servers and plugins
  export    Copy the config you already have into <folder>, laid out as a config repo
  version   Print the version
  --global  Use ~/.claude (or $CLAUDE_CONFIG_DIR) and install MCP servers and plugins
  --force   Overwrite files you edited, without asking
  --from    Install from a config repo: owner/repo on GitHub, a git URL or a local folder.
            ccc remembers it, so later updates come from the same repo
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

// configSource is the config to install: a fetched repo, or the files built into ccc when name is empty.
type configSource struct {
	files fs.FS
	name  string
	close func()
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

	repo, err := openSource(opts.from, target, mode)
	if err != nil {
		return err
	}
	defer repo.close()
	sync := install.Options{Content: repo.files, Source: repo.name, Target: target, Mode: mode}
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

// openSource picks the config to install: --from, else the repo the lock remembers, else the built-in one.
func openSource(from, target string, mode install.Mode) (configSource, error) {
	from = source.Resolve(from)
	if from == "" {
		locked, err := install.LockedSource(target, mode)
		if err != nil {
			return configSource{}, err
		}
		from = locked
	}
	if from == "" {
		return configSource{files: content, close: func() {}}, nil
	}

	fmt.Printf("%s %s\n\n", term.Bold("Config"), term.Dim(from))
	if mode == install.Project && filepath.IsAbs(from) {
		fmt.Printf("  %s  %s\n\n", term.Yellow("note  "), term.Dim("a local folder only works on this machine, use a git URL before you commit .claude/ccc-lock.json"))
	}
	files, closeSource, err := source.Open(from)

	return configSource{files: files, name: from, close: closeSource}, err
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

	next := "ccc init --from <you>/<repo> in your projects"
	if mode == install.Global {
		next = "ccc update --global --from <you>/<repo> on your other machines"
	}
	fmt.Printf("\n  %s\n", term.Bold(fmt.Sprintf("%d copied", copied)))
	fmt.Printf("  %s\n", term.Dim("settings.json and hooks stay on this machine: settings.json can hold API keys."))
	fmt.Printf("  %s\n\n", term.Dim(fmt.Sprintf("Push %s to GitHub, then run %s.", to, next)))

	return nil
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
