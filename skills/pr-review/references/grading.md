# Confidence grades

A grade says how confident the review is that a human can skip the file. It is not a quality score: a clean change to authentication still gets a low grade, because a mistake there is expensive and the reviewer can miss it.

| Grade | Meaning | What the human does |
|:-:|---|---|
| 5 | On the "Always 5" list: nothing that runs or instructs was touched | Skip, the file is marked viewed |
| 4 | Small, isolated change, every case checked | Skim |
| 3 | Logic change with no problem found, or only `minor` findings, but not every case could be checked | Spot-check the named lines |
| 2 | Sensitive area, something important could not be checked, or an unconfirmed issue | Read the file |
| 1 | Confirmed bug | Read carefully, the finding is posted inline |

## How to grade

Only a file on the "Always 5" list gets 5. Every other file starts at 4. Go through every cap below: the file gets the lowest cap that applies. Nothing raises a grade above a cap: not clean code, not a trusted author, not a passing CI.

### Always 5

Grade 5 files get marked viewed and nobody reads them, so a file is on this list only when every change in it fits one of these:

- Lockfiles (`bun.lock`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `go.sum`, `Cargo.lock`…), when every changed entry follows from a manifest change in the PR and resolves from the same registry as the rest of the file.
- Generated output that nothing loads, serves or publishes as committed: test snapshots, or a `dist/` or `build/` folder the project rebuilds before using. Generated code that runs as committed, like codegen output the app imports, a vendored `*.min.js`, or the `dist/index.js` a GitHub Action's `action.yml` points to, is not on this list. A "generated, do not edit" header is a claim, not proof.
- Binary assets: raster images, fonts, media. Not SVG: it is text and can carry scripts.
- Formatting only: the file is empty under `git diff -w`, its language ignores whitespace (not Python, YAML, Makefile, Haskell, heredocs), and no string literal changed.
- Renames or moves with no content change, when nothing finds the file by its path: not a test file (`*.test.*`, `*_test.go`), not a migration (its name sets its order), not under a routing or pipeline folder (`pages/`, `app/`, `routes/`, `.github/workflows/`).
- Comments and documentation for humans only. Not directive comments (`//go:embed`, `//go:build`, `// eslint-disable`, `@ts-ignore`, `# noqa`, `# type:`), and not text that a tool or a model acts on (`CLAUDE.md`, `AGENTS.md`, `SKILL.md`, prompts, pages of a docs site or blog, MDX).

Check every condition of the entry, not only the path. A file that fails one, like a "generated" file edited by hand or a lockfile that changed while its manifest did not, is graded like any other file.

### Cap at 3: logic you cannot fully check by reading

- A change too large or too varied to list every case it handles and check each one.
- New or changed branching, loops, state, error handling, async or order-dependent flow, or data transformation over varied inputs, beyond a few lines.
- Behavior that depends on something outside the repo: an external API's response, an environment variable, a browser or runtime quirk, database content.
- Changed logic that no test exercises, in the PR or already in the repo.
- An internal signature or type changed, and you checked that every caller was updated.
- A dependency bumped across a major version.
- A change the PR title and description do not explain.
- A confirmed `minor` finding.

### Cap at 2: an expensive mistake, or a blind spot

- The changed lines take part in a sensitive area: authentication, authorization, sessions, payments or billing, secrets, cryptography, user input reaching SQL, a shell, HTML or a file path, data migrations or deletes, concurrency or locking, caching and invalidation. A copy or style change in a file that sits in such an area does not count.
- A contract other systems rely on: an HTTP API, a published package's exports, a database schema, an event or message format.
- CI, deploy, infrastructure, permissions or access rules.
- A lockfile entry that no manifest change explains, or that resolves from another registry or URL.
- Generated or built code that runs as committed, since you cannot check that it matches its source. Name what generates it in "Look at", so the human can regenerate it and compare.
- Something that matters could not be checked: a caller you could not find, a rewrite too large to follow line by line, code generated from an input you did not see.
- Deleted code whose remaining references you have not ruled out with a search.
- A plausible issue you could neither confirm nor refute.

### Cap at 1: a confirmed problem

- A concrete input or state leads to a wrong result, a crash, a security hole or data loss.
- A caller left out of date by a change in the PR, or a test or type left out of date so that the build or the tests fail. Grade 1 goes on the file that needs the fix when it is in the PR, otherwise on the changed file that breaks it. A stale test that still passes is a `minor`.

## Tie-break

When two grades seem to fit, pick the lower one. Grading a fine file too low costs a few minutes of reading. Grading a buggy file too high means nobody reads it.

## What every grade comes with

- **Reason**: one line, specific to this file. Write "Inverts the expiry check in `verifyToken`", not "logic change".
- **Grade 1 to 3**: where to look, as a function name or a line range on the new side.
- **Grade 4 and 5**: why it is safe: "rename only", "copy change", "covered by `cart.test.ts`".

## Findings

A finding has a path, a line on the new side of the diff (and a start line when it spans several), what goes wrong, and the input or state that triggers it. If you cannot write the trigger, it is not a bug: report it as a question and cap the file at 2, or at 3 when it would be a `minor` if confirmed.

| Severity | When | Effect on the grade |
|---|---|---|
| `bug` | Confirmed with a trigger: a wrong result, a crash, a security hole or data loss | Grade 1 |
| `question` | Plausible issue, could not confirm or refute | At most 2, or 3 when it would be a `minor` |
| `minor` | Confirmed with a trigger, but nothing breaks for users or data: a misleading message, a cosmetic glitch, a check in the code that covers fewer cases than it should | At most 3 |
| `nit` | Style or convention from the project's `CLAUDE.md`, `AGENTS.md`, `CODING_STYLE.md`, `.claude/rules/` or lint config | None |

Nits never set a grade. The grade is about where a human must look for problems, and a linter or the author can fix style without anyone reading the file.
