# Entry Index Registry

## Purpose

This file registers which files in the repository serve as entry index files.

An "entry index file" means a repository file that executors read directly and use for at least one of the following:

1. listing available commands
2. explaining how commands are matched
3. routing a user request into a governance flow

This file answers only two questions:

1. which entry index files are formally registered by default
2. how those registered entry files must stay aligned

This file does not define the template-source review scope of `spec_flow_review`.

---

## Registered Entry Index Files

- `AGENTS.md`
- `GEMINI.md`
- `CLAUDE.md`

Additional rules:

1. This registry defines the project-side registered entry files used for managed-block sync and manual consistency checks.
2. In `installed_project`, the registered paths resolve at the repository root.
3. In `source_repo`, the registered paths resolve under `<template-root>/` for template entry managed-block checks. Local source-repository `AGENTS.md`, `GEMINI.md`, and `CLAUDE.md` files are personal or host-owned files unless another rule registers them.
4. If a new project-side entry index file is added later, update this file first so sync and commit-time checks can include it.
5. Files with the same responsibility but not yet registered are not part of the project-side registered set.
6. `spec_flow_review` may read this file to verify project-side ownership and sync semantics.

---

## Managed Block Rule

All registered entry index files must contain exactly one managed block:

1. start marker: `<!-- SPECFLOW:BEGIN -->`
2. end marker: `<!-- SPECFLOW:END -->`

Content outside that managed block belongs to the host repository.

`specFlow` tooling may update only the managed block. Host-specific rules outside the managed block are not part of `specFlow` ownership.

## Consistency Rule

All registered entry index files must keep their managed blocks consistent.

Fixed rules:

1. Executors may edit any registered entry file directly. There is no permanent single source file requirement.
2. Host-owned content outside the managed block may differ across registered entry files.
3. If all managed blocks already match, no sync-source decision is needed for the current round.
4. If managed blocks differ and only one registered entry file was modified in the current round, that file is the default sync source.
5. If multiple registered entry files were modified but their managed blocks now match, treat them as already manually aligned.
6. If multiple registered entry files were modified and their managed blocks still differ, the task enters an explicit source-selection case:
   - do not guess the sync source from mtime, path order, or other environment metadata
   - explicitly choose which registered entry file is the source for this round before syncing
   - run `<tooling-root>/bin/specflowctl-<os>-<arch> entry sync --source <registered-entry-file>`
   - `<registered-entry-file>` must be one of the registered project-side entry files from this registry, for example `AGENTS.md`
7. Syncing is only responsible for re-aligning managed blocks across registered entry files. It does not narrow review scope or rewrite governance judgment rules.

This design has only two goals:

1. people using different tools can keep host-specific instructions where they need them
2. the default review scope and managed-block consistency both remain stable and predictable

---

## Manual Sync Trigger

Entry-file managed blocks must be synchronized before a task claims that registered entry edits are complete.

Rules:

1. If a round edits one registered entry managed block, run `<tooling-root>/bin/specflowctl-<os>-<arch> entry sync --source <registered-entry-file>` before closure.
2. If a round edits multiple registered entry managed blocks and those blocks already match, no sync command is required.
3. If a round edits multiple registered entry managed blocks and those blocks still differ, explicitly choose one registered entry file as the source and run `<tooling-root>/bin/specflowctl-<os>-<arch> entry sync --source <registered-entry-file>`.
4. Git staging remains a manual action outside specFlow governance.

---

## Non-Goals

This file does not:

1. define standard commands themselves
2. define the findings contract in place of `spec_flow_review.md`
3. automatically treat every draft file with a similar responsibility as a formal entry file
4. install or require a git hook
