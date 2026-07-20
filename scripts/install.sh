#!/usr/bin/env sh

set -eu

readonly repository="kazuki-kanaya/lab-userctl"
readonly project="lab-userctl"
readonly install_dir="${INSTALL_DIR:-/usr/local/bin}"

fail() {
  printf '%s\n' "error: $*" >&2
  exit 1
}

case "$(uname -s)" in
  Linux) os="linux" ;;
  *) fail "only Linux is supported" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  aarch64 | arm64) arch="arm64" ;;
  *) fail "unsupported architecture: $(uname -m)" ;;
esac

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v sha256sum >/dev/null 2>&1 || fail "sha256sum is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

archive="${project}_${os}_${arch}.tar.gz"
download_url="https://github.com/${repository}/releases/latest/download"
temporary_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$temporary_dir"
}

trap cleanup EXIT INT TERM

printf 'Downloading %s...\n' "$archive"
curl --fail --location --silent --show-error \
  "$download_url/$archive" \
  --output "$temporary_dir/$archive"
curl --fail --location --silent --show-error \
  "$download_url/checksums.txt" \
  --output "$temporary_dir/checksums.txt"

grep --fixed-strings --quiet "  $archive" "$temporary_dir/checksums.txt" ||
  fail "checksum for $archive was not found"

(
  cd "$temporary_dir"
  sha256sum --check --ignore-missing checksums.txt
) || fail "checksum verification failed"

tar --extract --gzip --file "$temporary_dir/$archive" --directory "$temporary_dir"

if [ "$(id -u)" -eq 0 ]; then
  install --mode 0755 "$temporary_dir/$project" "$install_dir/$project"
else
  sudo install --mode 0755 "$temporary_dir/$project" "$install_dir/$project"
fi

printf 'Installed %s to %s/%s\n' "$project" "$install_dir" "$project"
