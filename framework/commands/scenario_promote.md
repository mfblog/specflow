# Scenario Promote Command

## 1. Purpose

`scenario_promote:{scenario}` promotes the current candidate scenario into the new stable scenario truth.

## 2. Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `scenario_promote`-local entry, output, and stop rules.

Process-file consumption for `_verify_result/scenario/{scenario}.md` and stable acceptance summary writeback under `_verify_result/stable/scenario/{scenario}.md` must follow `specflow/framework/process_snapshot_contract.md` Section 9, including the tool-backed validation rule when snapshot validation tooling is available for scenario process files.

## 3. Preconditions

1. `_status.md` says `Object Type=scenario`, `Active Layer=candidate`, `Next Command=scenario_promote`
2. current valid `_verify_result/scenario/{scenario}.md` exists
3. read `specflow/framework/candidate_handoff_contract.md`
4. read `specflow/framework/recovery_policy.md` before promotion

## 4. Procedure

1. read and re-check the latest `_verify_result/scenario/{scenario}.md`
2. read and re-check `docs/specs/scenarios/candidate/c_scenario_{scenario}.md`
3. validate `_verify_result/scenario/{scenario}.md` according to the `scenario_verify -> scenario_promote` handoff in `specflow/framework/candidate_handoff_contract.md`
   - the verify result must cover the current candidate scenario acceptance item `id` set exactly
   - each current-gate acceptance item must have an allowed promotion state according to `scenario_verify` and any applicable downgrade policy
4. if `_verify_result/scenario/{scenario}.md` is invalid, stop before truth-file mutation:
   - if candidate truth, repository mapping snapshot, required unit bindings, bound Rule snapshots, or formal global baseline snapshots drifted, delete current-round scenario `_check_result/scenario/{scenario}.md` and `_verify_result/scenario/{scenario}.md`, update `_status.md` to `Next Command=scenario_check`, and use the matching standardized code: `truth_drift`, `binding_drift`, `rule_drift`, or `baseline_drift`
   - if only the scenario check gate process shape is malformed while current scenario truth and bindings still match, delete `_check_result/scenario/{scenario}.md`, update `_status.md` to `Next Command=scenario_check`, and use `failure_layer=gate_layer`
   - if only verification coverage is stale, malformed, or incomplete while the check gate still covers current truth, delete `_verify_result/scenario/{scenario}.md`, update `_status.md` to `Next Command=scenario_verify`, and use `fallback_reason_code=evidence_incomplete` with `failure_layer=evidence_layer`
5. before the first truth-file mutation, resolve every current `unit_refs` and `rule_refs` entry in the candidate scenario:
   - each `unit_refs` entry must resolve to existing stable-layer unit truth
   - each `rule_refs` entry must resolve to existing stable-layer Rule truth
   - `rule_refs=none` is valid only when the candidate scenario formally binds no Rule
   - do not treat `unit_snapshot`, `rule_snapshot`, `bound_objects`, repository history, or directory shape as a replacement for the formal refs
   - if any dependency is candidate-layer, missing, or not safely resolvable, stop before stable writeback, keep candidate semantics, report the affected dependency and its current legal next step from `_status.md` when present, use `fallback_reason_code=stable_dependency_not_ready` with `failure_layer=dependency_readiness_layer`, and keep `Next Command=scenario_promote`
   - after dependency landing, retry `scenario_promote` by revalidating the current verify handoff and stable dependency readiness; require fresh `scenario_check` and `scenario_verify` only when scenario truth, scenario bindings, repository mapping, Rule bindings, baseline, acceptance item ids, or verification evidence changed
6. continue only when candidate truth, verification coverage, required bindings, and stable dependency readiness still remain valid
7. before the first truth-file mutation, capture the recovery baseline required by `recovery_policy.md`
8. write `docs/specs/scenarios/stable/s_scenario_{scenario}.md`
9. write the minimal stable acceptance coverage summary for this promoted round before current-round verify cleanup:
   - target path: `docs/specs/_verify_result/stable/scenario/{scenario}.md`
   - record the promoted stable truth file, version, fingerprint, acceptance item `id` set, each item's final verification status, and the key evidence source refs from the current `_verify_result/scenario/{scenario}.md`
   - this summary is not behavior truth and must not replace the stable scenario Spec's `Testability / Acceptance Criteria` section
   - if this summary cannot be written while promotion otherwise needs to delete the current `_verify_result/scenario/{scenario}.md`, stop before cleanup rather than losing the only acceptance coverage record for the promoted round
10. write `_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=scenario_fork`
11. only after `_status.md` has already been updated to `Candidate=no`, delete:
   - `docs/specs/scenarios/candidate/c_scenario_{scenario}.md`
   - current-round scenario evidence appendix files
   - current-round scenario `_check_result/scenario/{scenario}.md`
   - current-round scenario `_verify_result/scenario/{scenario}.md`
12. if the command is interrupted after promotion internals started but before final cleanup finished, run incomplete promotion recovery according to `recovery_policy.md` instead of claiming success

## 5. Output Contract

The output must report:

1. stable truth file write result
2. candidate truth file delete result
3. evidence appendix deletion or absorption result
4. `_check_result/scenario/{scenario}.md` and `_verify_result/scenario/{scenario}.md` cleanup result
5. stable acceptance coverage summary write result
6. stable dependency readiness result, including affected dependency refs when promotion stopped before stable writeback
7. lifecycle-state transition result
8. `_status.md` update result
9. `handoff validation result`, including acceptance-item coverage validation and stable dependency readiness validation
10. fallback cleanup result when verification became invalid before promotion could start
11. `fallback_reason_code` if verification became invalid or stable dependency readiness failed
12. `fallback_reason_code=promotion_recovery` when incomplete promotion recovery occurred
13. recovery-state explanation if incomplete promotion occurred
14. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
   - when promotion recovery occurred, also report `resume signal`

Allowed checkpoint types:

1. none

Allowed `fallback_reason_code` values:

1. `binding_drift`
2. `truth_drift`
3. `baseline_drift`
4. `rule_drift`
5. `evidence_incomplete`
6. `stable_dependency_not_ready`
7. `promotion_recovery`

## 6. Non-Goals

1. unit promotion
2. changing repository mapping
