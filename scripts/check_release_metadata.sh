#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"

fail() {
  echo "release metadata check failed: $*" >&2
  exit 1
}

project_version="$(awk '/^PROJECT_VERSION := / { print $3 }' "${ROOT_DIR}/Makefile")"
shared_version="$(sed -n 's/^var Version = "\([^"]*\)".*/\1/p' "${ROOT_DIR}/shared/version/version.go")"
protocol_version="$(sed -n 's/^const Version = "\([^"]*\)".*/\1/p' "${ROOT_DIR}/shared/protocol/version.go")"

[ -n "${project_version}" ] || fail "PROJECT_VERSION is missing from Makefile"
[ -n "${shared_version}" ] || fail "shared version is missing from shared/version/version.go"
[ "${project_version}" = "${shared_version}" ] || fail "Makefile version ${project_version} does not match shared version ${shared_version}"
[ "${protocol_version}" = "1.0" ] || fail "protocol version must remain 1.0 before stable external GUI compatibility work"

tag="v${project_version}"
release_notes_file="${ROOT_DIR}/docs/release-notes/${tag}.md"
for doc in README.md docs/install.md docs/changelog.md; do
  if ! grep -F "${tag}" "${ROOT_DIR}/${doc}" >/dev/null 2>&1; then
    fail "${doc} does not mention current release tag ${tag}"
  fi
done

if [ ! -f "${release_notes_file}" ]; then
  fail "missing public release notes file ${release_notes_file}"
fi

if ! sh "${ROOT_DIR}/scripts/release_notes.sh" "${tag}" >/dev/null; then
  fail "could not generate release notes for ${tag} from docs/release-notes"
fi

echo "release metadata ok: ${tag}, protocol ${protocol_version}"
