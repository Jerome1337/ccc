# File lists

Triage matches each changed file against these lists by path, before anyone opens it. Always review wins over Skip: a file on both is reviewed.

Matching ignores case, and every pattern matches at any depth, so a monorepo package is treated like the root: `.env` matches `apps/web/.env`, `prisma/schema.prisma` matches `packages/db/prisma/schema.prisma`, `.vscode/tasks.json` matches `apps/web/.vscode/tasks.json`. `*` stays inside one folder, `**` crosses folders.

## Always review

The file goes on the must-review checklist with at most grade 2, whatever its diff looks like.

- **Secrets, grade 1, never opened**: `.env` and `.env.*`, except a name ending in `.example`, `.sample`, `.template` or `.dist` (`.env.local.example` is not a secret); `*.pem`, `*.key`, `*.p12`, `*.pfx`, `*.jks`, `*.keystore`, `id_rsa*` and `id_ed25519*` except `*.pub`; `.netrc`, `.pgpass`, `credentials*.json`, `service-account*.json`. Reason: "Secrets file committed". Do not read, print or quote its content.
- **Built output, grade 2, not read**: `dist/**`, `build/**`, `*.generated.*`, `*.min.js`. It may run as committed, like the `dist/index.js` of a GitHub Action, and reading it proves nothing. Reason: what builds it, so the human can rebuild and compare.
- **Runs on install or on a developer's machine**: `package.json` when `scripts` change or a dependency is added; `.npmrc`, `.yarnrc*`, `bunfig.toml`; `.husky/**`, `lefthook.y*ml`, `.pre-commit-config.yaml`; `.claude/settings*.json`, `.mcp.json`, `.cursor/mcp.json`; `.vscode/tasks.json`, which can run a command when the folder opens; `.devcontainer/**`.
- **CI, deploy, infrastructure**: `.github/workflows/**`, `.github/actions/**`, `action.y*ml`, `Dockerfile*`, `docker-compose*`, `compose.y*ml`, `*.tf`, `k8s/**`, `helm/**`, `vercel.json`, `netlify.toml`, `fly.toml`, `wrangler.toml`, `Procfile`, `.goreleaser.y*ml`, `CODEOWNERS`.
- **Database**: `migrations/**`, `*.sql`, `prisma/schema.prisma`, `drizzle/**`, `seeds/**`.
- **Auth, permissions, payments**: a code or config file, not an asset or a doc, whose path contains `auth`, `login`, `session`, `permission`, `role`, `payment`, `billing`, `checkout`, `stripe` or `webhook`, and `middleware.*`. A path where `auth` only appears inside the word `author` or `authors` does not count. `authorize` does.
- **Public contracts**: `openapi.*`, `swagger.*`, `*.proto`, `*.graphql`.
- **A lockfile that adds a new source**: a host or a `git+`, `github:` or `file:` source the base version did not use. It is the only guard on a file nobody opens, so list the sources of both versions and compare them, instead of reading the diff by eye. Any output sends the file here. A new lockfile has no base version, so all its sources count as new. `bun.lockb` is binary: review it like any other file.

  ```bash
  sources() { grep -oE '[a-z+]+://[^/" ]+|(github|gitlab|bitbucket|file|link):' | sort -u; }
  comm -13 <(git show <merge base>:'<path>' 2>/dev/null | sources) <(git show HEAD:'<path>' | sources)
  ```

- **Agent files that run commands**: a `.claude/**/*.md` file whose diff adds a `` !`command` `` line, which runs in a shell when the command loads, or changes `allowed-tools`. Check with `git --literal-pathspecs diff <base>...HEAD -- '<path>' | grep -E '^\+.*(!`|allowed-tools)'`.

These two checks are the only diffs triage reads for a file on the Skip list.

## Skip

Grade 5, marked viewed, never opened.

- **Lockfiles**: `bun.lock`, `package-lock.json`, `npm-shrinkwrap.json`, `pnpm-lock.yaml`, `yarn.lock`, `go.sum`, `Cargo.lock`, `composer.lock`, `Gemfile.lock`, `poetry.lock`, `uv.lock`.
- **Type, lint and format config, in data formats only**: `tsconfig*.json`, `jsconfig*.json`, `.eslintrc`, `.eslintrc.json`, `.eslintrc.y*ml`, `.eslintignore`, `.prettierrc`, `.prettierrc.json`, `.prettierrc.y*ml`, `.prettierrc.toml`, `.prettierignore`, `biome.json*`, `.stylelintrc`, `.stylelintrc.json`, `.stylelintrc.y*ml`, `.editorconfig`, `.markdownlint.json`, `.markdownlint.y*ml`, `.commitlintrc.json`, `.commitlintrc.y*ml`, `.golangci.y*ml`, `rustfmt.toml`, `ruff.toml`. Their `.js`, `.cjs`, `.mjs` and `.ts` forms, like `eslint.config.js`, run code when the tool loads them: they are reviewed like any other file.
- **Tool versions, ignore files, editor settings**: `.nvmrc`, `.node-version`, `.tool-versions`, `.python-version`, `.gitignore`, `.gitattributes`, `.dockerignore`, `.vscode/**`, `.idea/**`.
- **Instructions for AI agents**: `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.claude/**/*.md` (skills, rules, commands, agents), `.cursorrules`, `.cursor/rules/**`, `.windsurfrules`, `.github/copilot-instructions.md`, `*.prompt.md`, `llms.txt`.
- **Docs**: `README*`, `CHANGELOG*`, `CONTRIBUTING*`, `docs/**/*.md`. Not `.mdx`, which can run code.
- **Assets**: `*.png`, `*.jpg`, `*.jpeg`, `*.gif`, `*.webp`, `*.avif`, `*.ico`, `*.svg`, `*.woff`, `*.woff2`, `*.ttf`, `*.otf`, `*.mp4`, `*.webm`, `*.mp3`.
- **Test output**: `coverage/**`, `__snapshots__/**`, `*.snap`.
