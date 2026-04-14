#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPECFLOW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_ROOT="$(pwd)"
MANIFEST="${SPECFLOW_ROOT}/tooling/manifest.tsv"
FORCE="false"
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
  ./specflow/tooling/init.sh [--force]

Options:
  --force   Overwrite existing files managed by specFlow.
EOF
}

while (($# > 0)); do
  case "$1" in
    --force)
      FORCE="true"
      ;;
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

if [[ ! -f "${MANIFEST}" ]]; then
  echo "Manifest not found: ${MANIFEST}" >&2
  exit 1
fi

copied=0
skipped=0
failures=0

while IFS=$'\t' read -r src_rel dest_rel mode; do
  [[ -z "${src_rel}" ]] && continue
  src="${SPECFLOW_ROOT}/${src_rel}"
  dest="${TARGET_ROOT}/${dest_rel}"

  if [[ ! -f "${src}" ]]; then
    echo "Template not found: ${src}" >&2
    exit 1
  fi

  mkdir -p "$(dirname "${dest}")"

  if [[ -e "${dest}" && "${FORCE}" != "true" ]]; then
    echo "Skip existing ${mode} file: ${dest_rel}"
    skipped=$((skipped + 1))
    continue
  fi

  if [[ -e "${dest}" ]] && is_managed_entry_file "${dest_rel}"; then
    if ! replace_managed_block "${src}" "${dest}"; then
      echo "Failed to install managed block into existing ${dest_rel}" >&2
      failures=$((failures + 1))
      continue
    fi
    copied=$((copied + 1))
    echo "Installed managed block ${dest_rel}"
    continue
  fi

  cp "${src}" "${dest}"
  copied=$((copied + 1))
  echo "Installed ${dest_rel}"
done < "${MANIFEST}"

if ((failures > 0)); then
  echo "specFlow init failed. copied=${copied} skipped=${skipped} failures=${failures}" >&2
  exit 1
fi

echo "specFlow init completed. copied=${copied} skipped=${skipped}"
echo "If you want Git hooks to use .githooks, run: git config core.hooksPath .githooks"
