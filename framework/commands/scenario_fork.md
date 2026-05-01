# Scenario Fork Command

## 1. Purpose

`scenario_fork:{scenario}` opens a new candidate scenario round from the current stable scenario truth.

## 2. Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `scenario_fork`-local entry, output, and stop rules.

## 3. Preconditions

1. `_status.md` says `Object Type=scenario`, `Active Layer=stable`, `Next Command=scenario_fork`
2. stable scenario truth exists
3. if candidate truth, candidate-side process files, `_status.md`, or other commit-triggering governance files may change, read the git policy first
4. read `specflow/framework/onboarding_decision_policy.md` for stable-fork candidate source handling

## 4. Procedure

1. read stable scenario truth
2. apply the stable-fork candidate source rule from `specflow/framework/onboarding_decision_policy.md` Section 6.1
   - if the fork uses only stable formal truth plus the current round's selected design changes, prepare `source_basis=new_design` and `evidence_appendix_ref=none`
   - if the fork selects behavior from implementation, tests, runtime behavior, historical material, or other non-stable evidence, prepare the required `source_basis`, `evidence_appendix_ref`, and candidate evidence appendix in the same round
   - if that source decision or evidence appendix is not ready, stop before writing the candidate main Spec
3. create or overwrite `docs/specs/scenarios/candidate/c_scenario_{scenario}.md` and write the prepared `source_basis` and `evidence_appendix_ref` fields in the same candidate write
4. carry forward stable bindings
5. delete outdated candidate-side scenario process files if they exist
6. write `_status.md`:
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=scenario_check`
7. perform git close-out if required

## 5. Stop Conditions

1. the new `candidate` exists with valid `source_basis` and `evidence_appendix_ref`
2. candidate-side scenario process files are cleaned up
3. `_status.md` is updated

## 6. Output Contract

The output must report:

1. candidate truth file write result
2. initialized `source_basis`
3. initialized `evidence_appendix_ref` and evidence appendix write result when required
4. candidate-side scenario process file cleanup result
5. lifecycle-state transition result
6. `_status.md` update result
7. `round conclusion`
8. `current state`
9. `next step`
10. `why this next step`
11. `next-stage entry gap`
12. git close-out result
13. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
14. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 7. Non-Goals

1. stable verification
2. scenario promotion
