# Onboarding Decision Policy

This policy decides how unit onboarding handles the source of selected behavior truth.

It owns two decisions:

1. how a unit candidate records the source of selected behavior truth
2. whether a historical unit may land directly as the first stable Spec

No scenario onboarding path is supported.

## 1. Source Basis

Candidate unit frontmatter must record:

```yaml
source_basis: new_design|existing_implementation|mixed|replacement
evidence_appendix_ref: none
```

or an explicit appendix ref.

Allowed `source_basis` values:

1. `new_design`
   - the candidate does not use existing implementation as the source of selected behavior truth
   - `evidence_appendix_ref` must be `none`
2. `existing_implementation`
   - the candidate mainly captures behavior that already exists in implementation, tests, runtime behavior, or historical material
   - `evidence_appendix_ref` must point to a current candidate evidence appendix
3. `mixed`
   - the candidate combines retained existing behavior with new or changed design
   - `evidence_appendix_ref` must point to a current candidate evidence appendix
4. `replacement`
   - existing implementation may exist, but this candidate does not use it as the source of selected behavior truth
   - `evidence_appendix_ref` must be `none`

Rules:

1. `source_basis` never uses `repair`.
2. Repair is represented by `candidate_intent=repair`.
3. A repair candidate records `source_basis=new_design` because the selected behavior comes from stable formal truth, not from current implementation.
4. `unit_check` must reject unsupported, missing, or internally inconsistent source fields.

## 2. Candidate Intent

When a candidate is forked from a stable unit, the candidate must record `candidate_intent` according to `candidate_intent_policy.md`.

`candidate_intent` and `source_basis` answer different questions:

1. `candidate_intent` says why the candidate round exists.
2. `source_basis` says where the selected behavior truth came from.

## 3. Evidence Appendix

If the candidate depends on existing implementation evidence that is too large or too detailed for the main Spec, the evidence must be placed in a unit appendix and referenced from `evidence_appendix_ref`.

Evidence appendix rules:

1. `existing_implementation` requires an evidence appendix.
2. `mixed` requires an evidence appendix.
3. `new_design` requires `evidence_appendix_ref=none`.
4. `replacement` requires `evidence_appendix_ref=none`.

## 4. Direct First-Stable Onboarding

`unit_init` may create the first stable Spec for a historical unit only when the selected behavior baseline is already accepted and fully reviewable before stable writeback.
For this policy, accepted means the command can state the selected behavior without resolving business intent, evidence conflicts, material unknowns, or ownership boundaries during the stable writeback.

Direct first-stable onboarding is allowed only when all of the following are true:

1. the target is a historical unit whose current behavior is being captured, not redesigned
2. the selected behavior can be written into the stable main Spec and any explicitly referenced stable appendices in the same round without relying on raw implementation evidence as stable truth
3. implementation, test, runtime, or historical evidence has no unresolved conflict that affects selected behavior, unit responsibility, boundaries, acceptance, rule binding, or repository ownership
4. every material unknown is either resolved before stable writeback or explicitly irrelevant to the stable behavior being captured
5. any shared rule, global rule, reusable mechanism, exception, or cross-unit boundary needed by the first stable Spec is already resolved through the proper rule-governance path before stable writeback
6. repository mapping and path ownership are explicit enough to write the unit registration and any required implementation path or support-surface registration without guessing
7. the resulting stable Spec satisfies `specflow/framework/spec_authoring_baseline.md`

Direct first-stable onboarding must stop before stable writeback when any of the following are true:

1. the selected behavior exists only as raw implementation inspection, test observation, runtime observation, or historical material that still needs interpretation
2. evidence is incomplete, conflicting, or still needs business confirmation
3. the stable Spec would need the executor to invent a behavior, boundary, acceptance, ownership, or rule decision while writing it
4. shared or global truth is unresolved
5. repository mapping or path ownership cannot be written from current repository truth without guessing

When direct first-stable onboarding stops because the accepted baseline is not ready for stable writeback, route to candidate creation instead.
Candidate creation must use the `source_basis` and `evidence_appendix_ref` rules in this policy.
If the candidate selects behavior from implementation, tests, runtime behavior, or historical material, the candidate must use `source_basis=existing_implementation` or `source_basis=mixed` and create the required evidence appendix in the same round.

Raw evidence used during first-stable onboarding is not stable behavior truth by itself.
Any behavior selected for stable writeback must be stated directly in the stable main Spec or in an explicitly referenced stable appendix created or confirmed in the same round.

## 5. Stable-Fork Candidate Source

When `unit_fork` creates a candidate from an existing stable Spec, the stable Spec is formal truth.

If the fork uses only stable formal truth plus the current round's selected design change, write:

```yaml
source_basis: new_design
evidence_appendix_ref: none
```

If the fork selects behavior from implementation, tests, runtime behavior, historical material, or other non-stable evidence, use the normal source decision rules from this policy and create the required evidence appendix in the same round.
