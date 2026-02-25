---
description: Monorepo architecture, core stack, directory structure, and import conventions. Use when creating new files, adding imports, or understanding where code belongs in the project.
alwaysApply: true
---

# Architecture Overview

This project is organized as a **Turborepo monorepo** with Bun workspaces.

## Core Stack
- **Framework**: Next.js 16 with App Router
- **Monorepo**: Turborepo with Bun workspaces
- **Runtime**: Bun
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4 + shadcn/ui components
- **Linting/Formatting**: Biome (shared config at root `biome.json`)

## Monorepo Structure

```
my-project/
├── apps/
│   ├── project-a/                # @project/app — Main application
│   │   └── src/
│   │       ├── app/        # Next.js App Router (authenticated, auth, API routes, SEO pages)
│   │       ├── components/ # App-specific components
│   │       ├── contexts/   # React contexts
│   │       ├── db/         # Drizzle ORM (schema, queries, migrations, seed)
│   │       ├── helpers/    # App-specific helpers
│   │       ├── hooks/      # App-specific hooks
│   │       ├── lib/        # App-specific libs
│   │       ├── models/     # TypeScript models
│   │       ├── providers/  # React providers
│   │       ├── schemas/    # Zod schemas
│   │       ├── services/   # App services
│   │       ├── templates/  # Email templates
│   │       ├── types/      # TypeScript types
│   │       ├── utils/      # App utilities
│   │       └── env.ts      # App env validation (@t3-oss/env-nextjs)
│   └── project-b/            # @project/landing — Public landing page
│       └── src/
│           ├── app/        # Next.js App Router (home, blog, legal, sitemap, robots)
│           ├── components/ # Landing-specific components
│           ├── content/    # MDX blog content
│           ├── data/       # SEO data
│           ├── hooks/      # Landing hooks
│           ├── lib/        # Landing libs
│           └── env.ts      # Landing env validation (minimal)
├── packages/
│   ├── common/             # @project/common — Shared code used by both apps
│   │   └── src/
│   │       ├── components/ # Shared components
│   │       │   ├── analytics/  # Analytics, CookieConsent
│   │       │   ├── layout/     # Header, Footer
│   │       │   ├── seo/        # Structured data
│   │       │   └── ui/
│   │       │       ├── external/   # shadcn/ui components
│   │       │       └── internal/   # Custom shared UI
│   │       ├── data/       # Shared data (navigation)
│   │       ├── helpers/    # Shared helpers (date, number, validation)
│   │       ├── hooks/      # Shared hooks (useDebounce, useMobile, useAnalytics)
│   │       ├── lib/        # Shared libs (utils — cn())
│   │       ├── services/   # Shared services
│   │       └── styles/     # Shared base CSS (globals.css — keyframes, CSS vars, theme)
│   └── typescript-config/  # @project/typescript-config — Shared TS config
├── turbo.json              # Turborepo task configuration
├── biome.json              # Shared Biome config (root-level, applies to all packages)
├── tsconfig.base.json      # Base TypeScript config extended by all packages
└── package.json            # Root workspace config (workspaces, turbo scripts)
```

## Import Conventions

### Within an app (`apps/project-a/` or `apps/project-b/`)
- **Internal imports**: Use `@/src/...` path alias (e.g., `import { env } from '@/src/env'`)
- **Shared imports**: Use `@project/common/...` package imports (e.g., `import { cn } from '@project/common/lib/utils'`)

### Within `packages/common/`
- **Internal imports**: Use relative paths only (e.g., `import { cn } from '../../lib/utils'`)
- No `@/src/` alias inside common — all imports must be relative

### Shared vs App-specific
- Code used by **both apps** belongs in `packages/common/`
- Code used by **only one app** stays in that app's `src/`
- Components that depend on `env.ts` must accept config as **props/parameters** to remain shareable (Analytics, Crisp, Footer accept props instead of importing env directly)
