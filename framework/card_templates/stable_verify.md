<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit}

## STATUS
- Stage: unit_stable_verify | Next: unit_fork (after verification)
- Layer: stable

## GUIDANCE
Check whether the current implementation still conforms to the stable-layer truth.

> **Pre-check:** Read `docs/specs/_status.md` to confirm this unit has Active=stable and Next Command != unit_promote.

**Execution steps:**
1. Read `docs/specs/units/stable/s_unit_{unit}.md` + appendices.
2. Read implementation and test files to assess alignment.
3. For each acceptance criterion in the stable spec, check whether the implementation satisfies it.
4. Write the result to `docs/specs/_stable_verify_result/unit/{unit}.md` per `framework/process_snapshot_contract.md`.
5. Independent review is required. Follow `framework/operations/entry_routing.md` Independent Review Stop, then `framework/core/independent_evaluation.md` for reviewer pack selection.

**Possible outcomes (6):**
| Outcome | Meaning | Next step |
|---|---|---|
| aligned | Implementation matches spec | No action needed |
| controlled_repair_required | Minor non-behavioral fix needed | unit_fork:{unit} with repair intent |
| controlled_change_required | Behavioral change needed | unit_fork:{unit} with change intent |
| small_repair_required | Tiny fix, no fork needed | Fix code directly |
| truth_rejudge_required | Stable truth is wrong | Must re-examine stable truth |
| evidence_incomplete | Cannot determine | Gather more evidence, re-run |

**Close:** `specflowctl command close --command unit_stable_verify --object-type unit --object {unit} --outcome <outcome>`

> If the current request does not involve spec changes: this stage is itself a verification operation — proceed as above.

## WRITES (owned by this unit)
- docs/specs/_stable_verify_result/unit/{unit}.md
- {impl_paths}/** (non-behavioral repair only when small_repair_required)

## READS (read-only context)
- docs/specs/_status.md
- docs/specs/units/stable/s_unit_{unit}.md + appendices
- docs/specs/repository_mapping.md
- Implementation and test files
- docs/specs/_stable_verify_result/unit/{unit}.md (if update needed)
- framework/lifecycle/unit_stable_verify.md
- framework/process_snapshot_contract.md

## BLOCKED
- Modifying stable-layer or candidate-layer truth
- Modifying lifecycle state
- Modifying rule truth
- Modifying implementation files (except small_repair_required)

## CLOSE
specflowctl command close --command unit_stable_verify --object-type unit --object {unit} --outcome <outcome>

## Next Steps
aligned → no action needed
controlled_repair_required / controlled_change_required → unit_fork:{unit}
Re-run: specflowctl context card
