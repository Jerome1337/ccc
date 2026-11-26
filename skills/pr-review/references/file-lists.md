# File lists

Triage matches each changed file against these lists by path, before anyone opens it. Always review wins over Skip: a file on both is reviewed.

## Always review

The file goes on the must-review checklist with at most grade 2, whatever its diff looks like.

- **Secrets, grade 1, never opened**: `.env` and `.env.*` except `.env.example`, `*.pem`, `*.key`, `*.p12`, `id_rsa*`, `id_ed25519*`, `credentials*.json`, `service-account*.json`. Reason: "Secrets file committed". Do not read, print or quote its content.
- **Runs on install or on a developer's machine**: `package.json` when `scripts` change or a dependency is added; `.npmrc`, `.yarnrc*`, `bunfig.toml`; `.husky/**`, `lefthook.y*ml`, `.pre-commit-config.yaml`; `.claude/settings*.json`, `.mcp.json`, `.cursor/mcp.json`; `.devcontainer/**`.
- **CI, deploy, infrastructure**: `.github/workflows/**`, `.github/actions/**`, `action.y*ml`, `Dockerfile*`, `docker-compose*`, `compose.y*ml`, `*.tf`, `k8s/**`, `helm/**`, `vercel.json`, `netlify.toml`, `fly.toml`, `wrangler.toml`, `Procfile`, `.goreleaser.y*ml`, `CODEOWNERS`.
- **Database**: `migrations/**`, `*.sql`, `prisma/schema.prisma`, `drizzle/**`, `seeds/**`.
- **Auth, permissions, payments**: a path containing, in any case, `auth` (not `author`), `login`, `session`, `permission`, `role`, `payment`, `billing`, `checkout`, `stripe` or `webhook`, and `middleware.*`.
- **Public contracts**: `openapi.*`, `swagger.*`, `*.proto`, `*.graphql`, `schema.graphql`.
- **A lockfile whose diff adds a URL on a host the file did not use before**: `git --literal-pathspecs diff <base>...HEAD -- '<path>' | grep '^+' | grep -oE 'https?://[^/"]+' | sort -u`, compared with the hosts in the base version.

## Skip

Grade 5, marked viewed, never opened.

- **Lockfiles**: `bun.lock`, `bun.lockb`, `package-lock.json`, `npm-shrinkwrap.json`, `pnpm-lock.yaml`, `yarn.lock`, `go.sum`, `Cargo.lock`, `composer.lock`, `Gemfile.lock`, `poetry.lock`, `uv.lock`.
- **Type, lint and format config**: `tsconfig*.json`, `jsconfig*.json`, `.eslintrc*`, `eslint.config.*`, `.eslintignore`, `.prettierrc*`, `prettier.config.*`, `.prettierignore`, `biome.json*`, `.stylelintrc*`, `stylelint.config.*`, `.editorconfig`, `.markdownlint*`, `commitlint.config.*`, `.golangci.y*ml`, `rustfmt.toml`, `ruff.toml`.
- **Tool versions, ignore files, editor settings**: `.nvmrc`, `.node-version`, `.tool-versions`, `.python-version`, `.gitignore`, `.gitattributes`, `.dockerignore`, `.vscode/**`, `.idea/**`.
- **Instructions for AI agents**: `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.claude/**/*.md` (skills, rules, commands, agents), `.cursorrules`, `.cursor/rules/**`, `.windsurfrules`, `.github/copilot-instructions.md`, `*.prompt.md`, `llms.txt`.
- **Docs**: `README*`, `CHANGELOG*`, `CONTRIBUTING*`, `docs/**/*.md`. Not `.mdx`, which can run code.
- **Assets**: `*.png`, `*.jpg`, `*.jpeg`, `*.gif`, `*.webp`, `*.avif`, `*.ico`, `*.svg`, `*.woff`, `*.woff2`, `*.ttf`, `*.otf`, `*.mp4`, `*.webm`, `*.mp3`.
- **Generated output**: `dist/**`, `build/**`, `coverage/**`, `__snapshots__/**`, `*.snap`, `*.generated.*`.
