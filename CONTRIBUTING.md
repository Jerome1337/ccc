# Contributing to ccc

Issues and pull requests are welcome. For a large change, open an issue first so we can agree on the approach before you write the code.

## Setup

You need Go, at the version in `go.mod`, and git.

```bash
git clone https://github.com/Jerome1337/ccc.git
cd ccc
go test -race ./...
```

## Checks

CI runs these on every pull request, along with `goreleaser check`. Run them before you push:

```bash
test -z "$(gofmt -l .)"
go vet ./...
go test -race ./...
```

## Trying a change

A build without a version reports `dev` and never updates itself. Point it at a temporary folder so it can't touch your own `~/.claude`:

```bash
go build -o /tmp/ccc .
CLAUDE_CONFIG_DIR="$(mktemp -d)" /tmp/ccc update --global --from ./path/to/config-repo
cd "$(mktemp -d)" && /tmp/ccc init --from ~/path/to/config-repo
```

A `ccc.yml` in the test repo makes ccc run `claude mcp add` and `claude plugin install`, so leave it out unless you're testing that step.

## Where the code lives

- `main.go` parses the arguments and runs the commands.
- `src/source` fetches a config repo: a local folder, or a `git clone` into a temporary folder.
- `src/install` maps the repo's `global/` or `project/` folder onto the target and writes the lock. `decide.go` holds every rule about a single file in one pure function, so a new rule is one more row in `decide_test.go`.
- `src/export` copies the config on a machine into a repo folder.
- `src/lock` reads and writes `ccc-lock.json`.
- `src/config` reads `ccc.yml`, and `src/claude` installs its MCP servers and plugins through the `claude` CLI.
- `src/selfupdate` replaces the binary with the latest release.
- `src/diff` and `src/term` print the diff before an overwrite, the colors and the y/N prompts.

## Commits and releases

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org), and they drive the releases. Each push to `main` runs the release workflow:

- [svu](https://github.com/caarlos0/svu) picks the next version from the commits and tags it: `fix:` is a patch, `feat:` a minor, and `!` or `BREAKING CHANGE:` a major.
- [GoReleaser](https://goreleaser.com) builds the binaries for macOS and Linux and uploads them with `checksums.txt` to the GitHub release, with notes grouped from the commits.
- The same notes go to the top of `CHANGELOG.md`, so don't edit that file by hand.

`chore:`, `docs:`, `test:` and `ci:` commits don't trigger a release.

## Style

- Format with `gofmt`.
- Add a test for any new rule or branch.
- Write a comment only for a reason the code can't show, in one line when you can.
- Keep docs, comments and error messages in plain English with short sentences.
