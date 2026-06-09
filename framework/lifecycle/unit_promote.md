# Unit Promote

`unit_promote:{unit}` promotes verified candidate truth to stable truth.

## Input

- `docs/specs/_status.md`
- `docs/specs/_verify_result/unit/{unit}.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- Current unit's candidate-layer appendix files

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_promote`.
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/_verify_result/unit/{unit}.md` — confirm verification passed with `ready_to_promote`.
4. [ ] Confirm both candidate-layer and stable-layer Spec files exist.
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

1. Write candidate truth (main Spec + appendices) as stable-layer truth
2. Update lifecycle state and refs
3. Clean up candidate-layer evidence files

This is a mechanical operation that does not involve new behavior judgment.
`unit_promote` does not need a new independent review — it consumes the evidence already verified by `unit_verify`.

## Not Allowed

- Introduce behavior, acceptance, ownership, or rule meaning outside the verified scope
- Modify implementation files
- Manually modify lifecycle state
- Delete candidate-layer evidence before `command close --apply` completes

## How to End

`promoted` → run `command close --command unit_promote --outcome promoted --apply`.
After success: `Active Layer=stable`, `Next Command=unit_fork`, candidate-layer evidence is cleaned up.
