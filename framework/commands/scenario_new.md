# Scenario New Command

## 1. Purpose

`scenario_new:{scenario}` creates the first candidate truth for a brand-new formal scenario object.

## 2. Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `scenario_new`-local entry, output, and stop rules.

## 3. Preconditions

1. the scenario ID is clear and non-conflicting
2. no current row for that scenario exists in `_status.md`
3. read `specflow/framework/onboarding_decision_policy.md` and decide the first candidate's `source_basis` and `evidence_appendix_ref`
4. if the first candidate uses `source_basis=existing_implementation` or `source_basis=mixed`, prepare the required scenario evidence appendix in the same round
5. if candidate truth, `_status.md`, or other commit-triggering governance files will change, read the git policy first

## 4. Procedure

1. create `docs/specs/scenarios/candidate/c_scenario_{scenario}.md`
2. initialize:
   - `source_basis`
   - `evidence_appendix_ref`
   - `repository_mapping_ref`
   - `unit_refs`
   - `shared_contract_refs`
   - `system_constraints_ref`
3. write or upsert `_status.md` row:
   - `Object Type=scenario`
   - `Object={scenario}`
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=scenario_check`
4. perform git close-out if required

## 5. Output Contract

The output must report:

1. candidate truth file write result
2. initialized `source_basis`
3. initialized `evidence_appendix_ref` and evidence appendix write result when required
4. `_status.md` registration result
5. lifecycle-state transition result
6. `round conclusion`
7. `current state`
8. `next step`
9. `why this next step`
10. `next-stage entry gap`
11. git close-out result
12. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
13. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. creating stable scenario truth
2. editing unit code
