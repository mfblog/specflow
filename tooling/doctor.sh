#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPECFLOW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_ROOT="$(pwd)"
MANIFEST="${SPECFLOW_ROOT}/tooling/manifest.tsv"
failures=0

while IFS=$'\t' read -r _ dest_rel _; do
  [[ -z "${dest_rel}" ]] && continue
  if [[ ! -e "${TARGET_ROOT}/${dest_rel}" ]]; then
    echo "MISSING ${dest_rel}"
    failures=$((failures + 1))
  fi
done < "${MANIFEST}"

if [[ -f "${TARGET_ROOT}/AGENTS.md" && -f "${TARGET_ROOT}/GEMINI.md" ]]; then
  if ! cmp -s "${TARGET_ROOT}/AGENTS.md" "${TARGET_ROOT}/GEMINI.md"; then
    echo "DIFF AGENTS.md and GEMINI.md are inconsistent"
    failures=$((failures + 1))
  fi
fi

if command -v git >/dev/null 2>&1; then
  hook_path="$(git -C "${TARGET_ROOT}" config --get core.hooksPath || true)"
  if [[ "${hook_path}" != ".githooks" ]]; then
    echo "WARN git core.hooksPath is not .githooks"
  fi
fi

if ((failures > 0)); then
  echo "specFlow doctor failed: ${failures} issue(s)"
  exit 1
fi

echo "specFlow doctor passed"
