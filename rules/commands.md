---
description: Development commands for the monorepo. Use when running builds, dev servers, linting, formatting, type-checking, or database operations.
alwaysApply: true
---

# Development Commands

## Essential Commands (run from monorepo root)
- `bun dev` - Start all apps via Turborepo (landing :3000, app :3001)
- `bun run build` - Build all apps
- `bun run check-types` - Type-check all packages
- `bun run lint` - Run Biome linter with auto-fix
- `bun run format` - Format code with Biome
- `bun run check` - Run Biome check with auto-fix

## Filtering by App
- `bunx turbo run dev --filter=@project/app` - Start app only
- `bunx turbo run dev --filter=@project/landing` - Start landing only
- `bunx turbo run build --filter=@project/app` - Build app only

## Database Commands
- `bun run db:generate` - Generate Drizzle migrations
- `bun run db:migrate` - Run Drizzle migrations
- `bun run db:studio` - Open Drizzle Studio

## Package Manager
This project uses **Bun** as the package manager and runtime. Use `bun install` from the root for all dependencies.
