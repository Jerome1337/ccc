# ccc

Reusable Claude Code Config, rules, skills, CLAUDE.md, RTK.md, CODING_STYLE.md and AGENTS.md for your projects.

## Contents

```
claude-rules/
├── CLAUDE.md       # Claude Code project instructions (security, RTK and coding style references)
├── RTK.md          # Rust Token Killer instructions (token-optimized CLI proxy)
├── CODING_STYLE.md # Framework-agnostic coding style: naming, structure, size and complexity limits
├── AGENTS.md       # General AI agent instructions (conventions, structure, workflow)
├── rules/          # Claude Code rules (auto-loaded by .claude/rules/)
│   ├── architecture.md   # Monorepo structure, stack, import conventions
│   ├── code-style.md     # Naming, exports, Biome, comments, reusability
│   ├── commands.md       # Dev commands (build, lint, db, turbo)
│   └── patterns.md       # Frontend patterns, auth routes, AI integration
├── skills/         # Claude Code skills (auto-loaded by .claude/skills/)
│   └── pr-review.md      # PR review skill
└── README.md
```

## Usage

```bash
npx @jerome1337/ccc init
```

This copies the following into your project:

| Source | Destination |
|---|---|
| `CLAUDE.md` | `./CLAUDE.md` |
| `AGENTS.md` | `./AGENTS.md` |
| `RTK.md` | `./RTK.md` |
| `CODING_STYLE.md` | `./CODING_STYLE.md` |
| `rules/*.md` | `./.claude/rules/*.md` |
| `skills/*.md` | `./.claude/skills/*.md` |

Existing files are skipped by default. Use `--force` to overwrite:

```bash
npx @jerome1337/ccc init --force
```

To update a project that already has the files, run `update`. It adds new files, skips the ones that are up to date and asks before overwriting each file that differs from the package version. Without a terminal it skips them. `--force` overwrites them without asking:

```bash
npx @jerome1337/ccc@latest update
```

Add `--global` to `init` or `update` to use `~/.claude/` instead of the current project. `AGENTS.md` is not copied there, `ccc.yml` is, and rules and skills go to `~/.claude/rules/` and `~/.claude/skills/`.

`--global` also installs the MCP servers and plugins listed in `~/.claude/ccc.yml` for your user with the `claude` CLI. The ones already installed are left as they are. Edit the file and run `ccc update --global` again to apply it:

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

The default `ccc.yml` installs `heroui-react` and `elevenlabs` as MCP servers, and `context7`, `stripe`, `code-simplifier`, `frontend-design`, `brag`, `claude-seo` and `ponytail` as plugins.

A server whose `env` variables are missing is skipped, so `elevenlabs` is only added when `ELEVENLABS_API_KEY` is set. Export it from your shell profile or secret manager rather than typing it on the command line, so it stays out of your shell history. Like any `claude mcp add -e`, the value is then stored in `~/.claude.json`. `elevenlabs` also needs `uv`.

`stripe` asks you to log in the first time you use it in Claude Code (`/mcp`).

## How It Works

- **CLAUDE.md** — Loaded automatically by Claude Code at the project root. Contains security guidelines and imports RTK.md and CODING_STYLE.md.
- **CODING_STYLE.md** — Personal coding style that works with any framework (React, Astro, Vue…): language, formatting, naming, functions, types, errors, comments, and size and complexity limits based on SonarQube's TypeScript rules. Framework-specific rules sit in their own sections at the end. It can also be used globally: run `ccc init --global` and add `@CODING_STYLE.md` to `~/.claude/CLAUDE.md` if that file already existed.
- **AGENTS.md** — Read by AI coding agents (Claude, Cursor, Copilot, etc.) for project conventions.
- **rules/** — Files in `.claude/rules/` are automatically loaded by Claude Code based on their `alwaysApply` / `globs` frontmatter. All rules here use `alwaysApply: true`.
- **skills/** — Files in `.claude/skills/` define custom slash commands for Claude Code. Invoke with `/skill-name` (e.g., `/pr-review`).

## Customization

`CODING_STYLE.md` doesn't depend on the stack and needs no changes. The files in `rules/` are tailored for a Next.js 16 + Turborepo + Bun + Drizzle + Biome + shadcn/ui stack. To adapt them for a different project:

1. Update `rules/architecture.md` with your project structure
2. Update `rules/commands.md` with your dev commands
3. Update `rules/patterns.md` with your architecture patterns
4. Adjust `rules/code-style.md` naming conventions if different
