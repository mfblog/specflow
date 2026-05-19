#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
Usage: push_with_release.sh

Push the current SpecFlow branch to origin.
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

echo "Pushing ${branch} to origin..."
git push -u origin "${branch}"

if [[ "${branch}" != "main" ]]; then
  echo "Current branch is ${branch}, not main."
  echo "Release tag push is skipped."
  exit 0
fi

fingerprint_root=""
parent_root="$(cd "${REPO_ROOT}/.." && pwd)"
if [[ -f "${parent_root}/specflow/tooling/manifest.tsv" ]]; then
  fingerprint_root="${parent_root}"
else
  fingerprint_root="$(mktemp -d)"
  trap 'rm -rf "${fingerprint_root}"' EXIT
  ln -s "${REPO_ROOT}" "${fingerprint_root}/specflow"
fi

fingerprint="$("${fingerprint_root}/specflow/tooling/scripts/tooling_fingerprint.sh" --short)"
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
