#!/usr/bin/env sh
set -eu

. "${0%/*}/project_meta.sh"

if [ "$#" -lt 4 ] || [ "$#" -gt 5 ]; then
  echo "usage: $0 <version> <archive-dir> <checksums-file> <output-path> [formula-name]" >&2
  exit 1
fi

VERSION="$1"
ARCHIVE_DIR="$2"
CHECKSUMS_FILE="$3"
OUTPUT_PATH="$4"
FORMULA_NAME="${5:-crona}"
BASE_URL="${CRONA_HOMEBREW_BASE_URL:-}"

abs_path() {
  path="$1"
  dir="$(dirname "$path")"
  base="$(basename "$path")"
  (cd "$dir" && printf '%s/%s\n' "$(pwd -P)" "$base")
}

archive_dir="$(abs_path "$ARCHIVE_DIR")"
checksums_path="$(abs_path "$CHECKSUMS_FILE")"
output_path="$(abs_path "$OUTPUT_PATH")"

formula_class_name() {
  case "$1" in
    crona) printf '%s\n' "Crona" ;;
    crona-beta) printf '%s\n' "CronaBeta" ;;
    *) printf '%s\n' "${1}" | awk '
      {
        out = ""
        n = split($0, parts, /[-_]/)
        for (i = 1; i <= n; i++) {
          part = parts[i]
          if (part == "") continue
          out = out toupper(substr(part, 1, 1)) tolower(substr(part, 2))
        }
        print out
      }'
      ;;
  esac
}

CLASS_NAME="$(formula_class_name "$FORMULA_NAME")"

if [ ! -f "$checksums_path" ]; then
  echo "missing checksums file: $checksums_path" >&2
  exit 1
fi

if [ -z "$BASE_URL" ]; then
  BASE_URL="file://${archive_dir}"
fi

archive_name() {
  goos="$1"
  goarch="$2"
  printf 'crona-bundle-%s-%s-%s.zip\n' "${VERSION}" "${goos}" "${goarch}"
}

member_name() {
  binary="$1"
  goos="$2"
  goarch="$3"
  printf '%s-v%s-%s-%s\n' "${binary}" "${VERSION#v}" "${goos}" "${goarch}"
}

emit_install_lines() {
  indent="$1"
  goos="$2"
  goarch="$3"
  printf '%sbin.install "%s" => "crona"\n' "$indent" "$(member_name crona "$goos" "$goarch")"
  printf '%sbin.install "%s" => "crona-kernel"\n' "$indent" "$(member_name crona-kernel "$goos" "$goarch")"
  printf '%sbin.install "%s" => "crona-tui"\n' "$indent" "$(member_name crona-tui "$goos" "$goarch")"
}

emit_install_dispatch() {
  printf '  def crona_runtime_home\n'
  printf '    if ENV["CRONA_HOME"] && !ENV["CRONA_HOME"].strip.empty?\n'
  printf '      return ENV["CRONA_HOME"].strip\n'
  printf '    end\n'
  printf '    home = Dir.home\n'
  printf '    if OS.mac?\n'
  printf '      File.join(home, "Library", "Application Support", "Crona")\n'
  printf '    else\n'
  printf '      data_home = ENV["XDG_DATA_HOME"]\n'
  printf '      if data_home && !data_home.strip.empty?\n'
  printf '        File.join(data_home.strip, "crona")\n'
  printf '      else\n'
  printf '        File.join(home, ".local", "share", "crona")\n'
  printf '      end\n'
  printf '    end\n'
  printf '  end\n'
  printf '\n'
  printf '  def write_install_source(source, formula_name)\n'
  printf '    runtime_home = crona_runtime_home\n'
  printf '    FileUtils.mkdir_p(runtime_home)\n'
  printf '    File.write(\n'
  printf '      File.join(runtime_home, "install.json"),\n'
  printf '      "{\\n  \\"installSource\\": \\"" + source + "\\",\\n  \\"brewFormula\\": \\"" + formula_name + "\\"\\n}\\n",\n'
  printf '    )\n'
  printf '  end\n'
  printf '\n'
  printf '  def install\n'
  printf '    if OS.mac?\n'
  printf '      if Hardware::CPU.arm?\n'
  emit_install_lines '        ' darwin arm64
  printf '      else\n'
  emit_install_lines '        ' darwin amd64
  printf '      end\n'
  printf '    elsif OS.linux?\n'
  printf '      if Hardware::CPU.arm?\n'
  emit_install_lines '        ' linux arm64
  printf '      else\n'
  emit_install_lines '        ' linux amd64
  printf '      end\n'
  printf '    end\n'
  printf '  end\n'
  printf '\n'
  printf '  def post_install\n'
  printf '    write_install_source("brew", "%s")\n' "$FORMULA_NAME"
  printf '  end\n'
}

archive_url() {
  file="$1"
  printf '%s/%s\n' "$BASE_URL" "$file"
}

checksum_for() {
  file="$1"
  awk -v name="$file" '$2 == name { print $1; exit }' "$checksums_path"
}

emit_platform_block() {
  ruby_block="$1"
  goos="$2"
  arm_file="$(archive_name "$goos" arm64)"
  intel_file="$(archive_name "$goos" amd64)"
  arm_path="${archive_dir}/${arm_file}"
  intel_path="${archive_dir}/${intel_file}"
  arm_sha="$(checksum_for "$arm_file" || true)"
  intel_sha="$(checksum_for "$intel_file" || true)"

  if [ ! -f "$arm_path" ] && [ ! -f "$intel_path" ]; then
    return 0
  fi

  printf '  on_%s do\n' "$ruby_block"
  if [ -f "$arm_path" ] && [ -f "$intel_path" ]; then
    if [ -z "$arm_sha" ] || [ -z "$intel_sha" ]; then
      echo "missing checksum for one or more archives in $goos" >&2
      exit 1
    fi
    printf '    if Hardware::CPU.arm?\n'
    printf '      url "%s"\n' "$(archive_url "$arm_file")"
    printf '      sha256 "%s"\n' "$arm_sha"
    printf '    else\n'
    printf '      url "%s"\n' "$(archive_url "$intel_file")"
    printf '      sha256 "%s"\n' "$intel_sha"
    printf '    end\n'
  elif [ -f "$arm_path" ]; then
    if [ -z "$arm_sha" ]; then
      echo "missing checksum for $arm_file" >&2
      exit 1
    fi
    printf '    url "%s"\n' "$(archive_url "$arm_file")"
    printf '    sha256 "%s"\n' "$arm_sha"
  else
    if [ -z "$intel_sha" ]; then
      echo "missing checksum for $intel_file" >&2
      exit 1
    fi
    printf '    url "%s"\n' "$(archive_url "$intel_file")"
    printf '    sha256 "%s"\n' "$intel_sha"
  fi
  printf '  end\n'
}

mkdir -p "$(dirname "$output_path")"

cat <<EOF > "$output_path"
require "fileutils"

class ${CLASS_NAME} < Formula
  desc "local-first work tracker for developers"
  homepage "https://crona.work"
  version "${VERSION#v}"

EOF

emit_platform_block macos darwin >> "$output_path"
printf '\n' >> "$output_path"
emit_platform_block linux linux >> "$output_path"

printf '\n' >> "$output_path"
emit_install_dispatch >> "$output_path"

cat <<'EOF' >> "$output_path"
  test do
    system "#{bin}/crona", "--version"
  end
end
EOF
