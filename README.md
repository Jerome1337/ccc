# claude-rules

Reusable Claude Code rules, CLAUDE.md, and AGENTS.md for your projects.

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
└── README.md
```

## Usage

### Option 1: Symlink into your project

```bash
# From your project root
ln -s /path/to/claude-rules/CLAUDE.md ./CLAUDE.md
ln -s /path/to/claude-rules/AGENTS.md ./AGENTS.md
ln -s /path/to/claude-rules/rules ./.claude/rules
```

### Option 2: Copy into your project

```bash
# From your project root
cp /path/to/claude-rules/CLAUDE.md ./CLAUDE.md
cp /path/to/claude-rules/AGENTS.md ./AGENTS.md
mkdir -p .claude/rules
cp /path/to/claude-rules/rules/*.md ./.claude/rules/
```

### Option 3: Git submodule

```bash
git submodule add https://github.com/<your-username>/claude-rules .claude-rules
ln -s .claude-rules/CLAUDE.md ./CLAUDE.md
ln -s .claude-rules/AGENTS.md ./AGENTS.md
ln -s .claude-rules/rules ./.claude/rules
```

## How It Works

- **CLAUDE.md** — Loaded automatically by Claude Code at the project root. Contains security guidelines and references RTK.md.
- **AGENTS.md** — Read by AI coding agents (Claude, Cursor, Copilot, etc.) for project conventions.
- **rules/** — Files in `.claude/rules/` are automatically loaded by Claude Code based on their `alwaysApply` / `globs` frontmatter. All rules here use `alwaysApply: true`.

## Customization

These rules are tailored for a Next.js 16 + Turborepo + Bun + Drizzle + Biome + shadcn/ui stack. To adapt for a different project:

1. Update `rules/architecture.md` with your project structure
2. Update `rules/commands.md` with your dev commands
3. Update `rules/patterns.md` with your architecture patterns
4. Adjust `rules/code-style.md` naming conventions if different
