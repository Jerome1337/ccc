---
name: pr-review
description: Reviews a GitHub pull request and grades every changed file from 1 to 5 by how confident the review is that a human can skip it, so the human reads the risky files instead of the whole PR. After the user confirms, posts the graded file list as a PR comment, concrete findings as inline comments, and marks the files that need no review as viewed. Use whenever the user wants to review a pull request, triage a large PR, know which files of a PR to look at, or comment on a PR. Triggers on "review PR", "review pull request", "review PR #N", "which files should I look at", "code review PR", "comment on PR", even when the user only gives a PR number.
---

# PR Review

Review a pull request and give every changed file a confidence grade from 1 to 5. The higher the grade, the more confident the review is that a human can skip the file. The human reads the low grades and skims the rest.

Grades follow `references/grading.md`. Read it before grading anything. The posted comments follow `references/summary.md`.

## Step 1: Load the PR

You need a PR number or URL. Ask for it if the user did not give one. If `gh auth status` fails, tell the user to run `gh auth login`.

```bash
gh pr view <N> --json id,number,title,body,author,state,isDraft,baseRefName,headRefName,headRefOid,changedFiles,url
gh api user --jq .login
```

- Add `--repo <owner>/<repo>` when the PR is not in the current repository. Take `<owner>/<repo>` from the PR's `url`.
- If the PR is merged or closed, say so and ask before going further.
- Read the title, the description and any linked issue. Files are graded against what the PR says it does.

Read what is already there, so you do not repeat it:

```bash
gh api repos/<owner>/<repo>/issues/<N>/comments --paginate
gh api repos/<owner>/<repo>/pulls/<N>/comments --paginate
gh api repos/<owner>/<repo>/pulls/<N>/reviews --paginate
```

The previous summary, if any, is the most recent issue comment written by the user's own login that starts with `<!-- pr-review sha=… -->`. Note its id and SHA. Ignore marker comments from other people.

## Step 2: Check out the PR head

Reviewing needs whole files and the rest of the repo to find callers, tests and types, not only the diff. Put the head in a temporary directory so the user's working tree is not touched. Use the session scratchpad if there is one.

When the current directory is a clone of the PR's repository, `<remote>` is the remote whose URL matches `<owner>/<repo>` in `git remote -v`. It is not always `origin`: in a fork, `origin` is the fork.

```bash
git fetch <remote> <baseRefName> "pull/<N>/head"
git worktree add --detach <tmp>/pr-<N> <headRefOid>
```

Otherwise clone it, and `<remote>` is `origin`:

```bash
gh repo clone <owner>/<repo> <tmp>/pr-<N> -- --filter=blob:none
git -C <tmp>/pr-<N> fetch origin <baseRefName> "pull/<N>/head"
git -C <tmp>/pr-<N> checkout --detach <headRefOid>
```

From inside it, with `<base>` as `<remote>/<baseRefName>`:

```bash
git diff --name-status -M <base>...HEAD          # the file list, renames as old and new path
git diff --stat -M <base>...HEAD                 # sizes
git diff -w --stat <base>...HEAD                 # formatting-only files drop out
git diff -M <base>...HEAD -- <path>              # one file
git diff -M <base>...HEAD -- <old path> <path>   # one renamed file: pass both paths
```

Take the file list from git, not from `gh pr view`, which stops at 100 files. If its count differs from `changedFiles`, tell the user and go on with git's list.

When you are done, even when the review stops early, remove it: `git worktree remove --force <tmp>/pr-<N>` for a worktree, `rm -rf <tmp>/pr-<N>` for a clone.

## Step 3: Triage

When a previous summary exists and `git merge-base --is-ancestor <previous sha> HEAD` succeeds, the files missing from `git diff --name-only <previous sha> HEAD` keep their grade, reason, findings and nits from the previous summary, marked "(unchanged)", unless a file they depend on changed. When the command fails, after a force-push for example, review every file.

Sort the other files into two lists:

- **Grade 5 without a full review**: the "Always 5" list in `references/grading.md`. Open each one to confirm it is what its path says.
- **To review**: everything else.

Pick the review mode from the files to review:

| Files to review | Mode |
|---|---|
| Up to 10 files and up to 500 changed lines in them | Single pass, Step 4a |
| Up to 60 files | Split, Step 4b |
| Over 60 files | Tell the user the size and how many subagents it takes, then ask whether to review everything or only some directories |

## Step 4a: Single pass

For each file to review: read the diff and the whole new file, search the worktree for callers of anything whose signature or behavior changed, check whether a test exercises the change, then grade it. Record the grade, the reason, where to look and the findings.

## Step 4b: Split across subagents

Group the files to review, keeping related files together: same feature or directory, a source file with its test, a type with its users. Close a group at 8 files or about 400 changed lines, whichever comes first, so one large file can be a group of its own. Launch one `general-purpose` subagent per group, at most 8 at a time in one message so they run in parallel. With more groups, run them in waves of 8.

```text
You are reviewing part of PR #<N> in <owner>/<repo>: "<title>".
What the PR says it does: <description, trimmed>

The PR head is checked out at <worktree>. Diff a file with:
git -C <worktree> diff -M <base>...HEAD -- <path>
For a renamed file, pass the old path too: -- <old path> <path>

Your files: <list, with the old path of renamed files>
Other files in this PR, for context only: <list>

Read <this skill's directory>/references/grading.md and grade each of your files by it.
Read whole files, not only the diff. Search the whole worktree for callers of anything
whose behavior or signature changed, including files that are not yours.
Do not edit files, do not commit, do not post anything to GitHub.

Return only a JSON array, one object per file:
[{"path": "", "grade": 1, "reason": "", "lookAt": "",
  "findings": [{"startLine": 0, "line": 0, "severity": "bug|question|nit",
                "summary": "", "trigger": "", "fix": ""}]}]
Use startLine only for a finding that spans several lines.
```

Merge the results into one entry per file and remove duplicate findings. A finding that spans files, like a caller left out of date, goes on the file that needs the fix. When that file is not in the PR, it goes on the changed file that breaks it, in the summary only.

## Step 5: Verify

Check every `bug` and `question` before it is posted. One false bug makes the human distrust every other grade.

- **Single pass**: re-read the cited code and try to prove the finding wrong: a guard upstream, a type that rules out the input, a test that covers the case.
- **Split**: launch one `general-purpose` verifier subagent with the findings (path, lines, summary, trigger), the worktree path and the diff command from Step 4b, without the reviewers' reasoning or grades. It answers CONFIRMED, REFUTED or UNSURE for each finding, with one line of evidence.

Then apply the result:

- CONFIRMED: a bug, or a question it found a trigger for, is a bug and the file gets grade 1. A question confirmed without a trigger stays a question.
- REFUTED: drop the finding and grade the file again without it.
- UNSURE: keep it as a question and cap the file at 2.

## Step 6: Show the result

Fill the summary and the inline comments from `references/summary.md`. Show the user the summary and the inline comments exactly as they will be posted, and the list of grade 5 files. Then ask which of these to do. More than one can be chosen:

1. Post the summary comment.
2. Post the inline comments.
3. Mark the grade 5 files as viewed. This changes only the user's own view of the PR.

Do nothing on GitHub before the user answers.

## Step 7: Post

First check that nobody pushed since the review: `gh pr view <N> --json headRefOid` must still return the reviewed SHA. If it does not, post nothing, tell the user, and offer to run the review again.

### Summary comment

- First run: `gh pr comment <N> --body-file <file>`.
- A previous summary exists: edit it in place so the PR keeps a single summary: `gh api repos/<owner>/<repo>/issues/comments/<id> --method PATCH -F body=@<file>`.

### Inline comments

Post them as one review with the `COMMENT` event:

```bash
gh api repos/<owner>/<repo>/pulls/<N>/reviews --method POST --input <review.json>
```

```json
{
  "commit_id": "<headRefOid>",
  "event": "COMMENT",
  "body": "Details for the findings in the review summary.",
  "comments": [{ "path": "src/api/auth.ts", "line": 47, "side": "RIGHT", "body": "…" }]
}
```

- `line` must be a line of the new file inside a diff hunk. For a finding over several lines, add `start_line` and `"start_side": "RIGHT"`, in the same hunk as `line`. GitHub rejects the whole review otherwise, so keep such a finding out of the inline comments.
- When the summary is not posted, the review body lists the findings that cannot go inline instead of pointing to the summary.
- Skip a finding that an existing comment already raises on the same lines.
- Never approve and never request changes. The grades point the human to the right files. The decision stays theirs.

### Mark files as viewed

For each grade 5 file, using the PR's `id` from Step 1:

```bash
gh api graphql \
  -f query='mutation($pr: ID!, $path: String!) { markFileAsViewed(input: {pullRequestId: $pr, path: $path}) { clientMutationId } }' \
  -f pr=<id> -f path=<path>
```

Only grade 5. A grade 4 file still needs a skim. GitHub clears the viewed mark by itself when a later commit changes the file.

## Rules

- Nothing posted mentions Claude, AI, an agent or a bot, and nothing is signed.
- No emojis. English. Short sentences.
- Be constructive: say what to change and why, not what the author did wrong.
- When a finding depends on how a library behaves, check its docs with context7 before posting.
- Style issues from the project's rules are nits: summary only, never inline, never part of a grade.
- Never edit the PR's files, push to its branch, or resolve other people's comments.
