# AGENTS.md

Instructions for AI coding agents working on this codebase.

## General Principles

- **Read before writing** — Always read existing code before modifying it. Understand the context, patterns, and conventions already in place.
- **Minimal changes** — Only modify what is necessary. Do not refactor, add comments, or "improve" code that is unrelated to the current task.
- **No over-engineering** — Keep solutions simple. Do not add abstractions, helpers, or utilities for one-time operations. Do not design for hypothetical future requirements.
- **Security first** — Never read, log, or expose `.env` files or secrets. Only reference environment variables by key name. Be careful not to introduce OWASP top 10 vulnerabilities.

## Code Conventions

- **End-of-file exports** — Never use inline exports. Always export at the bottom of the file.
- **Arrow functions only** — Use arrow function expressions, not function declarations.
- **No redundant comments** — Code should be self-explanatory. Only comment when the logic is non-obvious or requires domain context.
- **No code duplication** — Before writing new code, search for existing implementations. Reuse shared helpers and utilities.

## Project Structure

- This is a **Turborepo monorepo** with Bun workspaces
- Shared code lives in `packages/common/`
- App-specific code stays in the respective `apps/*/src/` directory
- Never create `components/` folders inside `src/app/` — components belong in `src/components/`

## File Naming

| Location | Convention | Example |
|---|---|---|
| React components | `PascalCase.tsx` | `SearchBar.tsx` |
| Next.js App Router | `snake_case.tsx` | `page.tsx`, `layout.tsx` |
| Hooks | `camelCase.ts` | `useDebounce.ts` |
| Everything else | `camelCase.ts` | `date.ts`, `supabase.ts` |

## Tech Stack

- **Runtime**: Bun
- **Framework**: Next.js 16 (App Router)
- **Database**: Drizzle ORM
- **Styling**: Tailwind CSS 4 + shadcn/ui
- **Auth**: Better Auth
- **AI**: Anthropic Claude API + MCP
- **Linting**: Biome

## Workflow

1. Prefer server components with direct DB access over client-side API fetches
2. Use `loading.tsx` for loading states — never embed skeletons in `page.tsx`
3. Authenticated routes already handle session checks via `ProtectedRoute`
4. Run `bun run check-types` and `bun run lint` before considering work complete
5. Use `@project/common/...` imports for shared code, `@/src/...` for app-internal imports
