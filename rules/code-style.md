---
description: Code style rules, naming conventions, export patterns, Biome config, comment policy, reusability guidelines, and TypeScript setup. Use when writing or reviewing any code in the project.
alwaysApply: true
---

# Code Style & Configuration

## Naming Conventions
- **React Components** (components/, contexts/, providers/): `PascalCase.tsx`
  - Example: `SearchBar.tsx`, `AuthContext.tsx`, `ThemeProvider.tsx`
- **Next.js App Router** (app/ directory): `snake_case.tsx`
  - Example: `page.tsx`, `layout.tsx`, `not_found.tsx`, `error.tsx`
  - API routes: `route.ts`
- **Hooks** (hooks/): `camelCase.ts`
  - Example: `useDebounce.ts`, `useAddressSearch.ts`
- **Everything else** (lib/, services/, models/, types/, helpers/, data/, scripts/): `camelCase.ts`
  - Example: `supabase.ts`, `mcp.ts`, `date.ts`, `profile.ts`

## Export Rules
- **ALWAYS use end-of-file exports** - Never use inline exports
- Good:
  ```typescript
  const MyComponent = () => { ... }
  export { MyComponent }
  ```
- Bad:
  ```typescript
  export const MyComponent = () => { ... }
  ```

## Biome Configuration
- 2-space indentation
- Single quotes for JavaScript/TypeScript
- 120 character line width
- Auto-organize imports enabled
- Strict linting with recommended rules

## Code Comments
- **NEVER add obvious or redundant comments** — the code should be self-explanatory
- Do not comment what the code does when it is already clear from the code itself
- Only add comments when the logic is non-obvious or requires domain context
- Bad: `// Delete the user` above `db.delete(userTable)`
- Bad: `// Set loading to true` above `setLoading(true)`
- Good: `// Delete sessions first to avoid FK constraint errors` above a delete cascade

## Code Reusability
- **NEVER duplicate code** — always reuse existing functions, helpers, and utilities
- Before writing new code, search for existing implementations across the monorepo:
  - `packages/common/src/helpers/` — shared helpers (date, number, validation)
  - `packages/common/src/hooks/` — shared hooks (useDebounce, useMobile, useAnalytics, useRecaptcha)
  - `packages/common/src/lib/` — shared libs (utils)
  - `apps/app/src/helpers/`, `apps/app/src/hooks/`, `apps/app/src/lib/` — app-specific code
- If code is used by both apps, it belongs in `packages/common/`
- If code is only used by one app, keep it in that app's `src/`
- Bad: Duplicating the same helper in both apps
- Good: Moving shared logic to `packages/common/` and importing via `@project/common/...`

## TypeScript Setup
- Strict mode enabled
- Path mapping: `@/*` points to the app/package root within each workspace
- `@project/common/*` for shared package imports
- Target ES2017 with modern lib support
- Only use arrow functions
