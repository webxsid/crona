#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
COVERAGE_DIR="${ROOT_DIR}/coverage"
GOCACHE_DIR="${GOCACHE:-/tmp/crona-go-cache}"

mkdir -p "${COVERAGE_DIR}"
rm -f "${COVERAGE_DIR}"/*.out "${COVERAGE_DIR}/coverage.out" "${COVERAGE_DIR}/summary.txt"

first=1
for module in shared kernel tui cli; do
  profile="${COVERAGE_DIR}/${module}.out"
  (
    cd "${ROOT_DIR}/${module}"
    GOCACHE="${GOCACHE_DIR}" go test ./... -covermode=count -coverprofile="${profile}"
  )
  if [ "${first}" -eq 1 ]; then
    cat "${profile}" > "${COVERAGE_DIR}/coverage.out"
    first=0
  else
    tail -n +2 "${profile}" >> "${COVERAGE_DIR}/coverage.out"
  fi
done

{
  echo "Crona coverage summary"
  echo
  for module in shared kernel tui cli; do
    echo "== ${module} =="
    GOCACHE="${GOCACHE_DIR}" go tool cover -func="${COVERAGE_DIR}/${module}.out" | tail -n 1
    echo
  done
  echo "== combined =="
  GOCACHE="${GOCACHE_DIR}" go tool cover -func="${COVERAGE_DIR}/coverage.out" | tail -n 1
} | tee "${COVERAGE_DIR}/summary.txt"
