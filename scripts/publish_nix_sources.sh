#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <version-tag>" >&2
  exit 1
fi

VERSION="$1"
VERSION_NO_V="${VERSION#v}"
DIST_DIR="${CRONA_GORELEASER_DIST_DIR:-${PWD}/dist}"
CHECKSUMS="${DIST_DIR}/checksums.txt"
SOURCES="${ROOT_DIR}/nix/sources.json"
BRANCH="${CRONA_NIX_SOURCES_BRANCH:-main}"

if [ ! -f "${CHECKSUMS}" ]; then
  echo "missing GoReleaser checksums file: ${CHECKSUMS}" >&2
  exit 1
fi

# The flake names its packages after the Homebrew formulae: crona and crona-beta.
if printf '%s' "${VERSION}" | grep -Eq -- '-beta'; then
  PACKAGE="crona-beta"
else
  PACKAGE="crona"
fi

checksum_for() {
  file="crona-bundle-${VERSION_NO_V}-$1.zip"
  sum="$(awk -v name="${file}" '$2 == name { print $1; exit }' "${CHECKSUMS}")"
  if [ -z "${sum}" ]; then
    echo "missing checksum for ${file}" >&2
    exit 1
  fi
  printf '%s\n' "${sum}"
}

TMP="$(mktemp)"
trap 'rm -f "${TMP}"' EXIT INT TERM
jq --indent 2 \
  --arg package "${PACKAGE}" \
  --arg version "${VERSION_NO_V}" \
  --arg darwinArm "$(checksum_for darwin-arm64)" \
  --arg linuxArm "$(checksum_for linux-arm64)" \
  --arg linuxIntel "$(checksum_for linux-amd64)" \
  '.[$package] = {
    version: $version,
    platforms: {
      "aarch64-darwin": $darwinArm,
      "aarch64-linux": $linuxArm,
      "x86_64-linux": $linuxIntel
    }
  }' "${SOURCES}" > "${TMP}"

if [ -n "${CRONA_NIX_SOURCES_DRY_RUN:-}" ]; then
  cat "${TMP}"
  exit 0
fi

# The tag points at a detached commit, so the sources are committed onto the branch tip.
cd "${ROOT_DIR}"
git fetch origin "${BRANCH}"
git checkout -B "${BRANCH}" "origin/${BRANCH}"
cp "${TMP}" "${SOURCES}"
if git diff --quiet -- "${SOURCES}"; then
  echo "nix/sources.json already points ${PACKAGE} at ${VERSION_NO_V}"
  exit 0
fi
git -c user.name="github-actions[bot]" -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
  commit -m "nix: point ${PACKAGE} at ${VERSION}" -- "${SOURCES}"
git push origin "${BRANCH}"
