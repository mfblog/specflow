# Unit Verify

`unit_verify:{unit}` verifies whether the implementation satisfies each acceptance item in the candidate truth.

## Input

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- The unit's implementation and test files
- `docs/specs/_check_result/unit/{unit}.md` (if present, for reference but not required)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_verify`.
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — confirm candidate truth and acceptance items are available.
4. [ ] Confirm the unit's implementation and test files exist and are accessible.
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

1. **Functional verification**: Verify each acceptance item is satisfied with inspectable evidence
2. **Scope verification**: Verify the `affects` declarations (files, appendices, rules, dependencies) are correctly implemented
3. **Retirement verification** (replacement scenario): Verify old code paths are fully removed with no remaining references
4. **Code quality check**: No dead code, no over-engineering, reasonable change volume

## Verification Evidence Requirements

- Every executable acceptance item must have a corresponding entry in `acceptance_item_evidence_matrix`
- Changes to primary protocols, APIs, UI, or generated artifacts must be verified against real output (screenshots, API return values, CLI output, etc.), not just "tests pass"
- Verification must not automatically delete code or infer business compatibility safety

## Note

- This step requires independent review — **self-approval is not allowed**. An independent reviewer must give `pass` for `ready_to_promote`
- If implementation issues are found during verification, they may be fixed and re-verified
- If the candidate Spec itself is problematic, return to `unit_check` to fix the Spec

## Not Allowed

- Modify candidate or stable-layer truth
- Modify lifecycle state
- Modify rule truth

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `ready_to_promote` | Verification passed, review passed | Write `_verify_result`, proceed to `unit_promote` |
| `spec_issue` | Candidate Spec needs repair | Return to `unit_check:{unit}`, fix the Spec, and re-check |
| `impl_issue` | Implementation needs repair | Fix code and rerun `unit_verify:{unit}` |
