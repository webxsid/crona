#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <version-tag>" >&2
  exit 1
fi

VERSION="$1"
VERSION_NO_V="${VERSION#v}"
SOURCES="${ROOT_DIR}/nix/sources.json"
BRANCH="${CRONA_NIX_SOURCES_BRANCH:-main}"
export NIX_CONFIG="${NIX_CONFIG:-}
experimental-features = nix-command flakes"

. "${ROOT_DIR}/scripts/project_meta.sh"

# The flake names its packages after the Homebrew formulae: crona and crona-beta.
if printf '%s' "${VERSION}" | grep -Eq -- '-beta'; then
  PACKAGE="crona-beta"
else
  PACKAGE="crona"
fi

# The flake builds each channel from source at its tag: the hash of the tag's source tree,
# as fetchFromGitHub unpacks it, and the tag's own vendored-modules hash.
HASH="$(nix flake prefetch --json "github:${PROJECT_REPO}/${VERSION}" | jq -r '.hash')"
VENDOR_HASH="$(git -C "${ROOT_DIR}" show "${VERSION}:nix/vendor.json" | jq -r '.vendorHash')"
if [ -z "${HASH}" ] || [ "${HASH}" = "null" ] || [ -z "${VENDOR_HASH}" ] || [ "${VENDOR_HASH}" = "null" ]; then
  echo "could not resolve the source or vendor hash for ${VERSION}" >&2
  exit 1
fi

TMP="$(mktemp)"
trap 'rm -f "${TMP}"' EXIT INT TERM
jq --indent 2 \
  --arg package "${PACKAGE}" \
  --arg version "${VERSION_NO_V}" \
  --arg rev "${VERSION}" \
  --arg hash "${HASH}" \
  --arg vendorHash "${VENDOR_HASH}" \
  '.[$package] = { version: $version, rev: $rev, hash: $hash, vendorHash: $vendorHash }' \
  "${SOURCES}" > "${TMP}"

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
