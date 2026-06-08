# Unit Stable Verify

`unit_stable_verify:{unit}` checks whether the current implementation still conforms to the stable-layer truth.

## Input

- `docs/specs/_status.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- Stable-layer appendices and rule files referenced by the unit
- The unit's entry in `docs/specs/repository_mapping.md`
- Current implementation and test files
- Existing `_stable_verify_result/unit/{unit}.md` (if an update is needed)

## What This Step Does

Check current implementation consistency with stable-layer truth.
Output should be `aligned` (consistent), `controlled_repair_required` (repair needed), or `controlled_change_required` (change needed).

## Note

- This step requires independent review, not self-approval
- Stable verification does not create candidate truth itself. If a change is needed, the result triggers a subsequent `unit_fork`
- For `aligned`, every acceptance item must have `pass` evidence

## Not Allowed

- Modify stable-layer or candidate-layer truth
- Modify implementation files
- Modify lifecycle state
- Modify rule truth

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `aligned` | Implementation matches stable truth | `unit_fork` |
| `controlled_repair_required` | Repair needed | `unit_fork` with repair intent |
| `controlled_change_required` | Change needed | `unit_fork` with change intent |
| `small_repair_required` | Small repair needed, no behavior truth change | `unit_stable_verify` (re-verify) |
| `truth_rejudge_required` | Stable-layer truth needs re-evaluation | `unit_stable_verify` (re-verify) |
| `evidence_incomplete` | Evidence insufficient | Supplement evidence and re-verify |

Close through `command close`.
