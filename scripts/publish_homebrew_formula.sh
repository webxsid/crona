#!/usr/bin/env sh
set -eu

. "${0%/*}/project_meta.sh"

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <version-tag> <release-dir>" >&2
  exit 1
fi

VERSION="$1"
RELEASE_DIR="$2"
DIST_DIR="${CRONA_GORELEASER_DIST_DIR:-${PWD}/dist}"
TAP_REPO="webxsid/homebrew-tap"
TAP_BRANCH="main"
TAP_DIR="$(mktemp -d)"
PRIMARY_FORMULA_NAME="${CRONA_BREW_FORMULA_NAME:-crona}"
trap 'rm -rf "${TAP_DIR}"' EXIT INT TERM

if [ -z "${HOMEBREW_TAP_GITHUB_TOKEN:-}" ]; then
  echo "HOMEBREW_TAP_GITHUB_TOKEN is required" >&2
  exit 1
fi

mkdir -p "${TAP_DIR}/Formula"

if [ ! -f "${DIST_DIR}/checksums.txt" ]; then
  echo "missing GoReleaser checksums file: ${DIST_DIR}/checksums.txt" >&2
  exit 1
fi

generate_formula() {
  formula_name="$1"
  formula_path="${TAP_DIR}/Formula/${formula_name}.rb"
  CRONA_HOMEBREW_BASE_URL="https://github.com/${PROJECT_REPO}/releases/download/${VERSION}" \
    sh "${0%/*}/generate_homebrew_formula.sh" "${VERSION}" "${DIST_DIR}" "${DIST_DIR}/checksums.txt" "${formula_path}" "${formula_name}"
}

formula_names() {
  case "${VERSION}" in
    *-beta*)
      printf '%s\n' "${CRONA_BREW_FORMULA_NAME:-crona-beta}"
      ;;
    *)
      printf '%s\n' "${PRIMARY_FORMULA_NAME}"
      if [ "${PRIMARY_FORMULA_NAME}" != "crona-beta" ]; then
        printf '%s\n' "crona-beta"
      fi
      ;;
  esac
}

for formula_name in $(formula_names); do
  generate_formula "${formula_name}"
done

git clone --branch "${TAP_BRANCH}" --single-branch \
  "https://x-access-token:${HOMEBREW_TAP_GITHUB_TOKEN}@github.com/${TAP_REPO}.git" \
  "${TAP_DIR}/repo" >/dev/null 2>&1

for formula_name in $(formula_names); do
  cp "${TAP_DIR}/Formula/${formula_name}.rb" "${TAP_DIR}/repo/Formula/${formula_name}.rb"
done

cd "${TAP_DIR}/repo"
git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"
for formula_name in $(formula_names); do
  git add "Formula/${formula_name}.rb"
done

if git diff --cached --quiet; then
  echo "Homebrew formula is already up to date"
  exit 0
fi

git commit -m "crona ${VERSION}"
git push origin "HEAD:${TAP_BRANCH}"
