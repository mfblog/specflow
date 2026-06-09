#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
Usage: pull_with_release.sh

Pull the current SpecFlow branch from origin.
After pulling, update existing project entry files' specFlow Addendum blocks
from the current templates.
Then make sure the current platform's specflowctl and specflow-reader binaries
match the pulled tooling source fingerprint. Missing or stale binaries are
downloaded from the matching GitHub Release.
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
BIN_DIR="${REPO_ROOT}/tooling/bin"
download_dir=""
trap 'rm -rf "${download_dir:-}"' EXIT

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

platform_suffix() {
  local os arch
  case "$(uname -s)" in
    Linux)
      os="linux"
      ;;
    Darwin)
      os="darwin"
      ;;
    MINGW*|MSYS*|CYGWIN*)
      os="windows"
      ;;
    *)
      echo "Error: unsupported operating system: $(uname -s)" >&2
      return 1
      ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64)
      arch="amd64"
      ;;
    aarch64|arm64)
      arch="arm64"
      ;;
    *)
      echo "Error: unsupported CPU architecture: $(uname -m)" >&2
      return 1
      ;;
  esac

  if [[ "${os}" == "windows" ]]; then
    printf '%s-%s.exe\n' "${os}" "${arch}"
  else
    printf '%s-%s\n' "${os}" "${arch}"
  fi
}

read_binary_fingerprint() {
  local binary_path="$1"
  if [[ ! -x "${binary_path}" ]]; then
    return 1
  fi
  "${binary_path}" __print-build-fingerprint 2>/dev/null || return 1
}

verify_checksums() {
	local dir="$1"
	shift
	local -a names=("$@")
	if [[ "$#" -eq 0 ]]; then
		echo "Error: no binary names provided for checksum verification." >&2
		return 1
	fi
	local current_sums status
	current_sums="$(mktemp)"
	local awk_expr=''
	for name in "${names[@]}"; do
		if [[ -n "${awk_expr}" ]]; then
			awk_expr+=" || "
		fi
		awk_expr+='$2 == "'"${name}"'"'
	done
	awk "${awk_expr} { print }" "${dir}/SHA256SUMS" >"${current_sums}"
	if [[ "$(wc -l <"${current_sums}" | tr -d ' ')" != "$#" ]]; then
		echo "Error: SHA256SUMS contains $(wc -l <"${current_sums}" | tr -d ' ')/$# expected entries." >&2
		rm -f "${current_sums}"
		return 1
	fi

	if command -v sha256sum >/dev/null 2>&1; then
		if (
			cd "${dir}"
			sha256sum -c "${current_sums}"
		); then
			status=0
		else
			status=$?
		fi
	elif command -v shasum >/dev/null 2>&1; then
		if (
			cd "${dir}"
			shasum -a 256 -c "${current_sums}"
		); then
			status=0
		else
			status=$?
		fi
	else
		echo "Error: sha256sum or shasum is required." >&2
		rm -f "${current_sums}"
		return 1
	fi

	rm -f "${current_sums}"
	return "${status}"
}

needs_download_all() {
	local expected_fingerprint="$1"
	local fp

	for name in "${ALL_BIN_NAMES[@]}"; do
		fp="$(read_binary_fingerprint "${BIN_DIR}/${name}" || true)"
		[[ "${fp}" == "${expected_fingerprint}" ]] || return 0
	done

	[[ -f "${BIN_DIR}/SHA256SUMS" ]] || return 0

	verify_checksums "${BIN_DIR}" "${ALL_BIN_NAMES[@]}" >/dev/null 2>&1 || return 0

	return 1
}

cd "${REPO_ROOT}"

branch="$(git branch --show-current)"
if [[ -z "${branch}" ]]; then
  echo "Error: detached HEAD is not supported. Check out a branch before pulling." >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Error: working tree is not clean. Commit or stash changes before pulling." >&2
  exit 1
fi

remote_url="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "${remote_url}" ]]; then
  echo "Error: git remote 'origin' is missing." >&2
  exit 1
fi

# Sync entry blocks BEFORE pulling, so in-memory script functions
# are used before git pull can modify the script file on disk.
sync_existing_entry_blocks

echo "Pulling ${branch} from origin..."
git pull --ff-only origin "${branch}"

fingerprint="$("${REPO_ROOT}/tooling/scripts/tooling_fingerprint.sh")"
short_fingerprint="${fingerprint:0:12}"
tag="specflow-tooling-${short_fingerprint}"
ALL_BIN_NAMES=()
for suffix in "linux-amd64" "linux-arm64" "darwin-amd64" "darwin-arm64" "windows-amd64.exe" "windows-arm64.exe"; do
	ALL_BIN_NAMES+=("specflowctl-${suffix}" "specflow-reader-${suffix}")
done

if ! needs_download_all "${fingerprint}"; then
	echo "Local binaries already match ${tag}."
	exit 0
fi

if ! git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; then
	echo "Error: release tag does not exist on origin: ${tag}" >&2
	echo "Run push_with_release.sh on main first, then run this pull script again." >&2
	exit 1
fi

download_dir="$(mktemp -d)"
base="https://github.com/Bingordinary/SpecFlow/releases/download/${tag}"

echo "Downloading ${tag} binaries for all platforms..."
for name in "${ALL_BIN_NAMES[@]}"; do
	curl -fL -o "${download_dir}/${name}" "${base}/${name}"
done
curl -fL -o "${download_dir}/SHA256SUMS" "${base}/SHA256SUMS"

verify_checksums "${download_dir}" "${ALL_BIN_NAMES[@]}"

mkdir -p "${BIN_DIR}"
for name in "${ALL_BIN_NAMES[@]}"; do
	mv "${download_dir}/${name}" "${BIN_DIR}/${name}"
done
mv "${download_dir}/SHA256SUMS" "${BIN_DIR}/SHA256SUMS"
for name in "${ALL_BIN_NAMES[@]}"; do
	chmod +x "${BIN_DIR}/${name}"
done

echo "Installed all platform binaries and SHA256SUMS from ${tag}."
