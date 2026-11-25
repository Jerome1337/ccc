# Summary comment

One issue comment on the PR. Fill the template below and drop any section that would be empty.

- Every changed file gets a row, lowest grade first, then by path.
- Files graded 1 to 3 go in the main table. Files graded 4 and 5 go in the collapsed table.
- Files left out of a partial review go last in the main table, with `-` as the grade and "Not reviewed" as the reason. Add "N not reviewed" to the opening line.
- The first line is a hidden marker with the reviewed SHA. Re-runs find the comment by it, so keep it exact.
- "(unchanged)" after a reason means the grade, the findings and the nits of that file come from the previous run, because the file has not changed since then.

```markdown
<!-- pr-review sha=<headRefOid> -->
## Review: where to look

Read 2 files, spot-check 1, skim or skip 3.

| Grade | File | Why | Look at |
|:-:|---|---|---|
| 1 | `src/api/auth.ts` | Expired tokens pass: the expiry check is inverted | `verifyToken`, L42-58 |
| 2 | `db/migrations/0012_orders_index.sql` | Builds an index without `CONCURRENTLY` on the orders table | whole file |
| 3 | `src/cart/total.ts` | New discount rounding, no test covers negative totals | `applyDiscount` |

<details>
<summary>3 files graded 4 or 5</summary>

| Grade | File | Why |
|:-:|---|---|
| 4 | `package.json` | Bumps `zod` from 3.23.8 to 3.24.1 |
| 4 | `src/components/Header.tsx` | Copy change only |
| 5 | `bun.lock` | Lockfile, every entry follows the `zod` bump |

</details>

### Findings

1. **Bug** `src/api/auth.ts:47`: a token that expired one second ago is accepted. Trigger: any request with an expired token. Fix: compare with `<`, not `>`.
2. **Question** `db/migrations/0012_orders_index.sql:3`: the index build locks writes on `orders`. How large is the table in production?

### Nits

- `src/cart/total.ts:12`: `.forEach`, where the project uses `for…of`.

<sub>5 skip, 4 skim, 3 spot-check, 2 read, 1 read carefully. A grade is how confident the review is that you can skip the file, not a quality score. Reviewed at `abc1234`.</sub>
```

The opening line counts the files a human has to act on: "Read" covers grades 1 and 2, "spot-check" grade 3, "skim or skip" grades 4 and 5.

## Inline comments

One per finding of severity `bug`, `minor` or `question`, never for nits. A finding outside a diff hunk, or on a file the PR does not change, stays out of the inline comments: Step 7 of `SKILL.md` says where it goes. Each body says what goes wrong, what triggers it, and the fix, in three short sentences at most. When the fix is a few changed lines, add a suggestion block:

````markdown
A token that expired one second ago is accepted, because the comparison is inverted.

```suggestion
  if (token.expiresAt < now) {
```
````

A suggestion replaces every line from `start_line` to `line`, so it must contain all of them, changed or not. On a single-line comment it replaces that one line.
