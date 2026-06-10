# Unit Stable Verify

`unit_stable_verify:{unit}` checks whether the current implementation still conforms to the stable-layer truth.

## Input

- `docs/specs/_status.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- Stable-layer appendices and rule files referenced by the unit
- The unit's entry in `docs/specs/repository_mapping.md`
- Current implementation and test files
- Existing `_stable_verify_result/unit/{unit}.md` (if an update is needed)
- `framework/process_snapshot_contract.md` (for stable verify result file format and validation rules)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Active Layer` is `stable` and `Next Command` is not `unit_promote` (see `framework/core/status.md` "Valid Next Commands" for `unit_stable_verify` check-command semantics).
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/units/stable/s_unit_{unit}.md` — confirm stable-layer truth exists and is the current accepted truth.
4. [ ] Confirm current implementation and test files are accessible.
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

## What This Step Does

Check current implementation consistency with stable-layer truth.
Output should be `aligned` (consistent), `controlled_repair_required` (repair needed), `controlled_change_required` (change needed), `small_repair_required` (small non-behavior repair), `truth_rejudge_required` (stable-layer truth re-evaluation needed), or `evidence_incomplete` (insufficient evidence). See the How to End table for advancing vs non-advancing outcomes.

## Note

- This step requires independent review, not self-approval (see `framework/operations/entry_routing.md` "Independent Review Stop" for report format and `framework/core/independent_evaluation.md` for reviewer pack selection)
- Stable verification does not create candidate truth itself. If a change is needed, the result triggers a subsequent `unit_fork`
- For `aligned`, every acceptance item must have `pass` evidence

## Not Allowed

- Modify stable-layer or candidate-layer truth
- Modify lifecycle state
- Modify rule truth
- Modify implementation files — **Exception**: when the current outcome is `small_repair_required`, non-behavior implementation repairs (comments, formatting, documentation strings, naming) are permitted before re-verify

## How to End

| Result | Meaning | Next Step | Command Close Writeback |
|--------|---------|-----------|------------------------|
| `aligned` | Implementation matches stable truth | `unit_fork` | command close keeps `Next Command=unit_fork`. Write `_stable_verify_result` at `docs/specs/_stable_verify_result/unit/{unit}.md`. |
| `controlled_repair_required` | Repair needed | Write `_stable_verify_result` at `docs/specs/_stable_verify_result/unit/{unit}.md`. `unit_fork` with repair intent | command close keeps `Next Command=unit_fork`. |
| `controlled_change_required` | Change needed | Write `_stable_verify_result` at `docs/specs/_stable_verify_result/unit/{unit}.md`. `unit_fork` with change intent | command close keeps `Next Command=unit_fork`. |
| `small_repair_required` | Small repair needed, no behavior truth change | Perform the non-behavior repair on implementation files. Then `unit_stable_verify` (re-verify) | command close keeps `Next Command=unit_fork`. `unit_stable_verify` may be re-entered per `framework/core/status.md` "Valid Next Commands" allows semantics. |
| `truth_rejudge_required` | Stable-layer truth needs re-evaluation | If truth text is valid: supplement interpretation evidence and re-verify. If truth text must change: `unit_fork` with repair intent | command close keeps `Next Command=unit_fork` (no change until durable action is taken). |
| `evidence_incomplete` | Evidence insufficient | Supplement evidence and re-verify | command close keeps `Next Command=unit_fork`. |

Close through `command close`.
Tooling invocation: `specflowctl command close --command unit_stable_verify --object-type unit --object <unit> --outcome <outcome> [--notes <notes>] [--apply]`
