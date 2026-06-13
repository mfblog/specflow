# Unit Verify

`unit_verify:{unit}` verifies whether the implementation satisfies each acceptance item in the candidate truth.

## Input

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- The unit's implementation and test files
- `docs/specs/_check_result/unit/{unit}.md` — present in standard flow (unit_check → unit_impl → unit_verify); may be absent or stale in re-validation flow (unit_check re-validation path)
- `framework/process_snapshot_contract.md` (for verify result file format and validation rules)
- `framework/spec_writing_guide.md` (for unit Spec format and appendix format)
- `docs/specs/repository_mapping.md` (for implementation file ownership discovery)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_verify`.
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — confirm candidate truth and acceptance items are available.
4. [ ] Confirm the unit's implementation and test files exist and are accessible.
5. [ ] Compare the current candidate spec fingerprint against `truth_fingerprint` stored in `docs/specs/_check_result/unit/{unit}.md` (see `framework/process_snapshot_contract.md` Section 6 for fingerprint calculation). If the fingerprint changed after the last `unit_check` pass, STOP: the spec was modified without re-validation. Route through `unit_check:{unit}` for re-validation first. If `_check_result/unit/{unit}.md` does not exist (absent in re-validation flow), treat as spec-modified — the candidate truth has never passed check in this session; route through `unit_check:{unit}` first.
6. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

## What This Step Does

1. **Functional verification**: Verify each acceptance item is satisfied with inspectable evidence
2. **Scope verification**: Verify the `affects` declarations (files, appendices, rules, dependencies) are correctly implemented
3. **Retirement verification** (replacement scenario): Verify old code paths are fully removed with no remaining references
4. **Code quality check**: No dead code, no over-engineering, reasonable change volume

## Verification Evidence Requirements

- Every executable acceptance item must have a corresponding entry in `acceptance_item_evidence_matrix`
- Changes to primary protocols, APIs, UI, or generated artifacts must be verified against real output (screenshots, API return values, CLI output, etc.), not just "tests pass"
- Verification must not automatically delete code or infer business compatibility safety

## Note

- This step requires independent review — **self-approval is not allowed**. An independent reviewer must give `pass` for `ready_to_promote`. See `framework/operations/entry_routing.md` "Independent Review Stop" for report format and `framework/core/independent_evaluation.md` for reviewer pack selection.
- If implementation issues are found during verification, they may be fixed and re-verified
- If the candidate Spec itself is problematic, return to `unit_check` to fix the Spec

## Not Allowed

- Modify candidate or stable-layer truth
- Modify lifecycle state
- Modify rule truth

## How to End

| Result | Meaning | Next Step | Command Close Writeback |
|--------|---------|-----------|------------------------|
| `ready_to_promote` | Verification passed, review passed | Write `_verify_result` at `docs/specs/_verify_result/unit/{unit}.md`. Proceed to `unit_promote:{unit}` | command close sets `Next Command=unit_promote`. `unit_promote:{unit}` requires explicit user decision. |
| `truth_fallback` | Candidate truth has drifted from stable baseline | Return to `unit_check:{unit}` with truth-layer fallback. See `framework/lifecycle/recovery.md` truth_layer for process evidence cleanup. | command close sets `Next Command=unit_check`. Apply truth repair first. |
| `spec_issue` | Candidate Spec needs repair | Return to `unit_check:{unit}`, fix the Spec, and re-check | command close sets `Next Command=unit_check`. |
| `evidence_incomplete` | Evidence insufficient for verification | Supplement evidence (`acceptance_item_evidence_matrix`) and rerun `unit_verify:{unit}` | command close keeps `Next Command=unit_verify`. |
| `human_verify` | Human decision required before promotion | Ask the user and rerun `unit_verify:{unit}` after input | command close keeps `Next Command=unit_verify`. |
| `impl_issue` | Implementation needs repair | Fix code and rerun `unit_verify:{unit}` | command close keeps `Next Command=unit_verify`. |

For non-standard failures (process validation failure, tooling error, corrupted state), read `framework/lifecycle/recovery.md` and apply fallback cleanup per the failure layer.

Tooling invocation: `specflowctl command close --command unit_verify --object-type unit --object <unit> --outcome <outcome> [--notes <notes>] [--apply]`
