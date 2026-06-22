# Unit Entry Commands

This file covers three entry commands: `unit_init:{unit}`, `unit_new:{unit}`, and `unit_fork:{unit}`.

| Command | Purpose |
|---------|---------|
| `unit_init` | Record an existing capability directly as first stable truth (no candidate layer needed) |
| `unit_new` | Create the first candidate truth for a brand-new unit |
| `unit_fork` | Branch a candidate round from existing stable truth for changes or repairs |

## Input

> **Reading guidance:** Must Read files are the truth and process data this command evaluates. May Reference files hold the format and policy contracts — read them when a specific check question needs the exact rule text. Procedural instructions are inline in "What This Step Does" and "How to End" below.

### Must Read

- `docs/specs/_status.md` (if the target unit may already be registered)
- `docs/specs/repository_mapping.md` (if path ownership or registration must be confirmed)
- `docs/specs/units/stable/s_unit_{unit}.md` + stable-layer appendices (unit_fork only)
- `docs/specs/_stable_verify_result/unit/{unit}.md` (for candidate intent determination — `unit_fork` only)

### May Reference

- `framework/spec_writing_guide.md` (unit Spec format, source field format, and appendix format)
- `framework/candidate_intent.md` (candidate_intent determination rules)
- `framework/process_snapshot_contract.md` (process evidence file format if applicable)
- `framework/operations/entry_routing.md` (Onboarding Source Decision — `unit_new` uses this to determine `source_basis` and whether `unit_init` applies)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit is in the expected state for the entry command:
     - `unit_new`: confirm the target unit does NOT yet have a row in `docs/specs/_status.md`
     - `unit_init`: confirm the status table is NOT empty and the target unit is not yet registered
     - `unit_fork`: confirm the target unit has a stable row and `Next Command` is `unit_fork`
2. [ ] Read the required Input files listed above — confirm they exist and are readable.
3. [ ] For `unit_fork`: confirm stable-layer unit truth exists at `docs/specs/units/stable/s_unit_{unit}.md`.
4. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "Requirements Per Command" below.

## What This Step Does

### unit_init
The existing accepted capability must be explicit enough to write stable truth without choosing new behavior, acceptance, or ownership.
If the status table is empty: STOP, report that no units are registered — `unit_init` requires an existing capability to onboard.

### unit_new
Candidate truth must be explicit enough to write the first candidate Spec and its source fields.
Confirm the target unit does not yet have a row in `docs/specs/_status.md` — `unit_new` requires no existing unit registration.

### unit_fork
- Current stable truth is the baseline for the candidate round
- `candidate_intent` must be determined (`change` or `repair`)
- If a valid stable verify result exists:
  - `controlled_repair_required` → write `repair`
  - `controlled_change_required` → write `change`
  - `truth_text_change_required` → write `repair`
  - `aligned` → no specific intent required
- Every stable-layer appendix must have a corresponding same-named candidate-layer appendix, unless the stable appendix declares `status: exempt` in its frontmatter (see `framework/spec_writing_guide.md` §Appendix Files)
- Rewrite Markdown document references within the candidate main Spec body AND within every copied candidate appendix file from stable appendix paths (`s_unit_*`) to candidate appendix paths (`c_unit_*`), ensuring the candidate body and appendix files reference the correct candidate-layer paths. Additionally, rewrite the `layer` frontmatter field in each copied appendix file from `stable` to `candidate`, and update the `version` field if applicable.
- If the status table is empty: STOP, report that no units are registered, and suggest `unit_new` as the first step

## Not Allowed

- Modify implementation files
- Manually modify lifecycle state
- Modify rule truth or global rules
- Modify other units' truth
- Modify stable-layer truth during `unit_new` / `unit_fork`
- Modify candidate-layer truth during `unit_init`
- Introduce behavior, acceptance, ownership, or rules not yet decided in the Required Context (the set of accepted formal truth files, lifecycle state, and binding constraints — `_status.md`, stable unit truth, rules, and `repository_mapping.md` — that define the current decision boundary for the target unit)

## Allowed Writes

- `docs/specs/units/candidate/c_unit_{unit}.md` — candidate main Spec (unit_new, unit_fork)
- `docs/specs/units/stable/s_unit_{unit}.md` — stable main Spec (unit_init)
- `docs/specs/repository_mapping.md` — path ownership registration
- `docs/specs/units/candidate/appendix/c_unit_{unit}_*.md` — candidate appendix files (unit_fork: copied from stable with frontmatter rewrites; stable appendix files with `status: exempt` are skipped)

## Note

`unit_check` is the required follow-up quality gate for unit_new and unit_fork. After unit_new and unit_fork complete, Next Command is set to `unit_check`.

## How to End

| Command | Success Result | Write Target | Next Step | Command Close Writeback |
|---------|---------------|-------------|-----------|------------------------|
| `unit_init` | `stable_created` | Write stable unit Spec at `docs/specs/units/stable/s_unit_{unit}.md` per `framework/spec_writing_guide.md` format. Add or update `docs/specs/repository_mapping.md` with a row: `kind=unit`, `registration_state=landed`, `implementation_paths` per existing implementation, `spec_files` referencing the stable Spec. | `unit_fork` | command close sets `Next Command=unit_fork` |
| `unit_new` | `candidate_created` | Write candidate unit Spec at `docs/specs/units/candidate/c_unit_{unit}.md`. Write `source_basis` — one of `new_design`, `existing_implementation`, `mixed`, or `replacement` — determined by whether behavior is sourced from a new design or an existing implementation (see `framework/operations/entry_routing.md` Onboarding Source Decision for per-value mapping). Add or update `docs/specs/repository_mapping.md` with a row: `kind=unit`, `registration_state=planned` when `source_basis` is `new_design` (no implementation paths exist); `registration_state=landed` when `source_basis` is `existing_implementation`, `mixed`, or `replacement` (implementation paths exist at registration time). `spec_files` referencing the candidate Spec. `candidate_intent` is not required for `unit_new`. | `unit_check` | command close sets `Next Command=unit_check` |
| `unit_fork` | `candidate_created` | Write candidate unit Spec at `docs/specs/units/candidate/c_unit_{unit}.md` with stable baseline and candidate appendices. Set `source_basis` — one of `new_design`, `existing_implementation`, `mixed`, or `replacement` as determined by the Onboarding Source Decision table in `framework/operations/entry_routing.md`. For repair candidates specifically, `source_basis` must be `new_design`. Ensure the unit's `docs/specs/repository_mapping.md` row exists. If it doesn't, add it with `registration_state=planned`. Update `spec_files` to reference the candidate Spec. | `unit_check` | command close sets `Next Command=unit_check` |

Close through `command close`. Before closing, ensure all writes are complete and no unresolved rule-governance or ownership issues remain.
General tooling invocation: `specflowctl command close --command unit_init|unit_new|unit_fork --object-type unit --object {unit} --outcome {outcome}`
==ATOM_BEGIN:close_fallback==
### Manual Command Close (when `specflowctl` is unavailable)

When `specflowctl command close` is unavailable (tooling not installed, broken, or
inaccessible), read `framework/lifecycle/command_close_fallback.md` for the complete
manual command close procedure.
==ATOM_END:close_fallback==
