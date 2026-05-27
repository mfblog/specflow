# Unit Entry Commands Context Card

This Context Card covers `unit_init:{unit}`, `unit_new:{unit}`, and `unit_fork:{unit}`.

`unit_init` captures an existing accepted capability as first stable truth.
`unit_new` creates the first candidate truth for a brand-new unit.
`unit_fork` creates a candidate round from existing stable truth.

## Required Context

Read only:

1. `framework/core/context_card.md`
2. `framework/core/lifecycle_authority.md`
3. `framework/core/object_model.md`
4. `framework/onboarding_decision_policy.md`
5. `framework/candidate_intent_policy.md`
6. `docs/specs/_status.md` when the target unit may already be registered.
7. `docs/specs/repository_mapping.md` when unit registration, path ownership, implementation path registration, or support-surface ownership must be proven before writeback.
8. `docs/specs/units/stable/s_unit_{unit}.md` and explicitly referenced stable appendices when running `unit_fork:{unit}`.
9. `framework/candidate_intents/change.md` or `framework/candidate_intents/repair.md` when `unit_fork:{unit}` creates a candidate with that intent.
10. the current valid `docs/specs/_stable_verify_result/unit/{unit}.md` when `_status.md` is stable, `Next Command` is `unit_fork`, and that stable verify result exists.

For `unit_init`, accepted first-stable behavior must already be explicit enough to write stable truth without choosing behavior, acceptance, ownership, or rule meaning during writeback.
For `unit_new`, selected candidate truth must be explicit enough to write the first candidate and its source fields.
For `unit_fork`, current stable truth is the baseline for the candidate round.
If a current valid stable verify result records `decision: controlled_repair_required`, `unit_fork` must write `candidate_intent=repair`.
If a current valid stable verify result records `decision: controlled_change_required`, `unit_fork` must write `candidate_intent=change`.
If a current valid stable verify result records `decision: aligned`, it does not force a candidate intent beyond the normal `unit_fork` rules.

## Allowed Writes

Allowed writes are:

1. `unit_init:{unit}` may create `docs/specs/units/stable/s_unit_{unit}.md`, explicitly referenced stable appendices, and required `docs/specs/repository_mapping.md` entries.
2. `unit_new:{unit}` may create `docs/specs/units/candidate/c_unit_{unit}.md`, required candidate appendices, and required `docs/specs/repository_mapping.md` entries.
3. `unit_fork:{unit}` may create or replace `docs/specs/units/candidate/c_unit_{unit}.md`, required candidate-only metadata, required `Repair Scope` content for repair candidates, required candidate appendices, and deterministic cleanup of obsolete candidate process files after close.
4. Lifecycle status may change only through successful `command close`.

Candidate units must declare `candidate_intent` when forked from stable truth.
Candidate units must declare `source_basis` and `evidence_appendix_ref` according to `framework/onboarding_decision_policy.md`.
Rule binding changes must be reconciled through `framework/governance/rule_system.md` and `framework/governance/impact_sync.md`.

## Forbidden Writes

Do not write:

1. implementation files.
2. lifecycle status by hand.
3. rule truth, global rules, or rule bindings unless rule governance explicitly allows the write.
4. unrelated unit truth or appendices.
5. process evidence for check, plan, verify, or stable-verify gates.
6. stable truth during `unit_new:{unit}` or `unit_fork:{unit}`.
7. candidate truth during `unit_init:{unit}`.
8. behavior, acceptance, ownership, dependency, or rule meaning that is not already decided by the required context.

Do not skip `unit_check`.
Entry commands create initial truth surfaces; `unit_check` is still the truth-closure gate before planning.

## On-Demand Expansions

Enter only when the trigger appears:

1. `framework/operations/entry_routing.md` when the request is not an exact `unit_init:{unit}`, `unit_new:{unit}`, or `unit_fork:{unit}` command, or the target object is unclear.
2. `framework/governance/rule_system.md` when shared rule, global rule, reusable mechanism, exception, or rule binding truth must change; use `framework/governance/rules/rule_escape.md` when current truth is insufficient to choose or finish the rule flow safely.
3. `framework/core/repository_mapping.md` when repository mapping structure or ownership semantics are unclear.
4. `framework/lifecycle/recovery.md` when an attempted entry write already happened and closure is no longer safe.
5. `framework/operations/migration.md` when existing files use an older shape that blocks the required writeback.

## Independent Evaluation

`unit_init`, `unit_new`, and `unit_fork` do not require an independent reviewer receipt.

These commands create entry truth and lifecycle position.
They do not approve readiness for planning.
The next candidate gate is `unit_check`, which requires independent evaluation for an advancing `pass`.

## Close Requirements

Outcomes:

| Command | Outcome | Status Result |
|---|---|---|
| `unit_init` | `stable_created` | Stable truth exists; next command is `unit_fork` |
| `unit_new` | `candidate_created` | Candidate truth exists; next command is `unit_check` |
| `unit_fork` | `candidate_created` | Candidate truth exists, obsolete candidate process files are cleaned, and next command is `unit_check` |

Do not close until every write named by the outcome is complete and no required rule-governance, repository-mapping, or onboarding-source decision remains unresolved.
Close through `command close` according to `framework/core/lifecycle_authority.md`.
After `unit_new` or `unit_fork`, run `unit_check:{unit}` before planning or implementation.
