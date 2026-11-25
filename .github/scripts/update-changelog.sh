#!/bin/sh
# Prepends the release notes GoReleaser wrote to dist/CHANGELOG.md to CHANGELOG.md,
# so the file and the GitHub release show the same notes.
# Usage: update-changelog.sh <tag> <owner/repo>
set -eu
tag=$1
repository=$2
version=${tag#v}

# A rerun of the same release must not add the entry twice.
if grep -q "^#\{1,2\} \[$version\]" CHANGELOG.md; then
  echo "CHANGELOG.md already has $tag"
  exit 0
fi

previous=$(git describe --tags --abbrev=0 "$tag^" 2>/dev/null || true)
link="https://github.com/$repository/releases/tag/$tag"
if [ -n "$previous" ]; then
  link="https://github.com/$repository/compare/$previous...$tag"
fi

# Same heading levels as the semantic-release entries below: # for major and minor, ## for patch.
heading="#"
case "$version" in
  *.*.0) ;;
  *) heading="##" ;;
esac

updated=$(mktemp)
{
  printf '%s [%s](%s) (%s)\n\n' "$heading" "$version" "$link" "$(date -u +%Y-%m-%d)"
  sed '/^## Changelog$/d' dist/CHANGELOG.md
  printf '\n'
  cat CHANGELOG.md
} > "$updated"
mv "$updated" CHANGELOG.md
