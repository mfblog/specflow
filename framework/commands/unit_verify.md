# Unit Verify Command

## 1. Purpose

This command verifies whether current implementation aligns with the current `candidate`.

## 2. Scope

By default it handles:

1. candidate-versus-implementation alignment verification
2. goal-backward verification from acceptance claims into implementation evidence
3. structured verification evidence generation
4. structural convergence verification for the changed execution surfaces of the round
5. writing `_verify_result/unit/{unit}.md`
6. deciding whether the unit may enter `unit_promote`
7. stopping at a `human_verify` checkpoint only when automation is still insufficient to close confidence
9. requiring fresh evidence for every pass claim made in the current verification round
10. reporting evidence against each explicit acceptance item `id` instead of only broad feature blocks

### 2.1 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_verify`-local entry, output, and stop rules.

Process-file consumption and writeback for `_check_result/unit/{unit}.md`, `_plans/active/{unit}.md`, and `_verify_result/unit/{unit}.md` must follow `specflow/framework/process_snapshot_contract.md` Section 10. When deterministic snapshot validation tooling is available for the current process kind, the matching `snapshot validate-process` command is the mandatory tool-backed validation step before treating a process file as consumable, reporting a verification pass, or advancing lifecycle state.

Before reading `_check_result/unit/{unit}.md` or `_plans/active/{unit}.md` as usable verification inputs, run `specflowctl command preflight --command unit_verify --object-type unit --object {unit}`. If command preflight is unavailable, run `snapshot validate-process` for both `check` and `plan` explicitly. After writing `_verify_result/unit/{unit}.md`, run `snapshot validate-process --process verify` before reporting a verification pass.

### 2.2 Slice Work-State Protocol Adoption

`unit_verify` adopts `specflow/framework/slice_work_state_protocol.md` only for command-owned verification evidence coverage.
It does not create a dedicated work-state or review run-state file.

Adoption rules:

1. the state carrier is `docs/specs/_verify_result/unit/{unit}.md` when verification evidence is written
2. the business slices are acceptance-item evidence rows and structure-convergence rows
3. the required domain fields are the structured verification evidence matrix, `Structure Convergence Matrix`, `Coverage Summary`, and the verify-result snapshot fields required by `process_snapshot_contract.md`
4. dynamic slices are not a separate carrier concept for this command
5. newly discovered evidence surfaces, integration paths, or deviations are recorded in the evidence matrix, structure convergence matrix, or deviation list
6. command-local convergence is goal-backward verification from acceptance item `id` values into current evidence, plus structural convergence for changed execution surfaces
7. verification closure can support promotion entry only when every required current-gate acceptance item has an allowed current-round evidence status and every required structure convergence check is closed
8. if verification discovers missing behavior truth, boundary truth, acceptance truth, or invalid planning coverage, it must stop through the command's layered recovery path instead of adding another verification slice to compensate

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=unit_verify`
3. a current valid `_check_result/unit/{unit}.md` exists
4. a current valid `_plans/active/{unit}.md` exists
5. the candidate still aligns with the current formal global baseline state
6. read required candidate appendix files and bound Rule files
7. if this round may raise a checkpoint, read `specflow/framework/checkpoint_protocol.md`
8. read `specflow/framework/candidate_intent_policy.md` and the selected intent standard for the current candidate

## 4. Procedure

1. read the candidate Spec, required appendix files, and Rule files
2. read `candidate_intent` from the candidate frontmatter and apply the selected intent standard from `candidate_intent_policy.md`
3. run command preflight for `unit_verify:{unit}` and stop before verification judgment if authoritative validation is unavailable
4. read the pass gate and active plan
5. validate all required bindings using the preflight or `snapshot validate-process` result as the authoritative validation source
6. confirm that the candidate acceptance item set still matches the pass gate and active plan coverage
7. if the pass gate or plan is invalid, or if the acceptance item set no longer matches the pass gate or active plan, stop through layered recovery:
   - `truth_layer`: truth, acceptance, Rule, baseline, appendix, or binding drift; delete the unit candidate-side process chain and set `_status.md` to `unit_check`
   - `gate_layer`: check gate process shape is invalid while current truth and bindings still match; delete `_check_result/unit/{unit}.md` and set `_status.md` to `unit_check`
   - `plan_layer`: active plan is missing, malformed, not tool-valid, or missing coverage while the check gate still covers current truth; delete `_plans/draft/{unit}.md`, `_plans/active/{unit}.md`, and `_verify_result/unit/{unit}.md` if present, then set `_status.md` to `unit_plan`
8. establish the current-round evidence basis before making any pass claim:
   - evidence must be collected or refreshed in the current `unit_verify` run
   - old test output, previous command output, agent reports, or implementation claims may be used only as pointers to what must be rechecked
   - each evidence item must name the command or inspection used, the checked target, the observed result, and the candidate requirement it proves
   - no acceptance claim may be marked pass without at least one current evidence item
9. perform goal-backward verification for each current-gate acceptance item instead of stopping at artifact existence
10. build the structured verification evidence matrix around acceptance item `id` values:
   - each row must name `acceptance_item_id`, `target`, `verification_surface`, `implementation_surface`, `verification_method`, `evidence`, and `status`
   - `status` must be exactly one of `pass`, `fail`, `partial`, `not_checked`, or `not_runnable_yet`
   - `pass` requires current-round evidence that directly proves the item's `pass_condition`
   - `not_runnable_yet` may be used only when the candidate item itself explicitly records `not_runnable_yet` and the current run confirms the same missing runnable surface still exists
   - passing tests that do not prove the item's `pass_condition` must be reported as insufficient evidence, not as `pass`
11. judge the verification target according to the selected intent standard:
   - `change` must prove current implementation satisfies the selected candidate truth
   - `repair` must prove current implementation satisfies the repair basis and the candidate acceptance items; new behavior or relaxed pass conditions do not count as repair success
12. for each key claim, judge at minimum:
   - `existence`: the required artifact, path, handler, test, or integration point exists
   - `substance`: the artifact contains meaningful implementation rather than hollow placeholder structure
   - `wiring`: the artifact is actually connected to the main execution path, user path, or protocol path required by the current candidate
13. if a required outcome depends on cross-file integration, name that integration path directly in the evidence matrix
14. if implementation pieces exist but are not wired into the claimed path, treat that as `implementation_deviation` rather than as successful existence evidence
15. verify the changed execution surfaces named by the active plan, not just the unit as an undifferentiated whole
16. produce both:
   - the structured verification evidence matrix
   - a `Structure Convergence Matrix`
17. for each changed execution surface in the `Structure Convergence Matrix`, report at minimum:
   - `execution_surface`
   - `behavior_alignment`
   - `target_path_evidence`
   - `legacy_not_required_evidence`
   - `retirement_result`
   - `deviation_reason`
18. treat any of the following as direct `implementation_deviation`:
   - the current execution surface still requires a legacy path before the planned target path can succeed
   - a legacy helper, patch, wrapper, or equivalent dependency named in `Retirement Targets` is still required
   - a new implementation exists but the target path was not actually cut over
   - a core retirement target is not achieved but the round still attempts to enter `unit_promote`
19. output `Coverage Summary` by acceptance item status, including totals for `pass`, `fail`, `partial`, `not_checked`, and `not_runnable_yet`
20. determine whether a `human_verify` checkpoint is required:
   - use it only when automated verification is insufficient but a small amount of human effect judgment can close the remaining uncertainty
   - if human verification confirms implementation deviation while candidate truth still stands, use `implementation_layer` and fall back to `unit_impl`
   - if human verification shows acceptance truth itself is still incomplete, use `truth_layer` and fall back to `unit_check`
21. classify deviations with the shared severity meanings defined by `specflow/framework/severity_policy.md`
22. conclude:
   - if `fail` exists, do not enter `unit_promote`
   - if a current-gate acceptance item is `partial`, `not_checked`, or `not_runnable_yet`, promotion is allowed only if `specflow/framework/downgrade_policy.md` explicitly allows that non-pass evidence state for the current round
   - if any current-gate acceptance item lacks current-round evidence, the result cannot be pass and must remain `not_checked`, `partial`, or `evidence_incomplete`
   - if tests pass but do not prove the candidate requirement, report the gap instead of treating the test result as requirement evidence
   - if key deviations are cleared, retirement targets are satisfied, and evidence is complete, promotion may proceed
23. write or update `docs/specs/_verify_result/unit/{unit}.md`
   - the verify result must include the acceptance-item evidence matrix and the acceptance item `id` set it covers
24. update `_status.md`:
   - if ready to promote -> `Next Command=unit_promote`
   - if implementation has deviations but candidate truth still stands -> `Next Command=unit_impl`
   - if candidate truth or formal global baseline must be re-closed -> `Next Command=unit_check`
   - if verification evidence is still incomplete but no upstream truth drift exists -> `Next Command=unit_verify`
   - the deterministic command closure may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_verify --object-type unit --object {unit} --outcome <ready_to_promote|implementation_deviation|evidence_incomplete|human_verify> --notes <status-note> --apply`
   - for `truth_fallback`, execute `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_verify --object-type unit --object {unit} --outcome truth_fallback --reason <fallback_reason_code> --notes <status-note> --apply`

## 5. Stop Conditions

1. candidate-versus-code alignment is clear
2. changed execution surfaces either prove structural convergence or produce an explicit deviation result
3. every explicit acceptance item has one allowed status
4. current-round evidence exists for every pass claim
5. whether promotion is allowed is clear
6. `_status.md` points to the real next executable step
7. if pass gate, plan, or acceptance-item coverage was invalid, verification stopped and `_status.md` fell back to `unit_check`

## 6. Output Contract

1. verification conclusion
2. structured verification evidence matrix by acceptance item `id`
3. candidate intent verification result
4. `Structure Convergence Matrix`
5. `Coverage Summary`
6. current-round evidence freshness result
7. goal-backward evidence result
8. downgrade decision when `partial`, `not_checked`, or `not_runnable_yet` exists
9. verify-result write-back result
10. cleanup result when verification fell back to `unit_check`
11. `checkpoint result` when a checkpoint stop was raised
   - when present, it must satisfy the fixed checkpoint fields defined by `specflow/framework/checkpoint_protocol.md`
12. `fallback_reason_code` for fallback or checkpoint stops
13. deviation list
14. fallback reason if pass gate or plan was invalid
15. next-step suggestion
16. `_status.md` update result
17. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - when a `human_verify` checkpoint was raised, also report `resume signal`
   - if `Next Command` remains `unit_verify` or falls back to `unit_impl`, `why this next step` must explicitly state whether the remaining blocker is missing evidence, human-effect judgment, or implementation deviation

Allowed checkpoint types:

1. `human_verify`

Allowed `fallback_reason_code` values:

1. `gate_missing`
2. `truth_drift`
3. `binding_drift`
4. `baseline_drift`
5. `rule_drift`
6. `implementation_deviation`
7. `evidence_incomplete`
8. `truth_incomplete`

## 7. Non-Goals

1. directly changing code
2. directly rewriting candidate truth
3. advancing an independent stable `g_` rule state machine
