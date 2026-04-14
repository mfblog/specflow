#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPECFLOW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_ROOT="$(pwd)"
MANIFEST="${SPECFLOW_ROOT}/tooling/manifest.tsv"
MANAGED_BEGIN='<!-- SPECFLOW:BEGIN -->'
MANAGED_END='<!-- SPECFLOW:END -->'

is_managed_entry_file() {
  case "$1" in
    AGENTS.md|GEMINI.md|CLAUDE.md) return 0 ;;
    *) return 1 ;;
  esac
}

extract_managed_block() {
  local file="$1"

  awk -v begin="${MANAGED_BEGIN}" -v end="${MANAGED_END}" '
    $0 == begin {
      begin_count++
      if (begin_count > 1 || in_block) {
        exit 2
      }
      in_block = 1
    }
    in_block {
      print
    }
    $0 == end {
      if (!in_block) {
        exit 3
      }
      end_count++
      in_block = 0
    }
    END {
      if (begin_count != 1 || end_count != 1 || in_block) {
        exit 4
      }
    }
  ' "${file}"
}

managed_blocks_match() {
  local src="$1"
  local dest="$2"
  local src_block
  local dest_block

  src_block="$(mktemp)"
  dest_block="$(mktemp)"
  extract_managed_block "${src}" > "${src_block}" || {
    rm -f "${src_block}" "${dest_block}"
    return 1
  }
  extract_managed_block "${dest}" > "${dest_block}" || {
    rm -f "${src_block}" "${dest_block}"
    return 1
  }

  if cmp -s "${src_block}" "${dest_block}"; then
    rm -f "${src_block}" "${dest_block}"
    return 0
  fi

  rm -f "${src_block}" "${dest_block}"
  return 1
}

replace_managed_block() {
  local src="$1"
  local dest="$2"
  local block_file
  local tmp_file

  block_file="$(mktemp)"
  tmp_file="$(mktemp)"

  extract_managed_block "${src}" > "${block_file}"

  awk -v begin="${MANAGED_BEGIN}" -v end="${MANAGED_END}" -v block_file="${block_file}" '
    $0 == begin {
      if (replaced) {
        exit 2
      }
      while ((getline line < block_file) > 0) {
        print line
      }
      close(block_file)
      in_block = 1
      replaced = 1
      next
    }
    $0 == end {
      if (!in_block) {
        exit 3
      }
      in_block = 0
      next
    }
    !in_block {
      print
    }
    END {
      if (!replaced || in_block) {
        exit 4
      }
    }
  ' "${dest}" > "${tmp_file}"

  mv "${tmp_file}" "${dest}"
  rm -f "${block_file}"
}

usage() {
  cat <<'EOF'
Usage:
  ./specflow/tooling/upgrade.sh
EOF
}

while (($# > 0)); do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
  shift
done

updated=0
skipped=0
failures=0

while IFS=$'\t' read -r src_rel dest_rel mode; do
  [[ -z "${src_rel}" ]] && continue
  src="${SPECFLOW_ROOT}/${src_rel}"
  dest="${TARGET_ROOT}/${dest_rel}"

  if [[ ! -e "${dest}" ]]; then
    mkdir -p "$(dirname "${dest}")"
    cp "${src}" "${dest}"
    echo "Installed missing ${dest_rel}"
    updated=$((updated + 1))
    continue
  fi

  if is_managed_entry_file "${dest_rel}"; then
    if managed_blocks_match "${src}" "${dest}"; then
      continue
    fi

    if ! replace_managed_block "${src}" "${dest}"; then
      echo "Failed to update managed block: ${dest_rel}" >&2
      failures=$((failures + 1))
      continue
    fi

    echo "Updated managed block ${dest_rel}"
    updated=$((updated + 1))
    continue
  fi

  if [[ "${mode}" != "framework" ]]; then
    echo "Skip project-owned file: ${dest_rel}"
    skipped=$((skipped + 1))
    continue
  fi

  if cmp -s "${src}" "${dest}"; then
    continue
  fi

  cp "${src}" "${dest}"
  echo "Updated ${dest_rel}"
  updated=$((updated + 1))
done < "${MANIFEST}"

if ((failures > 0)); then
  echo "specFlow upgrade failed. updated=${updated} skipped=${skipped} failures=${failures}" >&2
  exit 1
fi

echo "specFlow upgrade completed. updated=${updated} skipped=${skipped}"
