# specFlow

`specFlow` is a spec-driven development governance framework.

It is not about "how to write a document." It is about "how humans and AI can move design, planning, implementation, verification, and promotion forward around the same source of truth."

## Layout

- `framework/`
  - Source files for the framework rules.
- `templates/root/`
  - Template files installed into the target repository root and fixed paths.
- `tooling/`
  - Initialization, upgrade, and diagnosis tools.

## Quick Start

After adding `specflow/` to your repository, go to the repository root and run:

```bash
./specflow/tooling/init.sh
```

Windows PowerShell:

```powershell
.\specflow\tooling\init.ps1
```

After initialization, the target repository will receive:

- `AGENTS.md`
- `GEMINI.md`
- `CLAUDE.md`
- `.githooks/pre-commit`
- `docs/specs/**`

The framework governance files remain inside:

- `specflow/framework/docs/agent_guidelines/**`

You can then use:

- `./specflow/tooling/doctor.sh`
- `./specflow/tooling/upgrade.sh`

Windows PowerShell:

- `.\specflow\tooling\doctor.ps1`
- `.\specflow\tooling\upgrade.ps1`

## Template Ownership

Files listed in `specflow/tooling/manifest.tsv` use two ownership modes:

- `framework`
  - owned by `specFlow`
  - `upgrade` may refresh these files from the framework templates
- `project`
  - bootstrapped into the host repository by `init`
  - once the file already exists in the host repository, `upgrade` must not overwrite it
  - if the file is missing, `upgrade` may install the missing file from the template

In plain words:

1. `init` lays down the initial project-side bootstrap files.
2. After that, existing project-owned files belong to the host repository, not to `specFlow`.
3. `upgrade` is allowed to update framework-owned files and managed blocks, but it must not rewrite existing project-owned files.

Exception:

1. `docs/specs/_check_result/README.md`
2. `docs/specs/_plans/README.md`
3. `docs/specs/_verify_result/README.md`

These three files live under the project root, but they define the framework's process-file schema and gate semantics. Therefore they are treated as framework-owned in `manifest.tsv`, and `upgrade` should refresh them from `specflow/templates/root/`.

## Entry File Ownership

`AGENTS.md`, `GEMINI.md`, and `CLAUDE.md` are host-owned files with a `specFlow` managed block.

- Host-specific instructions belong outside:
  - `<!-- SPECFLOW:BEGIN -->`
  - `<!-- SPECFLOW:END -->`
- `init`, `upgrade`, `doctor`, and `specflow/tooling/sync_entry_docs.sh` operate only on the managed block.
- If an existing entry file does not contain exactly one managed block, `specFlow` refuses to guess and reports the file for manual repair.
