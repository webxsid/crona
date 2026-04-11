#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <version-tag>" >&2
  exit 1
fi

TAG="$1"
VERSION="${TAG#v}"
ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
CHANGELOG="${ROOT_DIR}/docs/changelog.md"

awk -v version="${VERSION}" -v tag="${TAG}" '
BEGIN {
  heading = "## [" version "]";
  found = 0;
}
index($0, heading) == 1 {
  found = 1;
  print "# Crona " tag;
  print "";
  next;
}
found && /^## \[/ {
  exit;
}
found {
  print;
}
END {
  if (!found) {
    exit 1;
  }
}
' "${CHANGELOG}"
