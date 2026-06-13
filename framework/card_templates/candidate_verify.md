<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit}

## STATUS
- Stage: unit_verify | Next: unit_promote
- Layer: candidate | Intent: {intent}

## GUIDANCE
Verify the implementation against each acceptance criterion in the candidate spec.

> **Pre-check:** Read `docs/specs/_status.md`. Confirm this unit has Next=unit_verify.

**Execution steps:**
1. Read `docs/specs/units/candidate/c_unit_{unit}.md` + appendices for acceptance criteria.
2. Read `docs/specs/_check_result/unit/{unit}.md` for the check baseline.
3. Read implementation and test files.
4. For each acceptance criterion, produce checkable evidence (evidence matrix per `framework/process_snapshot_contract.md`).
5. Write the verification result to `docs/specs/_verify_result/unit/{unit}.md`.
6. Independent review is required before promote. Follow `framework/operations/entry_routing.md` Independent Review Stop and `framework/core/independent_evaluation.md` for reviewer pack selection.

**Possible outcomes (3):**
| Outcome | Meaning | Next step |
|---|---|---|
| ready_to_promote | All criteria met, verified | run unit_promote:{unit} |
| spec_issue | Spec is wrong or incomplete | fix spec, run unit_check:{unit} |
| impl_issue | Implementation doesn't match spec | fix code, re-run unit_verify:{unit} |

**Close:** `specflowctl command close --command unit_verify --object-type unit --object {unit} --outcome <outcome>`

> If the current request does not involve spec changes (pure implementation, refactoring, testing, performance optimization):
> You may modify implementation code directly. If you find the spec needs to change, stop and use unit_check first.

## WRITES (owned by this unit)
- docs/specs/_verify_result/unit/{unit}.md
- docs/specs/_check_work/unit/{unit}.md (if re-checking spec)
- {impl_paths}/**
- {test_paths}/**

## READS (read-only context)
- docs/specs/_status.md
- docs/specs/units/candidate/c_unit_{unit}.md + appendices
- docs/specs/_check_result/unit/{unit}.md
- docs/specs/repository_mapping.md
- framework/lifecycle/unit_verify.md
- framework/process_snapshot_contract.md
- framework/core/independent_evaluation.md
- {rule_refs_paths}
- {unit_refs_paths}

## BLOCKED
- _status.md (use command close)
- Candidate spec (must go back to unit_check to modify)
- Stable-layer truth
- Any rule files
- Other units' specs or status

## CLOSE
specflowctl command close --command unit_verify --object-type unit --object {unit} --outcome <outcome>

## Next Steps
ready_to_promote → run unit_promote:{unit}
spec_issue → run unit_check:{unit} (fix spec first)
impl_issue → fix code, re-run unit_verify:{unit}
Re-run: specflowctl context card
