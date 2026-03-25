#!/usr/bin/env sh
set -eu

. "${0%/*}/project_meta.sh"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <version-tag>" >&2
  exit 1
fi

VERSION="$1"
ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
RELEASE_DIR="${ROOT_DIR}/release/${VERSION}"
GOCACHE_DIR="${GOCACHE:-/tmp/crona-go-release-cache}"

TARGETS="
darwin arm64
darwin amd64
linux amd64
linux arm64
"

rm -rf "${RELEASE_DIR}"
mkdir -p "${RELEASE_DIR}"

expected_files() {
  echo "install-crona-tui.sh"
  echo "crona-assets-${VERSION}.tar.gz"
  printf '%s\n' "${TARGETS}" | while read -r GOOS GOARCH; do
    [ -n "${GOOS}" ] || continue
    echo "crona-kernel-${VERSION}-${GOOS}-${GOARCH}"
    echo "crona-tui-${VERSION}-${GOOS}-${GOARCH}"
  done
}

verify_release_artifacts() {
  if [ ! -f "${RELEASE_DIR}/checksums.txt" ]; then
    echo "Missing required release artifact: checksums.txt" >&2
    exit 1
  fi
  for file in $(expected_files); do
    if [ ! -f "${RELEASE_DIR}/${file}" ]; then
      echo "Missing required release artifact: ${file}" >&2
      exit 1
    fi
    if ! grep "  ${file}\$" "${RELEASE_DIR}/checksums.txt" >/dev/null 2>&1; then
      echo "checksums.txt is missing ${file}" >&2
      exit 1
    fi
  done
}

for target in ${TARGETS}; do
  :
done

echo "${TARGETS}" | while read -r GOOS GOARCH; do
  [ -n "${GOOS}" ] || continue

  echo "Building ${GOOS}/${GOARCH}"
  env CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" GOCACHE="${GOCACHE_DIR}" \
    go build -o "${RELEASE_DIR}/crona-kernel-${VERSION}-${GOOS}-${GOARCH}" ./kernel/cmd/crona-kernel
  env CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" GOCACHE="${GOCACHE_DIR}" \
    go build -o "${RELEASE_DIR}/crona-tui-${VERSION}-${GOOS}-${GOARCH}" ./tui
done

sed \
  -e "s#__VERSION__#${VERSION}#g" \
  -e "s#__REPO__#${PROJECT_REPO}#g" \
  "${ROOT_DIR}/scripts/install_tui.sh.tmpl" > "${RELEASE_DIR}/install-crona-tui.sh"
chmod +x "${RELEASE_DIR}/install-crona-tui.sh"

mkdir -p "${RELEASE_DIR}/assets"
cp -R "${ROOT_DIR}/assets/export" "${RELEASE_DIR}/assets/"
(
  cd "${RELEASE_DIR}/assets"
  tar -czf "../crona-assets-${VERSION}.tar.gz" export
)
rm -rf "${RELEASE_DIR}/assets"

(
  cd "${RELEASE_DIR}"
  shasum -a 256 ./* > checksums.txt
)

verify_release_artifacts

echo "Release artifacts written to ${RELEASE_DIR}"
