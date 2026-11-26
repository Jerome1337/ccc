# ccc

[![Release](https://img.shields.io/github/v/release/Jerome1337/ccc)](https://github.com/Jerome1337/ccc/releases/latest)
[![Build](https://github.com/Jerome1337/ccc/actions/workflows/publish.yml/badge.svg?branch=main)](https://github.com/Jerome1337/ccc/actions/workflows/publish.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/Jerome1337/ccc)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/Jerome1337/ccc)](https://goreportcard.com/report/github.com/Jerome1337/ccc)
[![License: MIT](https://img.shields.io/github/license/Jerome1337/ccc)](LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/Jerome1337/ccc?style=social)](https://github.com/Jerome1337/ccc/stargazers)

A command-line config manager for [Claude Code](https://code.claude.com). Keep your `CLAUDE.md`, rules, skills, agents, MCP servers and plugins in one git repo. One command installs them on a new machine or in a new project, and the same command keeps them up to date. When you have edited a file, ccc shows you the diff and asks before it overwrites it.

```bash
curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh
ccc update --global                             # first run: turns ~/.claude into a repo you push
ccc update --global --from you/claude-config    # on every other machine
```

## Why

A Claude Code setup lives in several places: `~/.claude` for your user, `.claude/` in each repo, and `~/.claude.json` for MCP servers. Getting the same setup on a second laptop, a dev server or a teammate's machine usually means copying files by hand and retyping `claude mcp add` commands, and the copies drift apart as you change one of them.

With ccc, the setup lives in one repo. Every machine that runs `ccc update` installs what the repo holds, so they all end up with the same files. If you don't have a repo yet, the first run offers to make one from the config already on your machine.

- One static Go binary for macOS and Linux, on amd64 and arm64. No Node, no npm.
- Installs for your user (`--global`) or for the current project (`init`).
- Reads your config from a GitHub repo, any URL `git clone` accepts, or a local folder.
- Adds the MCP servers and plugins listed in a YAML file, through the `claude` CLI.
- Records a hash of every file it writes, so it can tell your edits from its own files.
- Updates itself from GitHub releases and checks each download against `checksums.txt`.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh
```

The script puts `ccc` in `~/.local/bin`. Set `CCC_INSTALL_DIR` to pick another folder.

## Commands

| Command | What it does |
|---|---|
| `ccc update --global` | Installs or updates your user config in `~/.claude`, plus the MCP servers and plugins |
| `ccc init` | Installs your project config in the current project |
| `ccc update` | Updates the current project's config |
| `ccc export --global <folder>` | Copies the config in `~/.claude` into `<folder>/global/`, ready to push as a config repo |
| `ccc export <folder>` | Copies the current project's config into `<folder>/project/` |
| `ccc version` | Prints the installed version |

`--from <repo>` picks the config repo, and ccc remembers it for later runs. Without a repo, `init` and `update` offer to export the config on this machine to `~/claude-config` and show you how to push it. `--force` (`-f`) overwrites the files you edited without asking, and `-g` is short for `--global`. With `CLAUDE_CONFIG_DIR` set, `--global` uses that folder instead of `~/.claude`. Run `ccc update` in a terminal, so it can show you diffs and ask about your edits.

## Your config repo

A config repo has two folders, each laid out like the place its files land. Every file in them is installed, and files outside them, like a README, are ignored.

```
claude-config/
├── global/              → ~/.claude
│   ├── CLAUDE.md
│   ├── ccc.yml          MCP servers and plugins, see below
│   ├── agents/
│   └── skills/
└── project/             → the project root
    ├── CLAUDE.md
    ├── AGENTS.md
    └── .claude/
        ├── rules/
        └── settings.json
```

`--from` takes `owner/repo` for GitHub, any URL `git clone` accepts (GitLab, a private repo over SSH), or a local folder such as `./claude-config` or `~/claude-config`, so you can try your config before you push it. ccc fetches repos with `git`, so a private repo works wherever `git clone` does. ccc only reads files inside the repo, and refuses a symlink that points outside it.

The repo is saved in `ccc-lock.json`, so later runs of `ccc init` and `ccc update` fetch it again without `--from`. Commit the project's lock, and your teammates get the same config when they run `ccc update`. A local folder only works on your machine, so switch to a git URL before you commit. Everyone who runs `ccc update` in the project installs what that repo holds, hooks in `settings.json` included, so review a change to the lock's `source` as you would a new dependency.

When you switch to another source, ccc says so, and deletes the files of the old one you never edited. That includes the config older ccc versions had built in, so export it first if you want to keep any of it.

### Start from the setup you have

The first `ccc update --global` or `ccc init` without a repo offers to export your config to `~/claude-config`. To pick another folder, run the export yourself:

```bash
ccc export --global ~/claude-config         # from ~/.claude
cd my-project && ccc export ~/claude-config # from this project
cd ~/claude-config && git init && git add . && git commit -m "feat: claude code config"
gh repo create claude-config --private --source . --push
```

Export copies `CLAUDE.md`, `AGENTS.md`, the files they import with `@path`, and the `agents/`, `commands/`, `skills/`, `rules/` and `output-styles/` folders, from `.claude/` in a project. The global export also copies `ccc.yml`. It skips the skills another tool installs again on each machine: symlinks such as the ones `npx skills` creates, and `skills/synced/`, which Claude Code syncs from claude.ai. A symlinked `CLAUDE.md` or `skills/` folder from your dotfiles is still copied. It also skips dot files such as a skill's `.git` folder, and leaves out `settings.json` (which can hold API keys), hooks, history and caches, so add the ones you want by hand. It won't write into an existing `global/` or `project/` folder unless you pass `--force`.

Once the repo is pushed, run `ccc update --global --from you/claude-config` on each machine, this one included, and `ccc init --from you/claude-config` in your projects.

## Your edits are kept

ccc writes the hash of every file it installs to `ccc-lock.json`, in `~/.claude/` or in the project's `.claude/`. Commit the project one, so your teammates get the same behavior.

- A file you never edited is updated without a prompt.
- For a file you edited, `ccc update` shows the diff and asks before overwriting it. If you say no, it won't ask again until the repo changes that file. `ccc init`, and any run without a terminal, leave it alone.
- A file the repo no longer has is deleted if you never edited it, and kept if you did.

## MCP servers and plugins

`ccc update --global` installs the servers and plugins listed in your repo's `global/ccc.yml` with the `claude` CLI, and leaves the ones you already have alone. To add one, edit the file and run the command again:

```yaml
mcp:
  heroui-react:
    command: npx -y @heroui/react-mcp@latest   # stdio server, arguments split on spaces
  elevenlabs:
    command: uvx elevenlabs-mcp
    env: [ELEVENLABS_API_KEY]                  # read from your environment
  sentry:
    url: https://mcp.sentry.dev/mcp            # http server

marketplaces:                                  # passed to claude plugin marketplace add
  - anthropics/claude-plugins-official

plugins:                                       # passed to claude plugin install
  - context7@claude-plugins-official
```

ccc skips a server while one of its `env` variables is missing, so `elevenlabs` is only added once `ELEVENLABS_API_KEY` is set. Export the key from your shell profile or a secret manager to keep it out of your shell history. As with any `claude mcp add -e`, the value ends up in `~/.claude.json`. `elevenlabs` also needs `uv`.

Some servers need a login and some plugins ask for options. When ccc installed something, it reminds you to run `/mcp` and `/plugin` in Claude Code.

## Self-update

`ccc update` checks the latest GitHub release first. When a newer one exists, ccc downloads the binary for your system, checks it against `checksums.txt`, makes sure it runs, replaces itself and goes on with the new version. If any step fails, it keeps the current binary and still updates your files.

- `CCC_NO_SELF_UPDATE=1` turns this off.
- If the folder that holds `ccc` isn't writable, ccc tells you to rerun the install script. It never asks for `sudo`.
- macOS quarantines a binary downloaded with a browser. Run `xattr -d com.apple.quarantine ccc` once. The install script and `ccc update` aren't affected.

## Contributing

Issues and pull requests are welcome. [CONTRIBUTING.md](CONTRIBUTING.md) covers the setup, the checks CI runs and how releases work. If ccc saves you some setup time, a star helps other Claude Code users find it.

## License

[MIT](LICENSE)
