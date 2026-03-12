# ccc

Reusable Claude Code Config, rules, skills, CLAUDE.md, RTK.md and AGENTS.md for your projects.

## Contents

```
claude-rules/
├── CLAUDE.md       # Claude Code project instructions (security, RTK reference)
├── RTK.md          # Rust Token Killer instructions (token-optimized CLI proxy)
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
| `rules/*.md` | `./.claude/rules/*.md` |
| `skills/*.md` | `./.claude/skills/*.md` |

Existing files are skipped by default. Use `--force` to overwrite:

```bash
npx @jerome1337/ccc init --force
```

## How It Works

- **CLAUDE.md** — Loaded automatically by Claude Code at the project root. Contains security guidelines and references RTK.md.
- **AGENTS.md** — Read by AI coding agents (Claude, Cursor, Copilot, etc.) for project conventions.
- **rules/** — Files in `.claude/rules/` are automatically loaded by Claude Code based on their `alwaysApply` / `globs` frontmatter. All rules here use `alwaysApply: true`.
- **skills/** — Files in `.claude/skills/` define custom slash commands for Claude Code. Invoke with `/skill-name` (e.g., `/pr-review`).

## Customization

These rules are tailored for a Next.js 16 + Turborepo + Bun + Drizzle + Biome + shadcn/ui stack. To adapt for a different project:

1. Update `rules/architecture.md` with your project structure
2. Update `rules/commands.md` with your dev commands
3. Update `rules/patterns.md` with your architecture patterns
4. Adjust `rules/code-style.md` naming conventions if different
