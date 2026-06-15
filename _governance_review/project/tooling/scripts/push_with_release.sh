#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
Usage: push_with_release.sh

Push the current SpecFlow branch to origin.
Before pushing, update existing project entry files' specFlow Addendum blocks
from the current templates.
When the current branch is main, if the current tooling fingerprint does not
already have a remote release tag, create and push specflow-tooling-<fingerprint>
to trigger the GitHub Release workflow.
USAGE
}

for arg in "$@"; do
  case "${arg}" in
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 1
      ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
PROJECT_ROOT="$(cd "${REPO_ROOT}/.." && pwd)"

extract_managed_block() {
  local path="$1"
  awk '
    $0 == "==SPECFLOW:BEGIN==" {
      if (seen_begin) {
        err = "managed block begin marker must appear exactly once"
        exit 1
      }
      seen_begin = 1
      in_block = 1
    }
    in_block { print }
    $0 == "==SPECFLOW:END==" {
      if (seen_end) {
        err = "managed block end marker must appear exactly once"
        exit 1
      }
      if (!in_block) {
        err = "managed block end marker is out of order"
        exit 1
      }
      seen_end = 1
      in_block = 0
    }
    END {
      if (err != "") {
        print "Error: " err > "/dev/stderr"
        exit 1
      }
      if (!seen_begin || !seen_end || in_block) {
        print "Error: managed block markers are missing or out of order" > "/dev/stderr"
        exit 1
      }
    }
  ' "${path}"
}

replace_managed_block() {
  local target="$1"
  local block_file="$2"
  local temp_file
  temp_file="$(mktemp)"

  if ! grep -q '^==SPECFLOW:BEGIN==$' "${target}"; then
    # Markers not found in target — insert block at the beginning.
    {
      cat "${block_file}"
      echo ""
      cat "${target}"
    } >"${temp_file}"
  else
    awk -v block_file="${block_file}" '
      BEGIN {
        while ((getline line < block_file) > 0) {
          block = block line ORS
        }
        close(block_file)
        sub(ORS "$", "", block)
      }
      $0 == "==SPECFLOW:BEGIN==" {
        if (seen_begin) {
          err = "managed block begin marker must appear exactly once"
          exit 1
        }
        seen_begin = 1
        in_block = 1
        print block
        next
      }
      $0 == "==SPECFLOW:END==" {
        if (!in_block) {
          err = "managed block end marker is out of order"
          exit 1
        }
        seen_end = 1
        in_block = 0
        next
      }
      in_block { next }
      { print }
      END {
        if (err != "") {
          print "Error: " err > "/dev/stderr"
          exit 1
        }
        if (!seen_begin || !seen_end || in_block) {
          print "Error: managed block markers are missing or out of order" > "/dev/stderr"
          exit 1
        }
      }
    ' "${target}" >"${temp_file}" || {
      rm -f "${temp_file}"
      return 1
    }
  fi

  if cmp -s "${target}" "${temp_file}"; then
    rm -f "${temp_file}"
  else
    mv "${temp_file}" "${target}"
    return 2
  fi
}

sync_existing_entry_blocks() {
  local entry source target block_file changed found status
  changed=0
  found=0
  for entry in AGENTS.md CLAUDE.md GEMINI.md; do
    source="${REPO_ROOT}/templates/${entry}"
    target="${PROJECT_ROOT}/${entry}"
    [[ -f "${target}" ]] || continue
    found=1
    if [[ ! -f "${source}" ]]; then
      echo "Error: entry template missing: ${source}" >&2
      exit 1
    fi

    block_file="$(mktemp)"
    extract_managed_block "${source}" >"${block_file}" || {
      rm -f "${block_file}"
      exit 1
    }
    if replace_managed_block "${target}" "${block_file}"; then
      :
    else
      status=$?
      rm -f "${block_file}"
      if [[ "${status}" -eq 2 ]]; then
        echo "Updated ${entry} specFlow Addendum."
        changed=1
      else
        exit "${status}"
      fi
    fi
    rm -f "${block_file}"
  done

  if [[ "${found}" -eq 0 ]]; then
    echo "No existing project entry files found to update."
  elif [[ "${changed}" -eq 0 ]]; then
    echo "Existing project entry Addendum blocks are already current."
  fi
}

cd "${REPO_ROOT}"

branch="$(git branch --show-current)"
if [[ -z "${branch}" ]]; then
  echo "Error: detached HEAD is not supported. Check out a branch before pushing." >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Error: working tree is not clean. Commit or stash changes before pushing." >&2
  exit 1
fi

remote_url="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "${remote_url}" ]]; then
  echo "Error: git remote 'origin' is missing." >&2
  exit 1
fi

sync_existing_entry_blocks

echo "Pushing ${branch} to origin..."
git push -u origin "${branch}"

if [[ "${branch}" != "main" ]]; then
  echo "Current branch is ${branch}, not main."
  echo "Release tag push is skipped."
  exit 0
fi

fingerprint="$("${REPO_ROOT}/tooling/scripts/tooling_fingerprint.sh" --short)"
tag="specflow-tooling-${fingerprint}"

if git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; then
  echo "Release tag already exists on origin: ${tag}"
  echo "No release tag push needed."
  exit 0
fi

if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
  tag_commit="$(git rev-list -n 1 "${tag}")"
  head_commit="$(git rev-parse HEAD)"
  if [[ "${tag_commit}" != "${head_commit}" ]]; then
    echo "Error: local tag ${tag} exists but does not point to HEAD." >&2
    echo "Delete or inspect the local tag manually before pushing a release." >&2
    exit 1
  fi
else
  git tag "${tag}"
fi

echo "Pushing release tag ${tag}..."
git push origin "${tag}"
echo "Release workflow triggered by ${tag}."
