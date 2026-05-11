# Candidate Intent Policy

## 1. Purpose

This file defines the intent selector for current unit candidates.

It exists so one candidate lifecycle can safely handle both:

1. restoring implementation to an accepted stable baseline
2. changing the accepted behavior truth through the normal candidate chain

This file is an index and shared contract. It does not create new commands, a new lifecycle state, or a second implementation path.

## 2. Scope

This policy applies only to `unit` candidate main Specs.

It does not apply to:

1. `scenario` candidates
2. stable `unit` Specs
3. Rule files
4. repository mapping
5. process files except when a command consumes a unit candidate and must choose the current candidate intent standard

## 3. Candidate Intent Field

Every current `unit` candidate main Spec must record:

```yaml
candidate_intent: repair | change
```

Allowed values:

1. `repair`
   - the stable Spec remains the selected behavior truth
   - the candidate opens a controlled work round to restore implementation, tests, or verification to that stable truth
   - the candidate must also record `repair_basis`
2. `change`
   - the candidate changes, replaces, or extends the selected behavior truth
   - the candidate must not record `repair_basis`

Rules:

1. missing `candidate_intent` makes the candidate incomplete for `unit_check`
2. unknown `candidate_intent` makes the candidate incomplete for `unit_check`
3. `candidate_intent` is behavior-governance metadata, not `_status.md` state
4. `source_basis` and `evidence_appendix_ref` keep their existing source-selection meaning and must not be used as a substitute for `candidate_intent`

## 4. Intent Standard Index

Commands that consume a unit candidate must read this file, then read exactly one selected standard:

1. `candidate_intent=repair`
   - read `specflow/framework/candidate_intents/repair.md`
2. `candidate_intent=change`
   - read `specflow/framework/candidate_intents/change.md`

If the selected standard file is missing, the active command must stop before writing a pass gate, active plan, implementation progress, verification pass, promotion result, or lifecycle advance.

## 5. Command Consumption Rules

The standard unit command chain remains:

```text
unit_fork -> unit_check -> unit_plan -> unit_impl -> unit_verify -> unit_promote
```

Shared rules:

1. commands keep their existing ownership and output contracts
2. the selected intent standard changes only the command-local judgment criteria
3. commands must not duplicate the full repair or change standard in their own files
4. if a command discovers that the selected intent no longer matches the work, it must stop at the nearest legal earlier gate named by the selected standard

Plain meaning:

1. `candidate_intent` selects how the existing chain judges the round
2. it does not create `unit_repair`, `unit_repair_plan`, or any other new command

## 6. Relationship To Existing Candidate Source Fields

`source_basis` answers where selected candidate truth came from.

`candidate_intent` answers what kind of work round the candidate is.

Rules:

1. a `repair` candidate normally uses `source_basis=new_design` and `evidence_appendix_ref=none` because the selected truth comes from stable formal truth, not current implementation
2. if a repair round needs current implementation as selected truth, it is no longer a repair round and must become `candidate_intent=change`
3. a `change` candidate uses the existing source decision rules from `specflow/framework/onboarding_decision_policy.md`
4. an evidence appendix may describe current implementation, tests, or runtime behavior, but it is never the selected behavior truth unless the candidate main Spec states that selected rule
