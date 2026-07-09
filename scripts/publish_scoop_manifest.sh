#!/usr/bin/env sh
set -eu

. "${0%/*}/project_meta.sh"
ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <version-tag>" >&2
  exit 1
fi

VERSION="$1"
DIST_DIR="${CRONA_GORELEASER_DIST_DIR:-${PWD}/dist}"
BUCKET_REPO="webxsid/scoop-bucket"
BUCKET_BRANCH="main"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT INT TERM

if [ -z "${SCOOP_BUCKET_GITHUB_TOKEN:-}" ]; then
  echo "SCOOP_BUCKET_GITHUB_TOKEN is required" >&2
  exit 1
fi

if [ ! -f "${DIST_DIR}/checksums.txt" ]; then
  echo "missing GoReleaser checksums file: ${DIST_DIR}/checksums.txt" >&2
  exit 1
fi

channel() {
  if printf '%s' "${VERSION}" | grep -Eq -- '-beta'; then
    printf '%s\n' "beta"
    return
  fi
  printf '%s\n' "stable"
}

manifest_name() {
  case "$1" in
    stable) printf '%s\n' "crona.json" ;;
    beta) printf '%s\n' "crona-beta.json" ;;
    *) exit 1 ;;
  esac
}

FLAVOR="$(channel)"
MANIFEST_NAME="$(manifest_name "${FLAVOR}")"

mkdir -p "${TMP_DIR}/release/${VERSION}"
cp "${DIST_DIR}/checksums.txt" "${TMP_DIR}/release/${VERSION}/"
cp "${DIST_DIR}/crona-bundle-${VERSION#v}-windows-amd64.zip" "${TMP_DIR}/release/${VERSION}/"
cp "${DIST_DIR}/crona-bundle-${VERSION#v}-windows-arm64.zip" "${TMP_DIR}/release/${VERSION}/"

(
  cd "${TMP_DIR}"
  CRONA_RELEASE_DIR="${TMP_DIR}/release/${VERSION}" \
    sh "${ROOT_DIR}/scripts/generate_scoop_manifest.sh" "${VERSION}" "${FLAVOR}" "${TMP_DIR}/generated"
)

git clone --branch "${BUCKET_BRANCH}" --single-branch \
  "https://x-access-token:${SCOOP_BUCKET_GITHUB_TOKEN}@github.com/${BUCKET_REPO}.git" \
  "${TMP_DIR}/repo" >/dev/null 2>&1

mkdir -p "${TMP_DIR}/repo/bucket"
cp "${TMP_DIR}/generated/${MANIFEST_NAME}" "${TMP_DIR}/repo/bucket/${MANIFEST_NAME}"

cd "${TMP_DIR}/repo"
git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"
git add "bucket/${MANIFEST_NAME}"

if git diff --cached --quiet; then
  echo "Scoop manifest is already up to date"
  exit 0
fi

git commit -m "crona ${VERSION}"
git push origin "HEAD:${BUCKET_BRANCH}"
