---
description: Architecture patterns including frontend structure rules, authenticated routes, AI integration, data flow, component patterns, and environment setup. Use when building pages, components, or API routes.
alwaysApply: true
---

# Architecture Patterns

## Frontend Structure
- **NEVER create `components/` folders inside `src/app/`** — all React components must live in `src/components/`, organized by feature. The `src/app/` directory should only contain Next.js route files (`page.tsx`, `layout.tsx`, `route.ts`, etc.) and server actions (`actions/`).
- **Use Next.js `loading.tsx` for loading/skeleton states** — never embed loading skeletons directly in `page.tsx`. Instead, create a `loading.tsx` file in the same route folder.

## Authenticated Routes (`apps/app/src/app/(authenticated)/`)
- The `(authenticated)` layout wraps all child pages with `AuthProvider` + `ProtectedRoute`, which handles session checks and redirects unauthenticated users to `/login`
- **Pages inside this folder do NOT need to check or redirect for unauthenticated users** — `ProtectedRoute` handles this automatically
- Pages still need to call `auth.api.getSession({ headers: await headers() })` to identify the current user (for DB queries), but can assume the user is authenticated
- **Prefer server components with direct DB access** (via Drizzle ORM queries from `src/db/queries/`) over client-side API fetches (`useEffect` + `fetch('/api/...')`)
- Reference implementations: `settings/profile/page.tsx`, `settings/profile/edit/page.tsx`, `profile/edit/page.tsx`

## AI Integration (`apps/app/src/app/api/search/route.ts`)
- Uses Anthropic Claude Sonnet 4 API for property search analysis
- Integrates Model Context Protocol (MCP) for external real estate data
- Environment variables required: `ANTHROPIC_API_KEY`, `ESTATE_ADS_MCP_SERVER`
- Returns structured JSON responses for property listings and market analysis

## Data Flow
1. User submits search query through main interface
2. Frontend sends POST request to `/search` API route
3. API route processes query using Claude + MCP server
4. Returns structured property data with pros/cons analysis
5. Frontend renders property cards with market insights

## Component Patterns
- Uses shadcn/ui for consistent design system (components live in `packages/common/`)
- Lucide React for iconography
- Custom utility function `cn()` for className merging (clsx + tailwind-merge) from `@project/common/lib/utils`
- Responsive design with mobile-first approach
- French language interface with localized formatting

## Environment Setup
- `.env` file lives at the **monorepo root** and is loaded by each app via `@next/env` in `next.config.js`
- Each app has its own `src/env.ts` with `@t3-oss/env-nextjs` validation scoped to the env vars it needs
- `apps/app/src/env.ts` — full env validation (AI, DB, auth, analytics, etc.)
- `apps/landing/src/env.ts` — minimal env validation (analytics, reCAPTCHA, Google Sheets)
- Landing runs on port 3000, App runs on port 3001

## Key Features
- Multi-step loading animation with property search progress
- Property cards with detailed analysis (pros/cons, market data)
- AI-powered zone recommendations with comparative analysis
- Animated skeleton loading states during API calls
- Integration with external MCP server for real estate data
