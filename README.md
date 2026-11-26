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
ccc export --global ~/claude-config             # once: turn the setup you have into a repo, then push it
ccc update --global --from you/claude-config    # on every other machine
```

## Why

A Claude Code setup lives in several places: `~/.claude` for your user, `.claude/` in each repo, and `~/.claude.json` for MCP servers. Getting the same setup on a second laptop, a dev server or a teammate's machine usually means copying files by hand and retyping `claude mcp add` commands, and the copies drift apart as you change one of them.

With ccc, the setup lives in one repo. Every machine that runs `ccc update` installs what the repo holds, so they all end up with the same files. If you don't have a repo yet, `ccc export` makes one from the setup you already have. And with no repo at all, ccc installs the starter config built into each release.

- One static Go binary for macOS and Linux, on amd64 and arm64. No Node, no npm.
- Installs for your user (`--global`) or for the current project (`init`).
- Reads your config from a GitHub repo, any URL `git clone` accepts, or a local folder.
- Adds the MCP servers and plugins listed in a YAML file, through the `claude` CLI.
- Records a hash of every file it writes, so it can tell your edits from its own files.
- Updates itself from GitHub releases and checks each download against `checksums.txt`.
- Comes with a PR review skill that grades every changed file from 1 to 5, so you know which files of a large PR to read.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh
```

The script puts `ccc` in `~/.local/bin`. Set `CCC_INSTALL_DIR` to pick another folder.

## Commands

| Command | What it does |
|---|---|
| `ccc update --global` | Installs or updates your user config in `~/.claude`, plus the MCP servers and plugins |
| `ccc init` | Installs the config in the current project |
| `ccc update` | Updates the current project's config |
| `ccc export --global <folder>` | Copies the config in `~/.claude` into `<folder>/global/`, ready to push as a config repo |
| `ccc export <folder>` | Copies the current project's config into `<folder>/project/` |
| `ccc version` | Prints the installed version |

`--from <repo>` makes `init` and `update` install from your config repo instead of the starter config, and ccc remembers it for later runs. `--force` (`-f`) overwrites the files you edited without asking, and `-g` is short for `--global`. Run `ccc update` in a terminal, so it can show you diffs and ask about your edits.

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

When you switch to another source, ccc says so, and deletes the files of the old one you never edited.

### Start from the setup you have

```bash
ccc export --global ~/claude-config        # from ~/.claude
cd my-project && ccc export ~/claude-config # from this project
cd ~/claude-config && git init && git add . && git commit -m "feat: my claude code config"
```

Export copies `CLAUDE.md`, `AGENTS.md`, the files they import with `@path`, and the `agents/`, `commands/`, `skills/`, `rules/` and `output-styles/` folders, from `.claude/` in a project. The global export also copies `ccc.yml`. It skips the skills another tool installs again on each machine: symlinks such as the ones `npx skills` creates, and `skills/synced/`, which Claude Code syncs from claude.ai. A symlinked `CLAUDE.md` or `skills/` folder from your dotfiles is still copied. It also skips dot files such as a skill's `.git` folder, and leaves out `settings.json` (which can hold API keys), hooks, history and caches, so add the ones you want by hand. It won't write into an existing `global/` or `project/` folder unless you pass `--force`.

Push the folder to GitHub, then run `ccc update --global --from you/claude-config` on your other machines, and `ccc init --from you/claude-config` in your projects.

## The starter config

Without `--from`, ccc installs the config built into each release:

| File | Project (`ccc init`) | Global (`--global`) | What it is |
|---|---|---|---|
| `CLAUDE.md` | `./` | `~/.claude/` | Loaded by Claude Code in every session: security rules, and imports of the next two files |
| `RTK.md` | `./` | `~/.claude/` | How to use RTK (Rust Token Killer), a CLI proxy that cuts the tokens command output costs |
| `CODING_STYLE.md` | `./` | `~/.claude/` | A TypeScript coding style for any framework, with size and complexity limits based on SonarQube's rules |
| `AGENTS.md` | `./` | | Project conventions for Cursor, Copilot and other coding agents |
| `rules/*.md` | `./.claude/rules/` | | Rules for a Next.js 16, Turborepo, Bun, Drizzle, Biome and shadcn/ui stack |
| `skills/pr-review/` | `./.claude/skills/` | `~/.claude/skills/` | The `/pr-review` skill |
| `ccc.yml` | | `~/.claude/` | The MCP servers and plugins to install |

The rules are written for one stack, so ccc only puts them in projects. With `CLAUDE_CONFIG_DIR` set, `--global` uses that folder instead of `~/.claude`.

## Your edits are kept

ccc writes the hash of every file it installs to `ccc-lock.json`, in `~/.claude/` or in the project's `.claude/`. Commit the project one, so your teammates get the same behavior.

- A file you never edited is updated without a prompt.
- For a file you edited, `ccc update` shows the diff and asks before overwriting it. If you say no, it won't ask again until that file changes in a later release. `ccc init`, and any run without a terminal, leave it alone.
- A file that ccc no longer ships is deleted if you never edited it, and kept if you did.

## MCP servers and plugins

`ccc update --global` installs the servers and plugins listed in `~/.claude/ccc.yml` with the `claude` CLI, and leaves the ones you already have alone. To add one, edit the file and run the command again:

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

The default file installs the `heroui-react` and `elevenlabs` MCP servers, and the `context7`, `stripe`, `code-simplifier`, `frontend-design`, `brag`, `claude-seo`, `ponytail` and `humanizer` plugins.

ccc skips a server while one of its `env` variables is missing, so `elevenlabs` is only added once `ELEVENLABS_API_KEY` is set. Export the key from your shell profile or a secret manager to keep it out of your shell history. As with any `claude mcp add -e`, the value ends up in `~/.claude.json`. `elevenlabs` also needs `uv`.

Some servers need a login and some plugins ask for options. When ccc installed something, it reminds you to run `/mcp` and `/plugin` in Claude Code.

## The PR review skill

Run `/pr-review 42` in Claude Code, or `/pr-review` on a branch that has an open PR. The skill reads the diff and grades every changed file from 1 to 5 by how sure it is that you can skip it:

| Grade | On GitHub |
|:-:|---|
| 5 and 4 | Marked as viewed, so they collapse in the PR's file list |
| 3 | Inline notes on the lines worth a glance |
| 2 and 1 | A must-review checklist in the summary comment, with inline comments on the problems it found |

Lockfiles, lint config, docs and assets get a 5 without being read. Secrets, CI, migrations, auth code, git hooks and install-time `package.json` scripts always go on the must-review list. Larger PRs are split across parallel reviewers, and the skill aims to finish in about three minutes. Nothing is posted until you confirm, unless you pass `--auto` for scripts and CI.

## Forking ccc

A config repo covers your own setup. Fork ccc only to change the starter config built into the binary:

1. Fork the repo and replace `CLAUDE.md`, `RTK.md`, `CODING_STYLE.md`, `AGENTS.md`, `rules/`, `skills/` and `ccc.yml` with yours. To drop a file instead, also remove it from the `go:embed` line in `main.go` and from the file lists in `internal/install/manifest.go`.
2. Point `DefaultBaseURL` and `InstallCommand` in `internal/selfupdate/selfupdate.go`, `releases_url` in `install.sh`, and the `curl` lines of this README at your fork.
3. Enable Actions in your fork, since GitHub turns them off on forks. Then push a `feat:` commit to `main`. The release workflow tags a version and uploads the binaries with GitHub's built-in token, so there is no secret to set up.

For small changes you don't need a fork or a repo: edit the installed files, and ccc keeps your edits.

## Self-update

`ccc update` checks the latest GitHub release first. When a newer one exists, ccc downloads the binary for your system, checks it against `checksums.txt`, makes sure it runs, replaces itself and goes on with the new version. If any step fails, it keeps the current binary and still updates your files.

- `CCC_NO_SELF_UPDATE=1` turns this off.
- If the folder that holds `ccc` isn't writable, ccc tells you to rerun the install script. It never asks for `sudo`.
- macOS quarantines a binary downloaded with a browser. Run `xattr -d com.apple.quarantine ccc` once. The install script and `ccc update` aren't affected.

## Coming from the npm version

The `@jerome1337/ccc` npm package is deprecated. Install the binary, run `ccc update --global`, then run `ccc update` in each project where you used `npx @jerome1337/ccc init`.

ccc recognizes the files the npm version installed. It updates the ones you never edited, and deletes the ones that moved: `~/.claude/rules/`, `~/.claude/AGENTS.md` and the old `skills/pr-review.md`. It keeps the ones you edited.

## Development

```bash
go test -race ./...
go build -ldflags "-X main.version=2.0.0" -o ccc .
```

The starter config is embedded in the binary with `go:embed`. Conventional Commits on `main` drive releases:

- [svu](https://github.com/caarlos0/svu) picks the next version from the commits (`fix:` is a patch, `feat:` a minor, `!` or `BREAKING CHANGE:` a major) and tags it.
- [GoReleaser](https://goreleaser.com) uploads the binaries and `checksums.txt` to the GitHub release, with notes grouped from the commits.
- The same notes go to the top of `CHANGELOG.md`.
- `chore:`, `docs:`, `test:` and `ci:` commits don't trigger a release.

Issues and pull requests are welcome. If ccc saves you some setup time, a star helps other Claude Code users find it.

## License

[MIT](LICENSE)
