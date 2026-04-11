#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
GOCACHE_DIR="${GOCACHE:-/tmp/crona-go-cache}"
LINT_CACHE_DIR="${GOLANGCI_LINT_CACHE:-/tmp/crona-golangci-lint-cache}"
MASON_LINTER="/Users/sm2101/.local/share/nvim/mason/bin/golangci-lint"
LINTER_BIN="${GOLANGCI_LINT_BIN:-}"

if [ -z "${LINTER_BIN}" ]; then
  if command -v golangci-lint >/dev/null 2>&1; then
    LINTER_BIN="$(command -v golangci-lint)"
  elif [ -x "${MASON_LINTER}" ]; then
    LINTER_BIN="${MASON_LINTER}"
  else
    echo "golangci-lint is not installed. Run: make install-lint" >&2
    exit 1
  fi
fi

version_output="$("${LINTER_BIN}" version 2>/dev/null || true)"
case "${version_output}" in
  *" version 2."*|*" v2."*)
    ;;
  *)
    echo "golangci-lint v2 is required. Run: make install-lint" >&2
    echo "Found: ${version_output:-unknown}" >&2
    exit 1
    ;;
esac

for module in shared kernel tui cli; do
  (
    cd "${ROOT_DIR}/${module}"
    GOCACHE="${GOCACHE_DIR}" GOLANGCI_LINT_CACHE="${LINT_CACHE_DIR}" "${LINTER_BIN}" run ./...
  )
done
