# Summary comment

One issue comment, short enough to read in 20 seconds. Fill the template and drop any section that would be empty.

```markdown
<!-- pr-review sha=<headRefOid> base=<merge base> -->
**Must review (2)**

- [ ] `src/api/auth.ts` (1): expired tokens pass, the expiry check is inverted
- [ ] `db/migrations/0012_orders_index.sql` (2): index build locks writes on `orders`
  - Question, L3: how large is `orders` in production?

**Worth a glance (1)**

- `src/cart/total.ts` (3): new discount rounding, no test for negative totals

<details>
<summary>12 files you can skip</summary>

| Grade | File | Why |
|:-:|---|---|
| 4 | `package.json` | Bumps `zod` from 3.23.8 to 3.24.1 |
| 5 | `bun.lock` | Lockfile |

</details>

**Nits**: `src/cart/total.ts:12` uses `.forEach`, the project uses `for…of`.

<sub>Grade 1 to 5: how sure the review is that you can skip the file. Reviewed at `abc1234`.</sub>
```

- Every changed file appears once, lowest grade first, then by path.
- Grades 1 and 2 go on the checklist. Grade 3 goes in the plain list. Grades 4 and 5 go in the collapsed table.
- Files left out of a partial review go last on the checklist, with `-` as the grade and "not reviewed" as the reason.
- A finding that cannot go inline goes as a sub-item under its file, one line. When the user posts the summary only, every finding goes there.
- At most 3 nits, on one line.
- On a re-run, a box the previous summary had ticked stays ticked when the file is unchanged. A changed file gets an empty box.
- The first line is a hidden marker with the reviewed SHA and the merge base. Re-runs find the comment by it, so keep it exact. "(unchanged)" after a reason means the grade comes from the previous run.
- A path is written by the PR author. Escape `|` as `\|`, and wrap a path that contains a backtick in double backticks with inner spaces.
- GitHub rejects a comment over 65,536 characters. Past about 60,000, drop the nits, then group the collapsed table by directory with a count. A grouped file has no row to carry over, so the next run triages it again. If it still does not fit, say so in Step 6 and ask what to leave out.

## Inline comments

Two kinds, both on lines of the new file inside a diff hunk:

- **Finding**, for a `bug`, `minor` or `question`. Start with the severity: "Bug:", "Minor:" or "Question:". Say what goes wrong and what triggers it, then the fix. Add a suggestion block when the fix is a few lines.
- **Note**, on a grade 1 to 3 file: the lines where a mistake is most likely and the concrete risk there, what would break and when. A note that only says "check that X matches Y" names no risk: leave it out. At most 2 per file, and none where a finding already covers the lines.

Each comment is one or two sentences, plus the suggestion block. No comment for nits, nor for a line outside a hunk.

````markdown
Bug: a token that expired one second ago passes, because the comparison is inverted.

```suggestion
  if (token.expiresAt < now) {
```
````

A suggestion replaces every line from `start_line` to `line`, so it must contain all of them.

## Writing rules

The `humanizer` skill does most of this. Apply these rules either way:

- Put the fact first. Name the function, the value, the input.
- No opener ("I noticed", "It looks like"), no closing line, no praise.
- No "not X but Y" contrasts, no lists of three for rhythm, no dashes as punctuation, no bold beyond the template's labels.
- One hedge at most. When unsure, write a question.
