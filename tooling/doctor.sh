#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPECFLOW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_ROOT="$(pwd)"
MANIFEST="${SPECFLOW_ROOT}/tooling/manifest.tsv"
failures=0
MANAGED_BEGIN='<!-- SPECFLOW:BEGIN -->'
MANAGED_END='<!-- SPECFLOW:END -->'
SYNC_SCRIPT_REL="specflow/tooling/sync_entry_docs.sh"
EXPECTED_HOOK_LINE='"${REPO_ROOT}/specflow/tooling/sync_entry_docs.sh" --stage'

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

while IFS=$'\t' read -r _ dest_rel _; do
  [[ -z "${dest_rel}" ]] && continue
  if [[ ! -e "${TARGET_ROOT}/${dest_rel}" ]]; then
    echo "MISSING ${dest_rel}"
    failures=$((failures + 1))
  fi
done < "${MANIFEST}"

if [[ -f "${TARGET_ROOT}/AGENTS.md" ]]; then
  agents_block="$(mktemp)"
  if ! extract_managed_block "${TARGET_ROOT}/AGENTS.md" > "${agents_block}"; then
    echo "INVALID managed block in AGENTS.md"
    failures=$((failures + 1))
    rm -f "${agents_block}"
    agents_block=""
  fi
  for peer in GEMINI.md CLAUDE.md; do
    if [[ -f "${TARGET_ROOT}/${peer}" ]]; then
      peer_block="$(mktemp)"
      if ! extract_managed_block "${TARGET_ROOT}/${peer}" > "${peer_block}"; then
        echo "INVALID managed block in ${peer}"
        failures=$((failures + 1))
        rm -f "${peer_block}"
        continue
      fi
      if [[ -n "${agents_block:-}" ]] && ! cmp -s "${agents_block}" "${peer_block}"; then
        echo "DIFF managed blocks in AGENTS.md and ${peer}"
        failures=$((failures + 1))
      fi
      rm -f "${peer_block}"
    fi
  done
  [[ -n "${agents_block:-}" ]] && rm -f "${agents_block}"
fi

if command -v git >/dev/null 2>&1; then
  hook_path="$(git -C "${TARGET_ROOT}" config --get core.hooksPath || true)"
  if [[ "${hook_path}" != ".githooks" ]]; then
    echo "WARN git core.hooksPath is not .githooks"
  fi
fi

if [[ ! -f "${TARGET_ROOT}/${SYNC_SCRIPT_REL}" ]]; then
  echo "MISSING ${SYNC_SCRIPT_REL}"
  failures=$((failures + 1))
fi

hook_file="${TARGET_ROOT}/.githooks/pre-commit"
if [[ -f "${hook_file}" ]] && ! grep -Fq "${EXPECTED_HOOK_LINE}" "${hook_file}"; then
  echo "INVALID .githooks/pre-commit does not call ${SYNC_SCRIPT_REL}"
  failures=$((failures + 1))
fi

if ((failures > 0)); then
  echo "specFlow doctor failed: ${failures} issue(s)"
  exit 1
fi

echo "specFlow doctor passed"
