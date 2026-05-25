#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
GOCACHE_DIR="${GOCACHE:-/tmp/crona-go-cache}"
GO_LINES_BIN="${GOLINES_BIN:-}"

if [ -z "${GO_LINES_BIN}" ]; then
  if command -v golines >/dev/null 2>&1; then
    GO_LINES_BIN="$(command -v golines)"
  elif [ -x "${ROOT_DIR}/bin/golines" ]; then
    GO_LINES_BIN="${ROOT_DIR}/bin/golines"
  else
    echo "golines is not installed. Run: make install-fmt" >&2
    exit 1
  fi
fi

for module in shared kernel tui cli; do
  (
    cd "${ROOT_DIR}/${module}"
    dirs="$(GOCACHE="${GOCACHE_DIR}" go list -f '{{.Dir}}' ./...)"
    if [ -n "${dirs}" ]; then
      GOCACHE="${GOCACHE_DIR}" gofmt -w ${dirs}
      "${GO_LINES_BIN}" -w --max-len=100 --shorten-comments --ignore-generated ${dirs}
    fi
  )
done
