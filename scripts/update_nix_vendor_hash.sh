#!/usr/bin/env sh
# Recomputes nix/vendor.json's vendorHash after go.mod or go.sum change: builds crona-git
# with a placeholder hash and keeps the hash Nix reports for the vendored modules.
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
VENDOR="${ROOT_DIR}/nix/vendor.json"
FAKE="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
export NIX_CONFIG="${NIX_CONFIG:-}
experimental-features = nix-command flakes"

cd "${ROOT_DIR}"
previous="$(cat "${VENDOR}")"
restore() { printf '%s\n' "${previous}" > "${VENDOR}"; }
printf '{\n  "vendorHash": "%s"\n}\n' "${FAKE}" > "${VENDOR}"

log="$(nix build --no-link ".#crona-git" 2>&1 || true)"
hash="$(printf '%s\n' "${log}" | sed -n 's/^[[:space:]]*got:[[:space:]]*\(sha256-[A-Za-z0-9+/=]*\).*/\1/p' | tail -1)"
if [ -z "${hash}" ]; then
  restore
  printf '%s\n' "${log}" | tail -20 >&2
  echo "could not read the vendor hash from nix build" >&2
  exit 1
fi
printf '{\n  "vendorHash": "%s"\n}\n' "${hash}" > "${VENDOR}"
echo "nix/vendor.json vendorHash = ${hash}"
