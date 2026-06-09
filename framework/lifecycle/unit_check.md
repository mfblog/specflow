# Unit Check

`unit_check:{unit}` is a pre-verify quality gate that checks whether candidate truth is sufficiently clear and complete. It does not itself advance lifecycle state — however, a `pass` outcome's `command close` sets `Next Command` to `unit_verify` and `Notes` to `pending_impl`. This is a side effect of the close operation, not a progression behavior of `unit_check` as a check step.

## Input

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_check`.
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — confirm it exists and has acceptance items.
4. [ ] Confirm candidate-layer appendix files exist (if required).
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

Check the following 7 questions. All must pass for a `pass` result:

1. Is the unit's goal and responsibility scope clear?
2. Are dependencies, rule bindings, and ownership boundaries explicit?
3. Are the main flow, data, protocol, states, and error paths complete enough for verification?
4. Can verification proceed without guessing behavior, boundaries, or acceptance?
5. Do all acceptance items have the correct format (`verification_type`, `evidence_requirements`, `affects`)?
6. If `candidate_intent: change` + `source_basis: replacement`, is there at least one `verification_type: inspectable` item with `evidence_requirements` including `old_code_deleted` and `no_remaining_refs`?
7. Are all `affects` scopes correct (must not be empty without reason)?

## Not Allowed

- Modify implementation files
- Modify stable-layer truth
- Modify lifecycle state
- Modify rule truth

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `pass` | Spec meets conditions | Write `_check_result`, requires independent review before entering `unit_verify`. `command close` sets `Next Command=unit_verify`, `Notes=pending_impl`. |
| `fix_required` | Spec needs repair | Fix the candidate Spec and re-check |
| `blocked` | Missing critical input | Ask the user |
