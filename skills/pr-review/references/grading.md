# Confidence grades

A grade says how confident the review is that a human can skip the file. It is not a quality score: a clean change to authentication still gets a low grade, because a mistake there is expensive.

| Grade | Meaning |
|:-:|---|
| 5 | Given by triage only: Skip list, formatting only, or a pure rename |
| 4 | Small, isolated change, every case it handles checked |
| 3 | Logic change with nothing wrong found, or only `minor` findings, but not every case could be checked |
| 2 | Always review list, sensitive area, blind spot, or an unconfirmed issue |
| 1 | Confirmed bug |

## How to grade

A file you read starts at 4 and gets the lowest cap below that applies. Nothing raises a grade above a cap: not clean code, not a trusted author, not a passing CI.

Grade 4 files are marked viewed and nobody reads them. Give 4 only when you can list every case the change handles and you checked each one. When in doubt, give 3.

### Cap at 3: logic you cannot fully check by reading

- Too large or too varied to list every case it handles.
- New or changed branching, loops, state, error handling, async flow or data transformation, beyond a few lines.
- Behavior that depends on something outside the repo: an external API, an environment variable, a browser or runtime quirk, database content.
- Changed logic that no test exercises.
- An internal signature or type changed, and you checked every caller.
- A dependency bumped across a major version.
- A change the PR title and description do not explain.
- A confirmed `minor` finding.

### Cap at 2: an expensive mistake, or a blind spot

- On the Always review list of `file-lists.md`.
- The changed lines take part in authentication, authorization, sessions, payments, secrets, cryptography, user input reaching SQL, a shell, HTML or a file path, data deletes, concurrency or caching, even when the path does not say so.
- A contract other systems rely on: an HTTP API, a published package's exports, a database schema, an event format.
- Generated or built code that runs as committed, like codegen output the app imports or a vendored `*.min.js`.
- Something that matters could not be checked: a caller you could not find, a rewrite too large to follow.
- Deleted code whose remaining references you have not ruled out with a search.
- A plausible issue you could neither confirm nor refute.

### Cap at 1: a confirmed problem

- A concrete input or state leads to a wrong result, a crash, a security hole or data loss.
- A caller, a test or a type left out of date, so that the build or the tests fail. A stale test that still passes is a `minor`.
- A secrets file from the Always review list.

When two grades fit, pick the lower one. Grading a fine file too low costs a few minutes. Grading a buggy file too high means nobody reads it.

## Reason

Every grade comes with one reason, 12 words at most, specific to the file: "Inverts the expiry check in `verifyToken`", not "logic change". For grades 4 and 5, say why it is safe: "rename only", "copy change", "covered by `cart.test.ts`".

## Findings

A finding has a path, a line on the new side of the diff (and a start line when it spans several), what goes wrong, and the input or state that triggers it. If you cannot write the trigger, it is a `question`.

| Severity | When | Effect on the grade |
|---|---|---|
| `bug` | Confirmed with a trigger: a wrong result, a crash, a security hole or data loss | Grade 1 |
| `question` | Plausible issue, could not confirm or refute | At most 2, or 3 when it would be a `minor` |
| `minor` | Confirmed, but nothing breaks for users or data: a misleading message, a cosmetic glitch, a check that covers fewer cases than it should | At most 3 |
| `nit` | Style from the project's `CLAUDE.md`, `AGENTS.md`, `CODING_STYLE.md`, `.claude/rules/` or lint config | None |
