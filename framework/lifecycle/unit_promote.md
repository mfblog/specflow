# Unit Promote

`unit_promote:{unit}` promotes verified candidate truth to stable truth.

## Input

- `docs/specs/_status.md`
- `docs/specs/_verify_result/unit/{unit}.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- Current unit's candidate-layer appendix files

## What This Step Does

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
