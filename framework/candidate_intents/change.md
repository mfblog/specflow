# Change Candidate Intent Standard

## 1. Purpose

This file defines how the standard unit candidate command chain behaves when:

```yaml
candidate_intent: change
```

A change candidate is used when the round changes, replaces, extends, or reselects the unit behavior truth.

## 2. Required Candidate Fields

A change candidate must record:

```yaml
candidate_intent: change
source_basis: new_design | existing_implementation | mixed | replacement
evidence_appendix_ref: none | <candidate appendix path>
```

Rules:

1. `repair_basis` is forbidden
2. `source_basis` and `evidence_appendix_ref` follow `specflow/framework/onboarding_decision_policy.md`
3. if selected behavior depends on current implementation, tests, runtime behavior, or historical material, the candidate must use `source_basis=existing_implementation` or `source_basis=mixed` and provide the required evidence appendix
4. if the candidate replaces existing behavior without using it as selected truth, it must use `source_basis=replacement` and `evidence_appendix_ref=none`

## 3. Command Standards

### 3.1 `unit_fork`

When creating a change candidate:

1. derive the candidate from the current stable main Spec unless the source decision requires additional evidence
2. write `candidate_intent=change`
3. apply the existing stable-fork candidate source rules
4. record the behavior delta from stable in the candidate body clearly enough for `unit_check`, `unit_plan`, and `unit_verify`

### 3.2 `unit_check`

`unit_check` must judge the selected candidate truth on its own terms.

It must confirm:

1. the behavior delta from stable is explicit
2. boundaries and ownership are clear
3. acceptance items are structured and directly verifiable
4. source fields and evidence appendix are consistent
5. current implementation evidence is not treated as truth unless the candidate main Spec selects that rule

### 3.3 `unit_plan`

`unit_plan` uses stable-to-candidate behavior diff, candidate acceptance items, and required appendix or Rule truth to produce implementation slices.

### 3.4 `unit_impl`

`unit_impl` implements the selected candidate truth and may change behavior only within the boundaries written in the candidate.

If implementation discovers missing behavior truth, boundary truth, or acceptance truth, it must stop and fall back to `unit_check`.

### 3.5 `unit_verify`

`unit_verify` proves current implementation satisfies the selected candidate truth and current-gate acceptance items.

### 3.6 `unit_promote`

`unit_promote` promotes the selected candidate truth into stable according to the existing promotion rules.

Candidate-only `candidate_intent` metadata must not be copied into the stable main Spec.

## 4. Non-Goals

A change candidate does not allow:

1. chat-only behavior decisions
2. implementation evidence to become truth without candidate main-Spec selection
3. bypassing source-basis and evidence-appendix rules
4. bypassing the standard unit candidate command chain
