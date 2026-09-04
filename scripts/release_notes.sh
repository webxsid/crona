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
RELEASE_NOTES_DIR="${ROOT_DIR}/docs/release-notes"
RELEASE_NOTE_FILE="${RELEASE_NOTES_DIR}/${TAG}.md"

if [ -f "${RELEASE_NOTE_FILE}" ]; then
	awk '
	NR == 1 && $0 == "---" {
		frontmatter = 1;
		next;
	}
	frontmatter && $0 == "---" {
		frontmatter = 0;
		frontmatter_done = 1;
		next;
	}
	frontmatter_done {
		frontmatter_done = 0;
		if ($0 == "") {
			next;
		}
	}
	!frontmatter {
		print;
	}
	' "${RELEASE_NOTE_FILE}"
	exit 0
fi

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
