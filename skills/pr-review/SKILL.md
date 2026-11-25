---
name: pr-review
description: Performs an automated GitHub PR code review and posts the review as comments directly on the PR. Use this skill whenever the user wants to review a pull request, analyze PR changes, comment on a PR, or do a code review on GitHub. Triggers on phrases like "review PR", "review pull request", "comment on PR", "review this PR", "give feedback on PR", "code review PR #N". Always use this skill when the user mentions reviewing a GitHub PR — even if they just say "review PR 123" without more context.
---

# PR Review Skill

This skill performs an automated code review of a GitHub Pull Request and posts the review as inline and summary comments directly on the PR using the GitHub CLI (`gh`).

---

## Prerequisites

- `gh` CLI must be installed and authenticated (`gh auth status`)
- The user must be inside a GitHub repository (or provide `owner/repo`)
- The PR must exist and be accessible

---

## Workflow

### Step 1: Ask for PR Number

If the user hasn't already provided a PR number, ask for it:

```
Which PR number would you like me to review?
```

### Step 2: Gather PR Information

Run the following commands to collect context:

```bash
# Get PR metadata
gh pr view <PR_NUMBER> --json title,body,author,baseRefName,headRefName,additions,deletions,changedFiles

# Get the diff
gh pr diff <PR_NUMBER>

# Get list of changed files
gh pr view <PR_NUMBER> --json files --jq '.files[].path'
```

If the repo is not the current directory, use `--repo owner/repo` flag.

### Step 3: Analyze the PR

Review the diff and metadata carefully. Your analysis should cover:

1. **Summary** – What does this PR do? Is the purpose clear from the title/description?
2. **Code Quality** – Are there readability, naming, or style issues?
3. **Logic & Correctness** – Are there bugs, off-by-one errors, null/undefined risks, race conditions?
4. **Security** – Any injection risks, exposed secrets, missing auth checks?
5. **Performance** – Any N+1 queries, unnecessary loops, memory issues?
7. **Breaking Changes** – Does this change any public APIs or interfaces?

### Step 4: Post the Review

Use `gh pr review` to post structured feedback:

#### Option A: Approve (if the PR looks good)

```bash
gh pr review <PR_NUMBER> --approve --body "$(cat <<'EOF'
## All's good

<your summary here>

### Highlights
- <positive point 1>
- <positive point 2>
EOF
)"
```

#### Option B: Request Changes

```bash
gh pr review <PR_NUMBER> --request-changes --body "$(cat <<'EOF'
## Code Review

### Summary
<brief summary of what the PR does>

### Issues Found
<list issues>

### Suggestions
<list suggestions>

### Questions
<any clarifying questions>
EOF
)"
```

#### Option C: Comment (neutral feedback, no approval/rejection)

```bash
gh pr review <PR_NUMBER> --comment --body "$(cat <<'EOF'
## Review Feedback

<feedback here>
EOF
)"
```

### Step 5: Add Inline Comments (optional, for specific lines)

For specific line-level feedback, use the GitHub API directly:

```bash
# Get the PR's latest commit SHA
COMMIT=$(gh pr view <PR_NUMBER> --json commits --jq '.commits[-1].oid')

# Post an inline comment
gh api repos/{owner}/{repo}/pulls/<PR_NUMBER>/comments \
  --method POST \
  --field body="Your inline comment here" \
  --field commit_id="$COMMIT" \
  --field path="path/to/file.ts" \
  --field line=42
```

Use `gh repo view --json nameWithOwner --jq '.nameWithOwner'` to get `owner/repo`.

---

## Review Format Template

Use this structure for the main review body:

```markdown
# Review

## What's Good
- <positive thing 1>
- <positive thing 2>

---

## Issues & Suggestions

### Bugs / Correctness
- [ ] `path/to/file.ts` line X: <issue description>

### Security
- [ ] <security concern if any>

### Performance
- [ ] <perf concern if any>

### Code Quality
- [ ] <style/readability issue>

---

### Questions
1. <clarifying question>

---

### Verdict
**[APPROVE / REQUEST CHANGES / COMMENT]** — <one-line summary>
```

---

## Error Handling

- If `gh` is not authenticated: tell the user to run `gh auth login`
- If the PR is not found: confirm the PR number and repo
- If the diff is very large (>1000 lines): warn the user and review the most critical files first, then offer to continue
- If the repo can't be inferred: ask the user for `owner/repo`

---

## Mandatory Instructions

- Never use emojis in reviews
- Do not write too long sentences, be clear and concise
- Always be constructive — phrase issues as suggestions, not criticisms
- Skip auto-generated files: `bun.lock`, `pnpm-lock.yaml`, `*.min.*`, migration files unless they look suspicious
- Focus on code security, code duplication, best practices
- Feel free to use context7@claude-plugins-official to check documentation
- Be aware of already done reviews to reduce comments duplication

---
## Tips

- For large PRs, focus on files with the most logic changes (not generated files, lock files, etc.)
- Use checkboxes `- [ ]` for actionable items so the PR author can track them
