# Onboarding Decision Policy

This policy decides how a unit candidate records the source of selected behavior truth.

Only unit candidates are supported.

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

## 4. Stable-Fork Candidate Source

When `unit_fork` creates a candidate from an existing stable Spec, the stable Spec is formal truth.

If the fork uses only stable formal truth plus the current round's selected design change, write:

```yaml
source_basis: new_design
evidence_appendix_ref: none
```

If the fork selects behavior from implementation, tests, runtime behavior, historical material, or other non-stable evidence, use the normal source decision rules from this policy and create the required evidence appendix in the same round.
