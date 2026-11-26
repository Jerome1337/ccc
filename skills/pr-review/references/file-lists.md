# File lists

Triage matches each changed file against these lists by path, before anyone opens it. Always review wins over Skip: a file on both is reviewed.

Matching ignores case. A pattern without `/` matches the file name at any depth, and `name/**` matches a folder of that name at any depth: `.env` matches `apps/web/.env`, `dist/**` matches `packages/ui/dist/index.js`. Any other pattern with `/` matches from the repo root. `*` stays inside one folder, `**` crosses folders.

## Always review

The file goes on the must-review checklist with at most grade 2, whatever its diff looks like.

- **Secrets, grade 1, never opened**: `.env` and `.env.*` except `.env.example`, `.env.sample`, `.env.template` and `.env.dist`; `*.pem`, `*.key`, `*.p12`, `id_rsa*` and `id_ed25519*` except `*.pub`; `credentials*.json`, `service-account*.json`. Reason: "Secrets file committed". Do not read, print or quote its content.
- **Built output, grade 2, not read**: `dist/**`, `build/**`, `*.generated.*`, `*.min.js`. It may run as committed, like the `dist/index.js` of a GitHub Action, and reading it proves nothing. Reason: what builds it, so the human can rebuild and compare.
- **Runs on install or on a developer's machine**: `package.json` when `scripts` change or a dependency is added; `.npmrc`, `.yarnrc*`, `bunfig.toml`; `.husky/**`, `lefthook.y*ml`, `.pre-commit-config.yaml`; `.claude/settings*.json`, `.mcp.json`, `.cursor/mcp.json`; `.vscode/tasks.json`, which can run a command when the folder opens; `.devcontainer/**`.
- **CI, deploy, infrastructure**: `.github/workflows/**`, `.github/actions/**`, `action.y*ml`, `Dockerfile*`, `docker-compose*`, `compose.y*ml`, `*.tf`, `k8s/**`, `helm/**`, `vercel.json`, `netlify.toml`, `fly.toml`, `wrangler.toml`, `Procfile`, `.goreleaser.y*ml`, `CODEOWNERS`.
- **Database**: `migrations/**`, `*.sql`, `prisma/schema.prisma`, `drizzle/**`, `seeds/**`.
- **Auth, permissions, payments**: a code or config file, not an asset or a doc, whose path contains `auth`, `login`, `session`, `permission`, `role`, `payment`, `billing`, `checkout`, `stripe` or `webhook`, and `middleware.*`. A path where `auth` only appears inside the word `author` or `authors` does not count. `authorize` does.
- **Public contracts**: `openapi.*`, `swagger.*`, `*.proto`, `*.graphql`.
- **A lockfile that adds a new source**: a URL on a host the file did not use before, or a new `git+`, `github:` or `file:` source. Compare `git --literal-pathspecs diff <base>...HEAD -- '<path>' | grep '^+'` with `git show <base>:'<path>'`. This is the only diff triage reads for a Skip-list file. `bun.lockb` is binary: skip the check.

## Skip

Grade 5, marked viewed, never opened.

- **Lockfiles**: `bun.lock`, `bun.lockb`, `package-lock.json`, `npm-shrinkwrap.json`, `pnpm-lock.yaml`, `yarn.lock`, `go.sum`, `Cargo.lock`, `composer.lock`, `Gemfile.lock`, `poetry.lock`, `uv.lock`.
- **Type, lint and format config**: `tsconfig*.json`, `jsconfig*.json`, `.eslintrc*`, `eslint.config.*`, `.eslintignore`, `.prettierrc*`, `prettier.config.*`, `.prettierignore`, `biome.json*`, `.stylelintrc*`, `stylelint.config.*`, `.editorconfig`, `.markdownlint*`, `commitlint.config.*`, `.golangci.y*ml`, `rustfmt.toml`, `ruff.toml`.
- **Tool versions, ignore files, editor settings**: `.nvmrc`, `.node-version`, `.tool-versions`, `.python-version`, `.gitignore`, `.gitattributes`, `.dockerignore`, `.vscode/**`, `.idea/**`.
- **Instructions for AI agents**: `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.claude/**/*.md` (skills, rules, commands, agents), `.cursorrules`, `.cursor/rules/**`, `.windsurfrules`, `.github/copilot-instructions.md`, `*.prompt.md`, `llms.txt`.
- **Docs**: `README*`, `CHANGELOG*`, `CONTRIBUTING*`, `docs/**/*.md`. Not `.mdx`, which can run code.
- **Assets**: `*.png`, `*.jpg`, `*.jpeg`, `*.gif`, `*.webp`, `*.avif`, `*.ico`, `*.svg`, `*.woff`, `*.woff2`, `*.ttf`, `*.otf`, `*.mp4`, `*.webm`, `*.mp3`.
- **Test output**: `coverage/**`, `__snapshots__/**`, `*.snap`.
