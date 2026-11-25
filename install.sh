#!/bin/sh
# Installs the latest ccc binary:
#   curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh
# CCC_INSTALL_DIR picks the folder (default ~/.local/bin). After this, `ccc update` keeps it up to date.
set -eu

releases_url="${CCC_RELEASES_URL:-https://github.com/Jerome1337/ccc}"
install_dir="${CCC_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
  echo "ccc: $1" >&2
  exit 1
}

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin | linux) ;;
  *) fail "unsupported system: $os" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) fail "unsupported architecture: $(uname -m)" ;;
esac

latest=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$releases_url/releases/latest") ||
  fail "can't reach $releases_url"
tag=${latest##*/}
asset="ccc_${os}_${arch}"

mkdir -p "$install_dir"
# Download next to the destination, so the final mv is an atomic rename on the same filesystem.
staged=$(mktemp "$install_dir/.ccc-install.XXXXXX")
trap 'rm -f "$staged"' EXIT

curl -fsSL -o "$staged" "$releases_url/releases/download/$tag/$asset" || fail "no $asset in release $tag"
expected=$(curl -fsSL "$releases_url/releases/download/$tag/checksums.txt" | awk -v name="$asset" '$2 == name { print $1 }')
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$staged" | awk '{ print $1 }')
else
  actual=$(shasum -a 256 "$staged" | awk '{ print $1 }')
fi
[ -n "$expected" ] && [ "$expected" = "$actual" ] || fail "checksum mismatch for $asset"

chmod 755 "$staged"
mv -f "$staged" "$install_dir/ccc"
trap - EXIT

echo "Installed ccc ${tag#v} to $install_dir/ccc"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add it to your PATH: export PATH=\"$install_dir:\$PATH\"" ;;
esac
echo "Then run: ccc update --global"
