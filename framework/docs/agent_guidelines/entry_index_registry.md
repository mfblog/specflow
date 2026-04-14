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

---

## Registered Entry Index Files

- `AGENTS.md`
- `GEMINI.md`
- `CLAUDE.md`

Additional rules:

1. When `spec_flow_review` reviews entry index files under its default scope, it must use the registry in this section instead of guessing from file responsibilities.
2. If a new entry index file is added later, update this file first, then include that file in the default review scope.
3. Files with the same responsibility but not yet registered are not part of the default entry set for `spec_flow_review`.

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
7. Syncing is only responsible for re-aligning managed blocks across registered entry files. It does not narrow review scope or rewrite governance judgment rules.

This design has only two goals:

1. people using different tools can keep host-specific instructions where they need them
2. the default review scope and managed-block consistency both remain stable and predictable

---

## Hook Trigger

The default time to sync entry files is before `git commit`.

Rules:

1. The tracked `pre-commit` hook lives at `.githooks/pre-commit`.
2. That hook calls `specflow/tooling/sync_entry_docs.sh` before `git commit`.
3. If the script succeeds, the synced managed blocks are re-added to the index and the commit continues.
4. If the script finds a case where multiple registered entry files were modified and their managed blocks still differ, so no source can be chosen automatically, it must block the commit and require an explicit source choice.
5. If the repository-level hook path is not enabled yet, run:
   - `git config core.hooksPath .githooks`

---

## Non-Goals

This file does not:

1. define standard commands themselves
2. define the findings contract in place of `spec_flow_review.md`
3. automatically treat every draft file with a similar responsibility as a formal entry file
