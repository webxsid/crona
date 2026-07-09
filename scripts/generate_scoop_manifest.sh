#!/usr/bin/env sh
set -eu

. "${0%/*}/project_meta.sh"

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <version-tag> <all|stable|beta> <output-dir>" >&2
  exit 1
fi

VERSION="$1"
CHANNEL="$2"
OUTPUT_DIR="$3"
ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
RELEASE_DIR="${ROOT_DIR}/release/${VERSION}"
TEMPLATE_PATH="${ROOT_DIR}/templates/scoop.json.tmpl"
VERSION_NO_V="${VERSION#v}"
DESCRIPTION="local-first work tracker for developers"
HOMEPAGE="https://crona.work"
LICENSE="MIT"

require_file() {
  if [ ! -f "$1" ]; then
    echo "missing required file: $1" >&2
    exit 1
  fi
}

ensure_release_dir() {
  if [ ! -d "${RELEASE_DIR}" ] || [ ! -f "${RELEASE_DIR}/checksums.txt" ]; then
    echo "Building local release artifacts for ${VERSION}"
    sh "${ROOT_DIR}/scripts/build_release.sh" "${VERSION}"
  fi
}

checksum_for() {
  file="$1"
  awk -v name="$file" '$2 == name { print $1; exit }' "${RELEASE_DIR}/checksums.txt"
}

manifest_name() {
  case "$1" in
    stable) printf '%s\n' "crona.json" ;;
    beta) printf '%s\n' "crona-beta.json" ;;
    *) exit 1 ;;
  esac
}

release_channel() {
  case "$1" in
    stable) printf '%s\n' "stable" ;;
    beta) printf '%s\n' "beta" ;;
    *) exit 1 ;;
  esac
}

release_url() {
  arch="$1"
  printf 'https://github.com/%s/releases/download/%s/crona-bundle-%s-windows-%s.zip\n' "${PROJECT_REPO}" "${VERSION}" "${VERSION_NO_V}" "${arch}"
}

render_manifest() {
  flavor="$1"
  output_path="${OUTPUT_DIR}/$(manifest_name "${flavor}")"
  amd64_asset="crona-bundle-${VERSION_NO_V}-windows-amd64.zip"
  arm64_asset="crona-bundle-${VERSION_NO_V}-windows-arm64.zip"
  amd64_hash="$(checksum_for "${amd64_asset}")"
  arm64_hash="$(checksum_for "${arm64_asset}")"

  if [ -z "${amd64_hash}" ] || [ -z "${arm64_hash}" ]; then
    echo "missing Windows checksums for ${VERSION}" >&2
    exit 1
  fi

  mkdir -p "${OUTPUT_DIR}"
  sed \
    -e "s|__VERSION__|${VERSION_NO_V}|g" \
    -e "s|__DESCRIPTION__|${DESCRIPTION}|g" \
    -e "s|__HOMEPAGE__|${HOMEPAGE}|g" \
    -e "s|__LICENSE__|${LICENSE}|g" \
    -e "s|__RELEASE_CHANNEL__|$(release_channel "${flavor}")|g" \
    -e "s|__URL_AMD64__|$(release_url amd64)|g" \
    -e "s|__HASH_AMD64__|${amd64_hash}|g" \
    -e "s|__URL_ARM64__|$(release_url arm64)|g" \
    -e "s|__HASH_ARM64__|${arm64_hash}|g" \
    "${TEMPLATE_PATH}" > "${output_path}"
}

ensure_release_dir
require_file "${TEMPLATE_PATH}"
require_file "${RELEASE_DIR}/checksums.txt"
require_file "${RELEASE_DIR}/crona-bundle-${VERSION_NO_V}-windows-amd64.zip"
require_file "${RELEASE_DIR}/crona-bundle-${VERSION_NO_V}-windows-arm64.zip"

case "${CHANNEL}" in
  all)
    if printf '%s' "${VERSION}" | grep -Eq -- '-beta'; then
      render_manifest beta
    else
      render_manifest stable
      render_manifest beta
    fi
    ;;
  stable)
    render_manifest stable
    ;;
  beta)
    render_manifest beta
    ;;
  *)
    echo "channel must be one of: all, stable, beta" >&2
    exit 1
    ;;
esac

echo "Local Scoop manifests written to ${OUTPUT_DIR}"
