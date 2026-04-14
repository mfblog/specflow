#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REGISTRY_FILE="${ROOT_DIR}/specflow/framework/docs/agent_guidelines/entry_index_registry.md"
MODE="sync"
STAGE_AFTER_SYNC="false"
MANAGED_BEGIN='<!-- SPECFLOW:BEGIN -->'
MANAGED_END='<!-- SPECFLOW:END -->'

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
  ./specflow/tooling/sync_entry_docs.sh [--check] [--stage] [--source <registered-file>]

Options:
  --check   Only verify consistency; do not overwrite files.
  --stage   After syncing, add registered entry files back to the git index.
  --source  Explicitly choose which registered entry file is the sync source.
EOF
}

SOURCE_FILE=""

while (($# > 0)); do
  case "$1" in
    --check)
      MODE="check"
      ;;
    --stage)
      STAGE_AFTER_SYNC="true"
      ;;
    --source)
      if (($# < 2)); then
        echo "Missing value for --source" >&2
        usage >&2
        exit 1
      fi
      SOURCE_FILE="$2"
      shift
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

if [[ ! -f "${REGISTRY_FILE}" ]]; then
  echo "Registry file not found: ${REGISTRY_FILE}" >&2
  exit 1
fi

mapfile -t REGISTERED_FILES < <(sed -n 's/^- `\([^`]*\)`$/\1/p' "${REGISTRY_FILE}")

if ((${#REGISTERED_FILES[@]} == 0)); then
  echo "No registered entry files found in ${REGISTRY_FILE}" >&2
  exit 1
fi

declare -A HASH_COUNTS=()
declare -A REGISTERED_SET=()

for rel_path in "${REGISTERED_FILES[@]}"; do
  abs_path="${ROOT_DIR}/${rel_path}"
  if [[ ! -f "${abs_path}" ]]; then
    echo "Registered entry file does not exist: ${rel_path}" >&2
    exit 1
  fi

  block_file="$(mktemp)"
  if ! extract_managed_block "${abs_path}" > "${block_file}"; then
    rm -f "${block_file}"
    echo "Registered entry file does not contain exactly one managed block: ${rel_path}" >&2
    exit 1
  fi

  file_hash="$(sha256sum "${block_file}" | awk '{print $1}')"
  rm -f "${block_file}"

  HASH_COUNTS["${file_hash}"]=$(( ${HASH_COUNTS["${file_hash}"]:-0} + 1 ))
  REGISTERED_SET["${rel_path}"]=1
done

if ((${#HASH_COUNTS[@]} == 1)); then
  echo "Managed entry blocks are already consistent."
  exit 0
fi

if [[ -n "${SOURCE_FILE}" && -z "${REGISTERED_SET["${SOURCE_FILE}"]:-}" ]]; then
  echo "Explicit source is not a registered entry file: ${SOURCE_FILE}" >&2
  exit 1
fi

if [[ -z "${SOURCE_FILE}" ]]; then
  STAGED_CHANGED=()
  while IFS= read -r rel_path; do
    [[ -z "${rel_path}" ]] && continue
    if [[ -n "${REGISTERED_SET["${rel_path}"]:-}" ]]; then
      STAGED_CHANGED+=("${rel_path}")
    fi
  done < <(cd "${ROOT_DIR}" && git diff --cached --name-only -- "${REGISTERED_FILES[@]}")

  if ((${#STAGED_CHANGED[@]} == 1)); then
    SOURCE_FILE="${STAGED_CHANGED[0]}"
  else
    echo "Registered entry docs differ and no explicit sync source was provided." >&2
    if ((${#STAGED_CHANGED[@]} > 0)); then
      printf 'Changed registered entry docs: %s\n' "${STAGED_CHANGED[*]}" >&2
    else
      echo "No unique changed registered entry doc could be inferred from the staged diff." >&2
    fi
    echo "Re-run with --source <registered-file> to confirm the sync source." >&2
    exit 1
  fi
fi

if [[ "${MODE}" == "check" ]]; then
  echo "Managed entry blocks are inconsistent. Sync source would be: ${SOURCE_FILE}" >&2
  exit 1
fi

SOURCE_ABS="${ROOT_DIR}/${SOURCE_FILE}"
UPDATED_FILES=()
for rel_path in "${REGISTERED_FILES[@]}"; do
  if [[ "${rel_path}" == "${SOURCE_FILE}" ]]; then
    continue
  fi

  target_abs="${ROOT_DIR}/${rel_path}"
  source_block="$(mktemp)"
  target_block="$(mktemp)"
  extract_managed_block "${SOURCE_ABS}" > "${source_block}"
  extract_managed_block "${target_abs}" > "${target_block}"
  if ! cmp -s "${source_block}" "${target_block}"; then
    replace_managed_block "${SOURCE_ABS}" "${target_abs}"
    UPDATED_FILES+=("${rel_path}")
  fi
  rm -f "${source_block}" "${target_block}"
done

if ((${#UPDATED_FILES[@]} == 0)); then
  echo "Managed entry blocks already matched source: ${SOURCE_FILE}"
else
  printf 'Synced managed entry blocks from %s to: %s\n' "${SOURCE_FILE}" "${UPDATED_FILES[*]}"
fi

if [[ "${STAGE_AFTER_SYNC}" == "true" ]]; then
  (
    cd "${ROOT_DIR}"
    git add -- "${REGISTERED_FILES[@]}"
  )
fi
