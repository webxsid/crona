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
windows amd64
windows arm64
"

SIZE_REPORT="${RELEASE_DIR}/sizes.txt"

binary_name() {
  base="$1"
  goos="$2"
  if [ "${goos}" = "windows" ]; then
    printf '%s.exe\n' "${base}"
  else
    printf '%s\n' "${base}"
  fi
}

bundle_name() {
  version="$1"
  goos="$2"
  goarch="$3"
  printf 'crona-bundle-%s-%s-%s.zip\n' "${version}" "${goos}" "${goarch}"
}

file_size_bytes() {
  wc -c < "$1" | tr -d ' '
}

human_size() {
  bytes="$1"
  awk -v bytes="${bytes}" 'BEGIN {
    split("B KB MB GB", units, " ");
    value = bytes + 0;
    idx = 1;
    while (value >= 1024 && idx < 4) {
      value /= 1024;
      idx++;
    }
    printf "%.1f %s", value, units[idx];
  }'
}

sum_file_sizes() {
  total=0
  for path in "$@"; do
    total=$((total + $(file_size_bytes "${path}")))
  done
  printf '%s\n' "${total}"
}

report_target_sizes() {
  goos="$1"
  goarch="$2"
  bundle_path="$3"
  shift 3
  raw_total="$(sum_file_sizes "$@")"
  bundle_total="$(file_size_bytes "${bundle_path}")"
  savings=$((raw_total - bundle_total))
  savings_pct=0
  if [ "${raw_total}" -gt 0 ]; then
    savings_pct=$((100 * savings / raw_total))
  fi

  {
    printf '%s/%s\n' "${goos}" "${goarch}"
    for path in "$@"; do
      name="$(basename "${path}")"
      bytes="$(file_size_bytes "${path}")"
      printf '  %-44s %10s  (%s)\n' "${name}" "${bytes}" "$(human_size "${bytes}")"
    done
    printf '  %-44s %10s  (%s)\n' "$(basename "${bundle_path}")" "${bundle_total}" "$(human_size "${bundle_total}")"
    printf '  %-44s %10s  (%s)\n' "raw total" "${raw_total}" "$(human_size "${raw_total}")"
    printf '  %-44s %10s  (%s saved, %s%%)\n' "bundle delta" "${savings}" "$(human_size "${savings}")" "${savings_pct}"
    printf '\n'
  } >> "${SIZE_REPORT}"
}

build_release_binary() {
  output="$1"
  package="$2"
  goos="$3"
  goarch="$4"
  env CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" GOCACHE="${GOCACHE_DIR}" \
    go build -trimpath -ldflags="-s -w" -o "${output}" "${package}"
}

archive_target_bundle() {
  bundle_output="$1"
  shift
  (
    cd "${RELEASE_DIR}"
    zip -q "${bundle_output}" "$@"
  )
}

rm -rf "${RELEASE_DIR}"
mkdir -p "${RELEASE_DIR}"
: > "${SIZE_REPORT}"

expected_files() {
  echo "install-crona-tui.sh"
  echo "install-crona-tui.ps1"
  echo "crona-assets-${VERSION}.tar.gz"
  echo "sizes.txt"
  printf '%s\n' "${TARGETS}" | while read -r GOOS GOARCH; do
    [ -n "${GOOS}" ] || continue
    bundle_name "${VERSION}" "${GOOS}" "${GOARCH}"
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

if ! command -v zip >/dev/null 2>&1; then
  echo "zip is required to create Windows release bundles" >&2
  exit 1
fi

echo "${TARGETS}" | while read -r GOOS GOARCH; do
  [ -n "${GOOS}" ] || continue

  echo "Building ${GOOS}/${GOARCH}"
  cli_output="$(binary_name "crona-${VERSION}-${GOOS}-${GOARCH}" "${GOOS}")"
  kernel_output="$(binary_name "crona-kernel-${VERSION}-${GOOS}-${GOARCH}" "${GOOS}")"
  tui_output="$(binary_name "crona-tui-${VERSION}-${GOOS}-${GOARCH}" "${GOOS}")"
  bundle_output="$(bundle_name "${VERSION}" "${GOOS}" "${GOARCH}")"
  build_release_binary "${RELEASE_DIR}/${cli_output}" ./cli/cmd/crona "${GOOS}" "${GOARCH}"
  build_release_binary "${RELEASE_DIR}/${kernel_output}" ./kernel/cmd/crona-kernel "${GOOS}" "${GOARCH}"
  build_release_binary "${RELEASE_DIR}/${tui_output}" ./tui "${GOOS}" "${GOARCH}"
  archive_target_bundle "${bundle_output}" "${cli_output}" "${kernel_output}" "${tui_output}"
  report_target_sizes "${GOOS}" "${GOARCH}" "${RELEASE_DIR}/${bundle_output}" \
    "${RELEASE_DIR}/${cli_output}" "${RELEASE_DIR}/${kernel_output}" "${RELEASE_DIR}/${tui_output}"
  rm -f "${RELEASE_DIR}/${cli_output}" "${RELEASE_DIR}/${kernel_output}" "${RELEASE_DIR}/${tui_output}"
done

sed \
  -e "s#__VERSION__#${VERSION}#g" \
  -e "s#__REPO__#${PROJECT_REPO}#g" \
  "${ROOT_DIR}/scripts/install_tui.sh.tmpl" > "${RELEASE_DIR}/install-crona-tui.sh"
chmod +x "${RELEASE_DIR}/install-crona-tui.sh"

sed \
  -e "s#__VERSION__#${VERSION}#g" \
  -e "s#__REPO__#${PROJECT_REPO}#g" \
  "${ROOT_DIR}/scripts/install_tui.ps1.tmpl" > "${RELEASE_DIR}/install-crona-tui.ps1"

mkdir -p "${RELEASE_DIR}/assets"
cp -R "${ROOT_DIR}/assets/export" "${RELEASE_DIR}/assets/"
(
  cd "${RELEASE_DIR}/assets"
  tar -czf "../crona-assets-${VERSION}.tar.gz" export
)
rm -rf "${RELEASE_DIR}/assets"

(
  cd "${RELEASE_DIR}"
  : > checksums.txt
  for file in *; do
    [ "${file}" = "checksums.txt" ] && continue
    shasum -a 256 "${file}" >> checksums.txt
  done
)

verify_release_artifacts

echo
echo "Artifact size summary"
cat "${SIZE_REPORT}"

echo "Release artifacts written to ${RELEASE_DIR}"
