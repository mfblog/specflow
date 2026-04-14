#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPECFLOW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_ROOT="$(pwd)"
MANIFEST="${SPECFLOW_ROOT}/tooling/manifest.tsv"
FORCE="false"

usage() {
  cat <<'EOF'
Usage:
  ./specflow/tooling/upgrade.sh [--force]

Options:
  --force   Overwrite existing files, including project bootstrap files.
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

updated=0
skipped=0

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

  if [[ "${mode}" != "framework" && "${FORCE}" != "true" ]]; then
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

echo "specFlow upgrade completed. updated=${updated} skipped=${skipped}"
