# Unit Implementation

`unit_impl` is the implementation phase between candidate truth validation and verification.

## Input

- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- `docs/specs/_check_result/unit/{unit}.md` (if present)

## Pre-Execution Self-Check (MANDATORY)

Before writing any code, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_impl`.
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — this is the truth you must implement against.
4. [ ] Read the candidate-layer acceptance items — confirm every item is clear enough to implement.
5. [ ] If `Next Command` is NOT `unit_impl`: STOP, do not write code, and report the current state.
6. [ ] If the candidate Spec is missing or incomplete: STOP and report to the user.

If all checks pass: proceed to "What This Step Does" below.

Implement the code according to the acceptance items in the candidate Spec.
If the Spec is found to be missing or incorrect during implementation, stop and ask the user.

## Not Allowed

- Modify Spec files (candidate or stable layer)
- Modify lifecycle state
- Implement behavior beyond the candidate Spec scope
- Modify rule truth or global rules

## How to End

After all acceptance items have been implemented, close the implementation phase through `command close --command unit_verify --outcome ready_to_verify --apply`.
After success: `Next Command=unit_verify`, proceed to `unit_verify:{unit}`.
