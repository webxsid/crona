#!/usr/bin/env sh
set -eu

. "${0%/*}/project_meta.sh"

ACTION="${1:-test}"
SYSTEM_BREW_BIN="${BREW_BIN:-$(command -v brew)}"
TEST_ROOT="${CRONA_BREW_TEST_ROOT:-/tmp/crona-homebrew-test}"
TAP_NAME="${CRONA_BREW_TAP_NAME:-crona/local-test}"
PREFIX_ROOT="${CRONA_BREW_PREFIX_ROOT:-${TEST_ROOT}/homebrew}"
CELLAR_ROOT="${CRONA_BREW_CELLAR_ROOT:-${PREFIX_ROOT}/Cellar}"
CACHE_ROOT="${CRONA_BREW_CACHE_ROOT:-${TEST_ROOT}/Cache}"
REPO_ROOT="${CRONA_BREW_REPO_ROOT:-${PREFIX_ROOT}/Library/Homebrew}"
TAP_DIR="${CRONA_BREW_TAP_DIR:-${TEST_ROOT}}"
DIST_DIR="${CRONA_BREW_DIST_DIR:-${PWD}/dist}"
GO_CMD="${GO_CMD:-go}"
SYSTEM_BREW_REPO="$("${SYSTEM_BREW_BIN}" --repo)"
ISOLATED_BREW_BIN="${PREFIX_ROOT}/bin/brew"

cleanup_root() {
  rm -rf "${TEST_ROOT}"
}

current_goos() {
  case "$(uname -s)" in
    Darwin) printf 'darwin\n' ;;
    Linux) printf 'linux\n' ;;
    *)
      echo "unsupported host OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

current_goarch() {
  case "$(uname -m)" in
    arm64|aarch64) printf 'arm64\n' ;;
    x86_64|amd64) printf 'amd64\n' ;;
    *)
      echo "unsupported host arch: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

sha256_file() {
  file="$1"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{ print $1 }'
  else
    sha256sum "$file" | awk '{ print $1 }'
  fi
}

brew_env() {
  env -u HOMEBREW_PREFIX -u HOMEBREW_REPOSITORY -u HOMEBREW_CELLAR \
    HOMEBREW_CACHE="${CACHE_ROOT}" \
    HOMEBREW_NO_ANALYTICS=1 \
    HOMEBREW_NO_AUTO_UPDATE=1 \
    HOMEBREW_NO_INSTALL_CLEANUP=1 \
    PATH="/usr/bin:/bin:/usr/sbin:/sbin"
}

brew_cmd() {
  brew_env "${ISOLATED_BREW_BIN}" "$@"
}

ensure_dirs() {
  mkdir -p "${TEST_ROOT}" "${PREFIX_ROOT}/bin" "${PREFIX_ROOT}/Library/Taps" "${CELLAR_ROOT}" "${CACHE_ROOT}" "${TAP_DIR}/Formula"
  cp "${SYSTEM_BREW_BIN}" "${ISOLATED_BREW_BIN}"
  tmp_brew="$(mktemp)"
  awk -v prefix="${PREFIX_ROOT}" '
    $0 == "HOMEBREW_PREFIX=\"${HOMEBREW_BREW_FILE%/*/*}\"" {
      print "HOMEBREW_PREFIX=\"" prefix "\""
      next
    }
    { print }
  ' "${ISOLATED_BREW_BIN}" > "${tmp_brew}"
  mv "${tmp_brew}" "${ISOLATED_BREW_BIN}"
  chmod +x "${ISOLATED_BREW_BIN}"
  rm -rf "${REPO_ROOT}"
  mkdir -p "$(dirname "${REPO_ROOT}")"
  ln -sfn "${SYSTEM_BREW_REPO}" "${REPO_ROOT}"
}

archive_name() {
  version="$1"
  goos="$2"
  goarch="$3"
  printf 'crona-bundle-%s-%s-%s.zip\n' "${version}" "${goos}" "${goarch}"
}

expected_archive() {
  version="$1"
  goos="$2"
  goarch="$3"
  printf '%s/%s\n' "${DIST_DIR}" "$(archive_name "${version}" "${goos}" "${goarch}")"
}

current_platform_archive() {
  version="$1"
  expected_archive "${version}" "$(current_goos)" "$(current_goarch)"
}

version_from_archive() {
  archive="$1"
  basename "$archive" | sed -E 's/^crona-bundle-(.*)-(darwin|linux|windows)-(amd64|arm64)\.zip$/\1/'
}

snapshot_version_from_dist() {
  archive="$(find "${DIST_DIR}" -maxdepth 1 -name "crona-bundle-*-$(current_goos)-$(current_goarch).zip" | head -n 1)"
  if [ -z "${archive}" ]; then
    echo "unable to find current-platform GoReleaser archive in ${DIST_DIR}" >&2
    exit 1
  fi
  version_from_archive "${archive}"
}

verify_snapshot_artifacts() {
  version="$1"
  for goos in darwin linux windows; do
    for goarch in amd64 arm64; do
      if [ "${goos}" = "windows" ] && [ "${goarch}" = "arm64" ]; then
        continue
      fi
      archive="$(expected_archive "${version}" "${goos}" "${goarch}")"
      if [ ! -f "${archive}" ]; then
        echo "missing archive: ${archive}" >&2
        exit 1
      fi
      if ! grep "  $(basename "${archive}")$" "${DIST_DIR}/checksums.txt" >/dev/null 2>&1; then
        echo "checksums.txt is missing $(basename "${archive}")" >&2
        exit 1
      fi
    done
  done
}

generate_formula() {
  version="$1"
  archive_dir="$2"
  checksum_file="$3"
  output_path="$4"
  if [ -z "${CRONA_HOMEBREW_BASE_URL:-}" ]; then
    export CRONA_HOMEBREW_BASE_URL="file://$(cd "${archive_dir}" && pwd -P)"
  fi
  sh "${0%/*}/generate_homebrew_formula.sh" "${version}" "${archive_dir}" "${checksum_file}" "${output_path}"
}

run_status_checks() {
  expected_version="$1"
  status_output="$(PATH="${PREFIX_ROOT}/bin:/usr/bin:/bin:/usr/sbin:/sbin" crona update status)"
  printf '%s\n' "${status_output}" | grep -F "current: ${expected_version}" >/dev/null 2>&1
  printf '%s\n' "${status_output}" | grep -F "install-source: brew" >/dev/null 2>&1
  printf '%s\n' "${status_output}" | grep -F "update-command: brew upgrade crona" >/dev/null 2>&1
  if printf '%s\n' "${status_output}" | grep -Eq 'self-update|curl -fsSL https://crona.work/install.sh|go install github.com/webxsid/crona'; then
    echo "update status exposed a non-brew update path" >&2
    exit 1
  fi
}

run_brew_validation() {
  version="$1"
  brew_cmd tap "${TAP_NAME}" "${TAP_DIR}"
  brew_cmd install "${TAP_NAME}/crona"

  PATH="${PREFIX_ROOT}/bin:/usr/bin:/bin:/usr/sbin:/sbin" crona --version | grep -F "${version}" >/dev/null
  PATH="${PREFIX_ROOT}/bin:/usr/bin:/bin:/usr/sbin:/sbin" crona-kernel --version | grep -F "${version}" >/dev/null
  PATH="${PREFIX_ROOT}/bin:/usr/bin:/bin:/usr/sbin:/sbin" crona-tui --version | grep -F "${version}" >/dev/null

  run_status_checks "${version}"

  brew_cmd uninstall crona
  brew_cmd untap "${TAP_NAME}"

  if PATH="${PREFIX_ROOT}/bin" command -v crona >/dev/null 2>&1; then
    echo "crona still exists after uninstall" >&2
    exit 1
  fi
  if PATH="${PREFIX_ROOT}/bin" command -v crona-kernel >/dev/null 2>&1; then
    echo "crona-kernel still exists after uninstall" >&2
    exit 1
  fi
  if PATH="${PREFIX_ROOT}/bin" command -v crona-tui >/dev/null 2>&1; then
    echo "crona-tui still exists after uninstall" >&2
    exit 1
  fi
}

build_fake_release_dir() {
  version="$1"
  out_dir="$2"
  goos="$(current_goos)"
  goarch="$(current_goarch)"
  mkdir -p "${out_dir}"

  cli_bin="${out_dir}/crona"
  kernel_bin="${out_dir}/crona-kernel"
  tui_bin="${out_dir}/crona-tui"

  GOCACHE="${GOCACHE:-/tmp/crona-go-cache}" GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 \
    "${GO_CMD}" build -trimpath -ldflags "-s -w -X crona/shared/version.Version=${version}" -o "${cli_bin}" ./cli/cmd/crona
  GOCACHE="${GOCACHE:-/tmp/crona-go-cache}" GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 \
    "${GO_CMD}" build -trimpath -ldflags "-s -w -X crona/shared/version.Version=${version}" -o "${kernel_bin}" ./kernel/cmd/crona-kernel
  GOCACHE="${GOCACHE:-/tmp/crona-go-cache}" GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 \
    "${GO_CMD}" build -trimpath -ldflags "-s -w -X crona/shared/version.Version=${version}" -o "${tui_bin}" ./tui

  archive_file="${out_dir}/$(archive_name "${version}" "${goos}" "${goarch}")"
  (
    cd "${out_dir}"
    zip -q "$(basename "${archive_file}")" "$(basename "${cli_bin}")" "$(basename "${kernel_bin}")" "$(basename "${tui_bin}")"
  )
  {
    printf '%s  %s\n' "$(sha256_file "${archive_file}")" "$(basename "${archive_file}")"
  } > "${out_dir}/checksums.txt"
  rm -f "${cli_bin}" "${kernel_bin}" "${tui_bin}"
}

run_upgrade_validation() {
  base_dir="${TEST_ROOT}/upgrade"
  v1_dir="${base_dir}/v0.0.1-test"
  v2_dir="${base_dir}/v0.0.2-test"
  mkdir -p "${v1_dir}" "${v2_dir}"

  build_fake_release_dir "v0.0.1-test" "${v1_dir}"
  generate_formula "v0.0.1-test" "${v1_dir}" "${v1_dir}/checksums.txt" "${TAP_DIR}/Formula/crona.rb"
  brew_cmd tap "${TAP_NAME}" "${TAP_DIR}"
  brew_cmd install "${TAP_NAME}/crona"
  PATH="${PREFIX_ROOT}/bin:/usr/bin:/bin:/usr/sbin:/sbin" crona --version | grep -F "v0.0.1-test" >/dev/null

  build_fake_release_dir "v0.0.2-test" "${v2_dir}"
  generate_formula "v0.0.2-test" "${v2_dir}" "${v2_dir}/checksums.txt" "${TAP_DIR}/Formula/crona.rb"
  brew_cmd upgrade crona
  PATH="${PREFIX_ROOT}/bin:/usr/bin:/bin:/usr/sbin:/sbin" crona --version | grep -F "v0.0.2-test" >/dev/null
  run_status_checks "v0.0.2-test"

  brew_cmd uninstall crona
  brew_cmd untap "${TAP_NAME}"
}

run_snapshot_validation() {
  echo "Running GoReleaser snapshot build"
  GOCACHE="${GOCACHE:-/tmp/crona-go-cache}" goreleaser release --snapshot --clean --skip=publish

  snapshot_version="$(snapshot_version_from_dist)"
  verify_snapshot_artifacts "${snapshot_version}"
  generate_formula "${snapshot_version}" "${DIST_DIR}" "${DIST_DIR}/checksums.txt" "${TAP_DIR}/Formula/crona.rb"

  if [ "${ACTION}" = "generate" ] || [ "${ACTION}" = "generate-only" ]; then
    printf 'Generated local formula at %s\n' "${TAP_DIR}/Formula/crona.rb"
    return 0
  fi

  run_brew_validation "${snapshot_version}"
}

main() {
  case "${ACTION}" in
    test)
      cleanup_root
      ensure_dirs
      trap cleanup_root EXIT INT TERM
      run_snapshot_validation
      ;;
    generate-only|generate)
      cleanup_root
      ensure_dirs
      run_snapshot_validation
      ;;
    upgrade-test)
      cleanup_root
      ensure_dirs
      trap cleanup_root EXIT INT TERM
      run_upgrade_validation
      ;;
    clean)
      cleanup_root
      ;;
    *)
      echo "usage: $0 [test|generate-only|upgrade-test|clean]" >&2
      exit 1
      ;;
  esac
}

main "$@"
