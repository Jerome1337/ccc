# ccc

Reusable Claude Code config: `CLAUDE.md`, `RTK.md`, `CODING_STYLE.md`, `AGENTS.md`, rules, a PR review skill, and the MCP servers and plugins you want on every machine. One binary installs it and keeps it up to date.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh
```

This puts `ccc` in `~/.local/bin` (set `CCC_INSTALL_DIR` to pick another folder). It works on macOS and Linux, on amd64 and arm64.

## Usage

```bash
ccc update --global   # your user config in ~/.claude, plus MCP servers and plugins
ccc init              # the current project
ccc update            # the current project, later
```

`ccc update` first updates ccc itself to the latest release, then applies that version's files. Run it in a terminal, so it can ask about files you edited.

## What gets installed

| File | Project (`ccc init`) | Global (`--global`) |
|---|---|---|
| `CLAUDE.md`, `RTK.md`, `CODING_STYLE.md` | `./` | `~/.claude/` |
| `AGENTS.md` | `./` | — |
| `ccc.yml` | — | `~/.claude/` |
| `rules/*.md` | `./.claude/rules/` | — |
| `skills/pr-review/` | `./.claude/skills/` | `~/.claude/skills/` |

The rules are written for a Next.js stack, so they're only installed in projects. In `~/.claude` they would load in every project. With `CLAUDE_CONFIG_DIR` set, `--global` uses that folder instead of `~/.claude`.

## Your edits are kept

ccc records the hash of every file it writes in `ccc-lock.json`: `~/.claude/ccc-lock.json`, or `.claude/ccc-lock.json` in a project. Commit the project one, so teammates get the same behavior.

- A file you never edited is updated silently.
- A file you edited: `ccc update` shows a diff and asks before overwriting it. If you answer no, ccc doesn't ask again until the next version of that file. Without a terminal, and with `ccc init`, it's skipped. `--force` overwrites it without asking.
- A file ccc no longer ships is deleted if you never edited it, and kept otherwise.

## MCP servers and plugins

`--global` also installs the MCP servers and plugins listed in `~/.claude/ccc.yml` for your user, with the `claude` CLI. The ones already installed are left as they are. Edit the file and run `ccc update --global` again to apply it:

```yaml
mcp:
  heroui-react:
    command: npx -y @heroui/react-mcp@latest   # stdio server, arguments split on spaces
  elevenlabs:
    command: uvx elevenlabs-mcp
    env: [ELEVENLABS_API_KEY]                  # values read from your environment
  sentry:
    url: https://mcp.sentry.dev/mcp            # http server

marketplaces:                                  # passed to claude plugin marketplace add
  - anthropics/claude-plugins-official

plugins:                                       # passed to claude plugin install
  - context7@claude-plugins-official
```

The default `ccc.yml` installs `heroui-react` and `elevenlabs` as MCP servers, and `context7`, `stripe`, `code-simplifier`, `frontend-design`, `brag`, `claude-seo`, `ponytail` and `humanizer` as plugins.

A server whose `env` variables are missing is skipped, so `elevenlabs` is only added when `ELEVENLABS_API_KEY` is set. Export it from your shell profile or secret manager rather than typing it on the command line, so it stays out of your shell history. Like any `claude mcp add -e`, the value is then stored in `~/.claude.json`. `elevenlabs` also needs `uv`.

Then, in Claude Code, run `/mcp` to log in to the servers that need it (like the one in the `stripe` plugin) and `/plugin` to set the options a plugin asks for (like `ponytail`). ccc reminds you when it installed something.

## Updating ccc itself

`ccc update` checks the latest GitHub release. When it's newer, ccc downloads the binary for your system, verifies its checksum, checks that it runs, replaces itself and continues with the new version. If anything fails, it keeps the current binary and still updates your files.

- `CCC_NO_SELF_UPDATE=1` turns this off.
- If the folder holding `ccc` isn't writable, ccc tells you to reinstall with the curl command instead of asking for `sudo`.
- A binary downloaded with a browser on macOS is quarantined; run `xattr -d com.apple.quarantine ccc` once. The install script and `ccc update` don't have this problem.

## Coming from the npm version

The `@jerome1337/ccc` npm package is deprecated. Install the binary, then run `ccc update --global`, and `ccc update` in each project that used `npx @jerome1337/ccc init`.

ccc recognizes the files the npm version installed. The ones you never edited are updated, or deleted when they moved: `~/.claude/rules/`, `~/.claude/AGENTS.md` and the old `skills/pr-review.md`. The ones you edited are kept.

## Contents

- **CLAUDE.md**: loaded automatically by Claude Code. Contains security guidelines and imports `RTK.md` and `CODING_STYLE.md`.
- **RTK.md**: instructions for Rust Token Killer, a token-optimized CLI proxy.
- **CODING_STYLE.md**: a coding style that works with any framework (React, Astro, Vue…): language, formatting, naming, functions, types, errors, comments, and size and complexity limits based on SonarQube's TypeScript rules.
- **AGENTS.md**: read by AI coding agents (Claude, Cursor, Copilot…) for project conventions.
- **rules/**: loaded by Claude Code from `.claude/rules/`. They all use `alwaysApply: true`.
- **skills/**: each `skills/<name>/SKILL.md` adds a skill to Claude Code, like `/pr-review`.

## Customization

`CODING_STYLE.md` doesn't depend on the stack. The files in `rules/` are tailored for a Next.js 16 + Turborepo + Bun + Drizzle + Biome + shadcn/ui stack. To adapt them, edit them in your project: ccc keeps your edits.

1. Update `rules/architecture.md` with your project structure
2. Update `rules/commands.md` with your dev commands
3. Update `rules/patterns.md` with your architecture patterns
4. Adjust `rules/code-style.md` naming conventions if different

## Development

```bash
go test -race ./...
go build -ldflags "-X main.version=2.0.0" -o ccc .
```

The config files are embedded in the binary, so a release ships the code and the content together. Conventional Commits on `main` drive releases: semantic-release picks the version, writes `CHANGELOG.md` and tags it, and GoReleaser uploads the binaries and `checksums.txt` to the GitHub release.
