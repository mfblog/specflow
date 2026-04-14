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

  cp "${src}" "${dest}"
  copied=$((copied + 1))
  echo "Installed ${dest_rel}"
done < "${MANIFEST}"

echo "specFlow init completed. copied=${copied} skipped=${skipped}"
echo "If you want Git hooks to use .githooks, run: git config core.hooksPath .githooks"
