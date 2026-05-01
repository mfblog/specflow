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
7. if the check handoff is missing or invalid, or if repository mapping, unit, shared-contract, or baseline bindings drifted, stop verification and fall back to `scenario_check`:
   - delete `_check_result/scenario/{scenario}.md` when it exists but no longer qualifies as a current valid check gate
   - delete `_verify_result/scenario/{scenario}.md` when it exists because it cannot remain current after the check handoff failed
   - update `_status.md` to `Next Command=scenario_check`
   - report the matching `fallback_reason_code` allowed by the `scenario_check -> scenario_verify` handoff in `specflow/framework/candidate_handoff_contract.md`
8. perform git close-out if required

## 5. Stop Conditions

1. the scenario check handoff is either valid and consumed, or invalid and cleaned up
2. whether the declared trigger-to-outcome path verifies under current bindings is clear
3. affected units are reported when downstream implementation work is still required
4. if verification passes, `_verify_result/scenario/{scenario}.md` holds the current verify gate
5. if verification falls back to `scenario_check`, no invalid scenario check or verify result remains
6. `_status.md` points to the real next executable step

## 6. Output Contract

The output must report:

1. verification gate result
2. `_verify_result/scenario/{scenario}.md` write, delete, or keep result
3. `_check_result/scenario/{scenario}.md` cleanup result when verification fell back to `scenario_check`
4. `_status.md` update result
5. `affected_units` when downstream implementation is still required
6. `fallback_reason_code` when verification fell back to `scenario_check`
7. fallback reason when the check handoff was missing, invalid, or drifted
8. `round conclusion`
9. `current state`
10. `next step`
11. `why this next step`
12. `next-stage entry gap`
13. git close-out result
14. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
15. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

Allowed checkpoint types:

1. none

Allowed `fallback_reason_code` values:

1. `gate_missing`
2. `truth_drift`
3. `binding_drift`
4. `baseline_drift`
5. `shared_contract_drift`

These values are owned by the `scenario_check -> scenario_verify` handoff in `specflow/framework/candidate_handoff_contract.md`.
The standardized code must appear before the natural-language explanation.

## 7. Non-Goals

1. replacing `unit_impl:{unit}`
2. implicitly repairing affected units
