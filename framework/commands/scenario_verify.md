# Scenario Verify Command

## 1. Purpose

`scenario_verify:{scenario}` verifies whether the current candidate scenario is actually wired from trigger to outcome under the current bound object set.

## 2. Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `scenario_verify`-local entry, output, and stop rules.

## 3. Preconditions

1. `_status.md` says `Object Type=scenario`, `Active Layer=candidate`, `Next Command=scenario_verify`
2. current valid `_check_result/scenario/{scenario}.md` exists
3. read `specflow/framework/candidate_handoff_contract.md`
4. if `_verify_result/scenario/{scenario}.md`, `_status.md`, or other commit-triggering governance files may change, read the git policy first

## 4. Procedure

1. read current candidate scenario truth
2. revalidate `_check_result/scenario/{scenario}.md` according to the `scenario_check -> scenario_verify` handoff in `specflow/framework/candidate_handoff_contract.md`
3. revalidate current repository mapping, bound units, shared contracts, and baseline snapshots
4. verify the declared trigger-to-outcome path from entry to claimed outcome
5. report `affected_units` when implementation work is still required downstream
6. if pass, write `_verify_result/scenario/{scenario}.md` so it satisfies the `scenario_verify -> scenario_promote` handoff, then advance `Next Command=scenario_promote`
7. if the check handoff is invalid or repository mapping, unit, shared-contract, or baseline bindings drifted, fall back to `scenario_check` with the matching allowed `fallback_reason_code`
8. perform git close-out if required

## 5. Output Contract

The output must report:

1. verification gate result
2. `_verify_result/scenario/{scenario}.md` write, delete, or keep result
3. `_status.md` update result
4. `affected_units` when downstream implementation is still required
5. `round conclusion`
6. `current state`
7. `next step`
8. `why this next step`
9. `next-stage entry gap`
10. git close-out result
11. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
12. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. replacing `unit_impl:{unit}`
2. implicitly repairing affected units
