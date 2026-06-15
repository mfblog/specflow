# Unit Check

`unit_check:{unit}` is a pre-verify quality gate that checks whether candidate truth is sufficiently clear and complete. It does not itself advance lifecycle state — however, a `pass` outcome's `command close` sets `Next Command` to `unit_verify` with `Notes=pending_impl` (supplied by the caller alongside the close). This is a side effect of the close operation, not a progression behavior of `unit_check` as a check step.

## Input

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- `framework/process_snapshot_contract.md` (for check result file format and validation rules)
- `framework/spec_writing_guide.md` (for unit Spec format and source field format)
- `framework/candidate_intent.md` (for candidate_intent field rules, source_basis consistency, and repair candidate requirements)
- `docs/specs/_check_result/unit/{unit}.md` — present in re-validation flow; absent in standard flow (first check)
- `docs/specs/_check_work/unit/{unit}.md` — optional command-local checklist file for progress tracking (see `framework/process_snapshot_contract.md` Section 10)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm one of:
        - Standard entry: the target unit's `Next Command` is `unit_check`.
        - Re-validation: the target unit's `Next Command` is `unit_verify`, `Notes` contains
          `pending_impl`, and the candidate spec was modified after the last `unit_check` pass
          (re-validation during the implementation phase).
           To detect spec modification, compare the current spec fingerprint against `truth_fingerprint`
           stored in `docs/specs/_check_result/unit/{unit}.md` (see `framework/process_snapshot_contract.md` Section 6 for fingerprint calculation).
           If `docs/specs/_check_result/unit/{unit}.md` is absent during re-validation (e.g., the file was not yet created for this implementation session), proceed with re-validation — the absence itself indicates the spec has not been validated for the current implementation session.
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — confirm it exists and has acceptance items.
4. [ ] Confirm candidate-layer appendix files exist (if required).
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

## What This Step Does

Create or update the `_check_work` checklist at `docs/specs/_check_work/unit/{unit}.md` for progress tracking. Then check the following questions. All must pass for a `pass` result:

1. Is the unit's goal and responsibility scope clear?
2. Are dependencies, rule bindings, and ownership boundaries explicit?
3. Are the main flow, data, protocol, states, and error paths complete enough for verification?
4. Can verification proceed without guessing behavior, boundaries, or acceptance?
5. Do all acceptance items have the correct format (`verification_type`, `evidence_requirements`, `affects`)?
6. If `candidate_intent: change` + `source_basis: replacement`, is there at least one `verification_type: inspectable` item with `evidence_requirements` including `old_code_deleted` and `no_remaining_refs`?
7. Are all `affects` scopes correct (must not be empty without reason)?
8. If `evidence_appendix_ref` is not `none`, does the referenced appendix file exist with valid frontmatter and correct `unit`/`layer` values?
9. If `source_basis` is `existing_implementation` or `mixed`: (a) `evidence_appendix_ref` must be present and not `none`; (b) the referenced evidence appendix file must exist with valid frontmatter and correct unit/layer values. (This closes the `source_basis`–to–evidence-appendix consistency check: a claim of existing-implementation source status must be backed by a real, valid evidence appendix.)
10. If `candidate_intent: repair`, does the Spec include a `Repair Scope` section with the required sub-fields (acceptance item IDs being restored, observed deviations, expected implementation-side changes, verification evidence required)?
11. If `candidate_intent: repair`, is `repair_basis` present and correctly formatted (`s_unit_{unit}@<version>`), and are `source_basis=new_design` and `evidence_appendix_ref=none`?
12. If `candidate_intent: change`, is `repair_basis` absent (not allowed for change candidates)?
13. Does the repair candidate preserve stable behavior truth? If it modifies protocol, fields, ownership, or state machine semantics, it must require `fix_required` and recommend switching to `change`.

## Not Allowed

- Modify implementation files
- Modify stable-layer truth
- Modify lifecycle state
- Modify rule truth

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `pass` | Spec meets conditions | Write `_check_result` at `docs/specs/_check_result/unit/{unit}.md` with required fields: `unit_snapshot` (current candidate spec fingerprint, the `acceptance_behavior_fingerprint`, and the `rule_snapshot` from `_check_work` — see `framework/process_snapshot_contract.md` Section 6 for fingerprint calculation), `check_work_snapshot` (the checkpoint checklist state), and `truth_fingerprint` (the candidate spec fingerprint at check time). Write or update `_check_work` at `docs/specs/_check_work/unit/{unit}.md` with status tracking. Requires independent review before entering `unit_verify`: use the `unit_check_pass` reviewer pack from `framework/core/independent_evaluation.md`. When reporting a review stop, document: (1) the generated evaluation request file path, (2) the trigger instruction from `specflowctl evaluation request`, (3) that the reviewer must not modify repository files, (4) that execution resumes after the reviewer returns `pass`, `blocked`, or `needs_human_decision`. `command close` sets `Next Command=unit_verify` (caller supplies `Notes=pending_impl`). After command close, run `unit_impl:{unit}` to enter the implementation phase before proceeding to `unit_verify:{unit}`. |
| `checkpoint` | Progress saved, review not yet complete | Resume `unit_check:{unit}`. command close keeps `Next Command=unit_check`. |
| `fix_required` | Spec needs repair | Fix the candidate Spec and re-check. command close keeps `Next Command=unit_check`. |
| `blocked` | Missing critical input | Ask the user to resolve the missing input. command close keeps `Next Command=unit_check`. |

Tooling invocation: `specflowctl command close --command unit_check --object-type unit --object <unit> --outcome <outcome> [--notes <notes>]`
