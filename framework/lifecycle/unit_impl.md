# Unit Implementation

`unit_impl` is the implementation phase between candidate truth validation and verification.

## Input

- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- `docs/specs/_check_result/unit/{unit}.md` (if present)

## What This Step Does

Implement the code according to the acceptance items in the candidate Spec.
If the Spec is found to be missing or incorrect during implementation, stop and ask the user.

## Not Allowed

- Modify Spec files (candidate or stable layer)
- Modify lifecycle state
- Implement behavior beyond the candidate Spec scope
- Modify rule truth or global rules

## How to End

After all acceptance items have been implemented, proceed to `unit_verify:{unit}`.
No special close command is needed.
