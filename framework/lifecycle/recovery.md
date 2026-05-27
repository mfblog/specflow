# Recovery

Recovery resets unsafe process evidence to the smallest legal restart point.

It applies to unit lifecycle work, rule-governance work that already mutated files, and impact sync outcomes that invalidate downstream evidence.

## Unit Fallback Targets

| Failure Layer | Reason Codes | Deletes | Next Command |
|---|---|---|---|
| `truth_layer` | `truth_drift`, `binding_drift`, `baseline_drift`, `rule_drift`, `truth_incomplete` | check checklist, check result, plan, verify result | `unit_check` |
| `gate_layer` | `gate_missing` | check checklist, check result | `unit_check` |
| `plan_layer` | `plan_drift` | plan, verify result | `unit_plan` |
| `implementation_layer` | `implementation_deviation` | verify result | `unit_impl` |
| `evidence_layer` | `evidence_incomplete`, `stable_verify_invalid` | verify result or stable-verify result | `unit_verify` or `unit_stable_verify` |

Only reason codes in this table are valid for fallback cleanup.
Do not introduce alternate names for the same invalidated layer.
Use the earliest layer that is invalidated by current repository truth.
Do not delete upstream process files that still validate and are still supported by current truth.

## Candidate Recovery

When candidate truth changes, bound rule references change, repository mapping changes the unit boundary, or a global rule changes the candidate's constraints:

1. delete downstream evidence that was derived from the prior truth.
2. set the candidate unit's next command to the earliest required command from the fallback target table.
3. keep still-valid upstream evidence only when deterministic validation proves it still matches current truth.
4. rerun impact sync when the change may affect other units.

If a candidate main Spec changes after `unit_check`, the default fallback is `unit_check`.
Use a later fallback only when the change is proven not to affect the earlier gate.

## Stable Unit Recovery

Stable truth is not edited in place through recovery.
If a stable unit needs a behavior, boundary, acceptance, or rule-binding change, start a candidate through `unit_fork:{unit}` and use `candidate_intent=change` or `candidate_intent=repair` as appropriate.

Stable implementation drift routes to `unit_stable_verify:{unit}`.
Stable truth drift that requires a new version routes through `unit_fork:{unit}`.

## Promotion Recovery

If promotion has not mutated stable truth yet, recover by resetting the candidate to the earliest invalidated command.

If promotion has already mutated stable truth but closure is incomplete:

1. do not silently keep partial promotion state.
2. restore the unit to a deterministic candidate state when stable truth cannot be proven complete.
3. delete process evidence that references the incomplete promotion result.
4. set the next command to `unit_check` unless only verification evidence is invalid.
5. rerun impact sync for any stable dependency or rule consumer that could observe the promotion attempt.

## Rule-Governance Recovery

Rule-governance flows must capture a recovery baseline before the first file mutation.
The baseline must include the files that would need to be restored or revalidated if the flow cannot close.

If repository truth becomes insufficient before mutation, stop and route through `framework/governance/rules/rule_escape.md`.

If mutation already happened and the rule flow cannot safely close:

1. stop further rule mutation.
2. restore or complete the smallest set of rule files needed to make repository truth deterministic.
3. run `rule_sync` when any affected unit or rule consumer may have downstream drift.
4. apply candidate or stable unit recovery for every affected unit.
5. return to `framework/operations/entry_routing.md` only after the repository is no longer left in a partial rule-governance state.

## Success Cleanup

After a command or rule flow closes successfully:

1. remove process files that the closing command explicitly supersedes.
2. keep evidence that remains current and is still required by the next command.
3. never keep stale downstream evidence as a historical shortcut.
4. ensure `_status.md` names the next legal command for every affected unit.

## Removed Scenario Lifecycle

Requests that use `scenario_*`, `scenario_advance:{id}`, or `object-type=scenario` are not recoverable lifecycle work.
Stop and report that scenario lifecycle support has been removed.
