# Repair Candidate Intent Standard

## 1. Purpose

This file defines how the standard unit candidate command chain behaves when:

```yaml
candidate_intent: repair
```

A repair candidate is used when the stable Spec remains the correct behavior truth, but current implementation, tests, or verification need a controlled candidate work round to return to that truth.

## 2. Required Candidate Fields

A repair candidate must record:

```yaml
candidate_intent: repair
repair_basis: s_unit_{unit}@<version>
source_basis: new_design
evidence_appendix_ref: none
```

Rules:

1. `repair_basis` must name the stable unit version being restored
2. `repair_basis` must match the stable main Spec version available when the repair candidate is created
3. `source_basis=new_design` means the selected behavior comes from stable formal truth, not from current implementation
4. `evidence_appendix_ref` must be `none`
5. a repair candidate must include a `Repair Scope` section

## 3. Repair Scope

The `Repair Scope` section must directly state:

1. the stable acceptance item ids being restored
2. the observed implementation or evidence deviation that required the repair round
3. the implementation surfaces expected to change
4. the verification evidence that must prove the restored behavior

Rules:

1. `Repair Scope` is candidate-only repair guidance
2. it must not redefine behavior truth
3. promotion must not carry `Repair Scope` into the stable main Spec as stable behavior truth

## 4. Command Standards

### 4.1 `unit_fork`

When creating a repair candidate:

1. derive the candidate from the current stable main Spec
2. set the candidate version to the next `PATCH`
3. write `candidate_intent=repair`
4. write `repair_basis`
5. write `source_basis=new_design`
6. write `evidence_appendix_ref=none`
7. add the minimal `Repair Scope` section

### 4.2 `unit_check`

`unit_check` must verify that a repair candidate does not change stable behavior truth.

Blocking repair violations include:

1. changing public protocol meaning
2. changing field meaning, default values, validation rules, or error semantics
3. changing ownership boundary
4. changing state-machine or main-flow semantics
5. changing acceptance meaning or pass conditions
6. using current implementation as selected behavior truth

If any violation exists, the result must be `fix_required` and the next legal truth step is to convert the candidate to `candidate_intent=change`, then rerun `unit_check`.

### 4.3 `unit_plan`

`unit_plan` must plan from the repair gap, not from stable-to-candidate behavior diff.

The active plan must map:

1. each affected stable acceptance item id
2. the observed deviation
3. the implementation slices needed to restore the behavior
4. the verification target for the restored behavior

### 4.4 `unit_impl`

`unit_impl` may change implementation and tests only to restore the repair basis.

If implementation discovers that the repair cannot stand without changing behavior truth, ownership boundary, or acceptance meaning, implementation must stop and fall back to `unit_check`.

### 4.5 `unit_verify`

`unit_verify` must prove that the current implementation satisfies the repair basis and the repair candidate acceptance items.

It must not treat a new behavior, a relaxed pass condition, or an implementation-only workaround as repair success.

### 4.6 `unit_promote`

When promoting a repair candidate:

1. the stable version must be a `PATCH` version of the repair basis unless another rule-governed write in the same round requires a higher version
2. candidate-only fields `candidate_intent` and `repair_basis` must not be copied into the stable main Spec
3. candidate-only `Repair Scope` must not be copied into the stable main Spec as behavior truth
4. the stable acceptance coverage summary records the verification evidence for the restored behavior

## 5. Non-Goals

A repair candidate must not:

1. silently change accepted behavior
2. preserve current implementation merely because it exists
3. use an evidence appendix as behavior truth
4. create a new lifecycle outside the standard unit candidate chain
