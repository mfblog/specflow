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
4. if the check handoff is missing or invalid, or if repository mapping, unit, shared-contract, or baseline bindings drifted, stop verification and fall back to `scenario_check`:
   - delete `_check_result/scenario/{scenario}.md` when it exists but no longer qualifies as a current valid check gate
   - delete `_verify_result/scenario/{scenario}.md` when it exists because it cannot remain current after the check handoff failed
   - update `_status.md` to `Next Command=scenario_check`
   - report the matching `fallback_reason_code` allowed by the `scenario_check -> scenario_verify` handoff in `specflow/framework/candidate_handoff_contract.md`
5. verify the declared trigger-to-outcome path from entry to claimed outcome
6. if the path cannot pass because one or more bound units still require unit-local truth, planning, implementation, verification, binding, or baseline work, stop as `blocked_by_affected_units`:
   - do not write `_verify_result/scenario/{scenario}.md`
   - delete `_verify_result/scenario/{scenario}.md` when it exists because it cannot remain current while the scenario is blocked by affected unit work
   - keep `_check_result/scenario/{scenario}.md` only when it still qualifies as the current valid check gate
   - keep or set the scenario row in `_status.md` to `Active Layer=candidate` and `Next Command=scenario_verify`
   - read `_status.md` rows for each `affected_unit` and report that unit's current legal `Next Command`
   - report `blocked_reason_code` using the allowed values in Section 6 of this file
   - route follow-up through natural-language routing from current repository truth so each affected unit re-enters its own legal unit command chain
7. if pass, write `_verify_result/scenario/{scenario}.md` so it satisfies the `scenario_verify -> scenario_promote` handoff, then advance `Next Command=scenario_promote`
8. perform git close-out if required

## 5. Stop Conditions

1. the scenario check handoff is either valid and consumed, or invalid and cleaned up
2. whether the declared trigger-to-outcome path verifies under current bindings is clear
3. if affected units block scenario verification, no scenario verify result remains current, the scenario row still points to `scenario_verify`, and each affected unit's current legal next command is reported from `_status.md`
4. if verification passes, `_verify_result/scenario/{scenario}.md` holds the current verify gate
5. if verification falls back to `scenario_check`, no invalid scenario check or verify result remains
6. `_status.md` points to the real next executable step for the scenario row; when verification is blocked by affected units, the output names the affected units' current legal next commands as the immediate follow-up

## 6. Output Contract

The output must report:

1. verification gate result: `pass`, `fallback_to_scenario_check`, or `blocked_by_affected_units`
2. `_verify_result/scenario/{scenario}.md` write, delete, or keep result
3. `_check_result/scenario/{scenario}.md` cleanup result when verification fell back to `scenario_check`
4. `_status.md` update result
5. `affected_units` and each affected unit's current legal `Next Command` when downstream unit work is still required
6. `blocked_reason_code` when verification stopped as `blocked_by_affected_units`
7. `fallback_reason_code` when verification fell back to `scenario_check`
8. fallback reason when the check handoff was missing, invalid, or drifted
9. natural-language reroute requirement when verification stopped as `blocked_by_affected_units`
10. `round conclusion`
11. `current state`
12. `next step`
13. `why this next step`
14. `next-stage entry gap`
15. git close-out result
16. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
17. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

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

Allowed `blocked_reason_code` values for `blocked_by_affected_units`:

1. `truth_incomplete`
2. `gate_missing`
3. `truth_drift`
4. `binding_drift`
5. `baseline_drift`
6. `shared_contract_drift`
7. `implementation_unknown`
8. `implementation_deviation`
9. `evidence_incomplete`

These values are owned by the standard reason taxonomy in `specflow/framework/candidate_handoff_contract.md`.
The standardized code must appear before the natural-language explanation.

## 7. Non-Goals

1. replacing `unit_impl:{unit}`
2. implicitly repairing affected units
3. advancing affected unit lifecycle state
