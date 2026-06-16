#!/usr/bin/env sh
set -eu

. "${0%/*}/project_meta.sh"

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <version-tag> <release-dir>" >&2
  exit 1
fi

VERSION="$1"
RELEASE_DIR="$2"
TAP_REPO="webxsid/homebrew-tap"
TAP_BRANCH="main"
TAP_DIR="$(mktemp -d)"
trap 'rm -rf "${TAP_DIR}"' EXIT INT TERM

if [ -z "${HOMEBREW_TAP_GITHUB_TOKEN:-}" ]; then
  echo "HOMEBREW_TAP_GITHUB_TOKEN is required" >&2
  exit 1
fi

mkdir -p "${TAP_DIR}/Formula"
FORMULA_NAME="${CRONA_BREW_FORMULA_NAME:-crona}"
case "${VERSION}" in
  *-beta*)
    FORMULA_NAME="${CRONA_BREW_FORMULA_NAME:-crona-beta}"
    ;;
esac
FORMULA_PATH="${TAP_DIR}/Formula/${FORMULA_NAME}.rb"

CRONA_HOMEBREW_BASE_URL="https://github.com/${PROJECT_REPO}/releases/download/${VERSION}" \
  sh "${0%/*}/generate_homebrew_formula.sh" "${VERSION}" "${RELEASE_DIR}" "${RELEASE_DIR}/checksums.txt" "${FORMULA_PATH}" "${FORMULA_NAME}"

git clone --branch "${TAP_BRANCH}" --single-branch \
  "https://x-access-token:${HOMEBREW_TAP_GITHUB_TOKEN}@github.com/${TAP_REPO}.git" \
  "${TAP_DIR}/repo" >/dev/null 2>&1

cp "${FORMULA_PATH}" "${TAP_DIR}/repo/Formula/${FORMULA_NAME}.rb"

cd "${TAP_DIR}/repo"
git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"
git add "Formula/${FORMULA_NAME}.rb"

if git diff --cached --quiet; then
  echo "Homebrew formula is already up to date"
  exit 0
fi

git commit -m "crona ${VERSION}"
git push origin "HEAD:${TAP_BRANCH}"
