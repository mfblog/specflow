# Unit Plan Command

## 1. Purpose

This command creates or updates the implementation plan for the current candidate.

## 2. Scope

By default it handles:

1. reading the current valid candidate pass gate
2. optionally running research preflight when implementation-critical unknowns still block a stable plan
3. deriving a stable-to-candidate change surface with `git diff` when the unit already has `stable`
4. identifying the changed execution surfaces of this round and defining their convergence targets
5. generating or updating `_plans/active/{unit}.md`
6. writing or updating `_plans/draft/{unit}.md` when planning cannot yet produce a consumable active plan
7. keeping active-plan bindings aligned with the current candidate, current formal global baseline state, and current Rule snapshot
8. stopping at a structured decision checkpoint only when key implementation direction is still not locked
9. ensuring the active plan is executable without chat context, placeholders, or unstated verification meaning
10. mapping explicit acceptance item `id` values from the candidate into implementation slices and verification targets

### 2.1 Command Read Summary

Read this summary before the detailed rules below.
It is navigation only and does not replace the preconditions, procedure, stop conditions, or output contract.

1. `unit_plan` exists to turn a passed candidate into an executable implementation plan.
2. The minimum inputs are the current candidate, current `_check_result/unit/{unit}.md`, required appendix and Rule files, and current global baseline when relevant.
3. A plan-ready result writes `docs/specs/_plans/active/{unit}.md` and advances the object to `unit_impl`.
4. If candidate truth is incomplete, the command falls back to `unit_check`; if implementation facts or direction are still unresolved, it keeps planning state without creating a consumable active plan.
5. Draft plans are work-in-progress only; downstream commands may consume only an active plan whose bindings still match current truth.

### 2.2 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_plan`-local entry, output, and stop rules.

Process-file consumption and writeback for `_check_result/unit/{unit}.md` and `_plans/active/{unit}.md` must follow `specflow/framework/process_snapshot_contract.md` Section 10. When deterministic snapshot validation tooling is available for the current process kind, the matching `snapshot validate-process` command is the mandatory tool-backed validation step before treating a process file as consumable, reporting an active handoff, or advancing lifecycle state.

Before reading `_check_result/unit/{unit}.md` as a usable pass gate, run `specflowctl command preflight --command unit_plan --object-type unit --object {unit}`. If command preflight is unavailable, run `snapshot validate-process --object-type unit --object {unit} --process check` explicitly. Manual hash output must not classify gate drift or trigger fallback cleanup.

### 2.3 Slice Work-State Protocol Adoption

`unit_plan` adopts `specflow/framework/slice_work_state_protocol.md` only for command-owned business slice tracking.
It does not create a dedicated work-state or review run-state file.

Adoption rules:

1. the state carriers are `docs/specs/_plans/draft/{unit}.md` and `docs/specs/_plans/active/{unit}.md`
2. `draft/{unit}.md` is a non-consumable planning carrier for blocked planning, decision checkpoints, bounded research notes, and open implementation facts
3. `active/{unit}.md` is the downstream-consumable carrier only when this command reaches `plan-ready`
4. business slices are execution surface entries, implementation slices, and verification targets
5. the required domain fields are owned by `docs/specs/_plans/draft/README.md`, `docs/specs/_plans/active/README.md`, and this command's procedure
6. dynamic slices are not a separate carrier concept for this command; newly discovered implementation facts are recorded in command-owned fields such as `open_modeling_unknowns`, `research_notes`, `changed_execution_surfaces`, and `slice_cutover_plan`
7. command-local convergence is the acceptance-item coverage check, execution-surface target review, verification-target review, and `Plan Executability` review
8. closure can advance to implementation only when an active plan exists, covers the accepted acceptance item set, passes handoff validation, and satisfies the `plan-ready` rules in Section 4
9. if planning discovers missing behavior truth, boundary truth, or acceptance truth, the result is `truth-fallback`; do not add another planning slice to compensate

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=unit_plan`
3. a current valid `docs/specs/_check_result/unit/{unit}.md` exists
4. the current candidate still aligns with the current formal global baseline state
5. read any explicitly referenced candidate appendix files and bound Rule files
6. if this round may raise a checkpoint, read `specflow/framework/checkpoint_protocol.md`
7. read `specflow/framework/candidate_intent_policy.md` and the selected intent standard for the current candidate

## 4. Procedure

1. read the candidate Spec and all required appendix or Rule files
2. read `s_g_rule_repository_baseline.md` if it exists
3. read `candidate_intent` from the candidate frontmatter and apply the selected intent standard from `candidate_intent_policy.md`
4. run command preflight for `unit_plan:{unit}` and stop before fallback cleanup if authoritative validation is unavailable
5. read `docs/specs/_check_result/unit/{unit}.md`
6. verify the pass gate bindings are still valid using only the preflight or `snapshot validate-process` result as the authoritative validation source
7. when a required appendix is an evidence appendix, use it only to confirm the pass gate still reviewed the same candidate evidence; do not derive implementation requirements from it
8. re-read the candidate `Testability / Acceptance Criteria` section and confirm it still contains the same explicit acceptance item set accepted by the pass gate
9. if the pass gate is invalid, if the accepted acceptance-item set no longer matches current candidate truth, or if the candidate no longer satisfies the acceptance-item contract, stop immediately:
   - delete `_check_result/unit/{unit}.md`
   - delete `_plans/draft/{unit}.md` if it exists
   - delete `_plans/active/{unit}.md` if it exists
   - delete `_verify_result/unit/{unit}.md` if it exists
   - fall back `_status.md` to `unit_check`
10. if the unit already has `stable`, derive the planning basis required by the selected intent standard before judging the round:
   - use `git diff --no-index -- docs/specs/units/stable/s_unit_{unit}.md docs/specs/units/candidate/c_unit_{unit}.md`
   - for `candidate_intent=change`, use the diff to identify which candidate sections changed in this round and which implementation slices need direct focus first
   - for `candidate_intent=repair`, treat the diff only as diagnostic input; the planning basis is the selected standard's repair scope, acceptance item ids, and current implementation deviation
   - do not treat the diff as a substitute for reading the full candidate, required appendix truth, or bound Rule truth
   - do not assume unchanged lines are irrelevant, because unchanged candidate truth may still constrain implementation
   - if the unit has no `stable` yet, skip this step
11. determine the planning result shape for this round before plan write-back:
   - every `unit_plan` run must end in exactly one of these result shapes: `plan-ready`, `truth-fallback`, `plan-blocked`, or `decision-checkpoint`
   - if research preflight is not required because implementation-critical unknowns are already sufficiently closed, treat the round as `plan-ready`
12. determine whether research preflight is required:
   - use it only when current candidate truth is already closed enough to investigate implementation, but key implementation-critical unknowns still prevent a stable plan
   - do not use research preflight to replace missing behavior truth, boundary truth, or acceptance truth in the candidate
   - if research confirms that the real blocker is incomplete candidate truth, do not continue planning and fall back to `unit_check`
13. after research preflight, allow only these three result shapes:
   - plan-ready: implementation-critical unknowns are closed enough to write a stable plan
   - truth-fallback: research found that candidate truth itself is still incomplete, so planning must fall back to `unit_check`
   - plan-blocked: candidate truth still stands, but planning is blocked on a clearly named external condition, further bounded research result, or human-supplied implementation fact
14. `decision-checkpoint` is a distinct result shape:
   - use it only when a `decision` checkpoint is actually raised because implementation direction is still unresolved
   - do not merge it into `plan-blocked`, because unresolved direction and missing implementation facts are different blocking causes
   - do not create or update `docs/specs/_plans/active/{unit}.md`
   - keep `_status.md` at `unit_plan` unless the checkpoint answer must first be written back into candidate truth or appendix truth
15. if the result is `truth-fallback`:
   - delete `_check_result/unit/{unit}.md`
   - delete `_plans/draft/{unit}.md` if it exists
   - delete `_plans/active/{unit}.md` if it exists
   - delete `_verify_result/unit/{unit}.md` if it exists
   - do not create or update `docs/specs/_plans/active/{unit}.md`
   - update `_status.md` to `unit_check`
   - report `fallback_reason_code=truth_incomplete`
16. if the result is `plan-blocked`:
   - create or update `docs/specs/_plans/draft/{unit}.md`
   - do not create or update `docs/specs/_plans/active/{unit}.md`
   - if an old `active/{unit}.md` still exists for the same round, revalidate whether it remains consumable; if not, delete it rather than leaving a stale active plan available to downstream commands
   - keep `_status.md` at `unit_plan`
   - report `fallback_reason_code=implementation_unknown`
   - record the blocking point, the missing condition, and the exact resume signal
17. determine whether a `decision` checkpoint is required:
   - only use it when key implementation direction is still not locked
   - if the unresolved decision changes behavior truth, boundary truth, or acceptance truth, the resume path must go back to `unit_check` after writeback
   - do not treat the checkpoint as permission to continue without that writeback
18. if a `decision` checkpoint is raised:
   - set the result shape to `decision-checkpoint`
   - create or update `docs/specs/_plans/draft/{unit}.md`
   - do not create or update `docs/specs/_plans/active/{unit}.md`
   - if an old `active/{unit}.md` still exists for the same round, revalidate whether it remains consumable; if not, delete it rather than leaving a stale active plan available to downstream commands
   - keep `_status.md` at `unit_plan` when the unresolved decision is implementation-direction only
   - report `fallback_reason_code=direction_unresolved`
   - use `resume_next_step=unit_check` only when the checkpoint answer must first be written back into candidate truth or appendix truth
19. create or update `docs/specs/_plans/active/{unit}.md` only when no checkpoint blocks planning and the result is `plan-ready`
20. if `docs/specs/_plans/draft/{unit}.md` exists for the same round and the round is now `plan-ready`, extract only the stabilized planning content into `active/{unit}.md`; do not rename the draft file in place
21. after a successful active-plan write for the current round, delete `docs/specs/_plans/draft/{unit}.md` if it exists
22. identify the changed execution surfaces of this round before finalizing either draft or active planning output:
   - define an execution surface as the concrete capability path that this round is actually changing inside the unit
   - do not force one whole-unit owner or one whole-unit path when the round touches only a narrower capability slice
   - name each execution surface directly enough that `unit_impl` and `unit_verify` can reuse the same surface labels without reinterpretation
23. for each changed execution surface, record at minimum:
   - current known path
   - target path for the end of this round
   - retirement goal naming which legacy dependency should stop being a required prerequisite
   - the first stable cutover slices that can advance now
24. if current implementation facts are still insufficient to name a target path or retirement goal safely, but candidate truth still stands:
   - keep the round at `unit_plan`
   - update `docs/specs/_plans/draft/{unit}.md`
   - record the missing implementation fact under `open_modeling_unknowns`
25. if planning discovers that the real blocker is missing behavior truth, missing boundary truth, or missing acceptance truth:
   - do not compensate inside plan text
   - treat the round as `truth-fallback`
26. map acceptance items into the plan before treating the round as `plan-ready`:
   - every current-gate acceptance item `id` from the candidate must appear in at least one implementation slice or one `Verification Targets` entry
   - each `Verification Targets` entry must list `acceptance_item_ids`, `verification_surface`, the evidence command or inspection to run, and the pass evidence expected by `unit_verify`
   - `Execution Surface Plan` entries should reference the acceptance item `id` values they implement when the execution surface is driven by an acceptance item
   - if a current-gate acceptance item is omitted from both implementation slices and verification targets while candidate truth still stands, the result is `plan-blocked`
   - if omission reveals that the acceptance item itself is vague, missing, or wrongly marked runnable, the result is `truth-fallback`
27. ensure the active plan records:
   - `Execution Surface Plan`
   - `Retirement Targets`
   - `Verification Targets`
   - `acceptance_item_plan_coverage`
   - execution slices rather than one undifferentiated implementation block
   - for each slice: objective, file scope, dependencies, implementation action, verification action, done condition, and current status
   - progress, blockers, and verification focus for this round
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `unit_appendix_snapshot`
   - `rule_snapshot`
28. complete `Plan Executability` review before treating the round as `plan-ready`:
   - the active plan must not contain placeholder instructions such as `TBD`, `TODO`, `follow up later`, `similar to above`, `后续补充`, `类似上面处理`, or equivalent unresolved work markers
   - each execution slice must be understandable without chat context, guidance discussion, or rejected design options
   - each key candidate behavior and each current-gate acceptance item `id` must map to at least one execution slice or verification target
   - each verification action must name the command, inspection, or evidence to collect, and state which candidate requirement it proves
   - if the active plan lacks these details while candidate truth still stands, the result is `plan-blocked` and the missing planning content must be recorded in a draft plan or blocking reason
   - if the missing planning detail reveals incomplete behavior, boundary, or acceptance truth, the result is `truth-fallback`
29. treat `plan-ready` as valid only when all of the following hold:
   - the changed execution surfaces of this round are identified
   - each changed execution surface has a target path
   - each changed execution surface has at least one explicit retirement goal
   - every current-gate acceptance item is covered by implementation or verification planning
   - the first implementation slices are stable enough to enter `unit_impl`
   - `Plan Executability` passes
30. update `_status.md`:
   - if the candidate is now ready for implementation -> `Next Command=unit_impl`
   - if candidate truth drift was discovered -> `Next Command=unit_check`
   - if research preflight found candidate truth gaps -> `Next Command=unit_check`
   - if research preflight is blocked on implementation-critical unknowns but no truth rewrite is pending -> keep `Next Command=unit_plan`
   - if the result is `decision-checkpoint` and no truth writeback is pending -> keep `Next Command=unit_plan`
   - if a `decision` checkpoint stopped planning and no truth writeback is pending -> keep `Next Command=unit_plan`
   - the deterministic command closure may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_plan --object-type unit --object {unit} --outcome <plan_ready|blocked|decision_checkpoint> --notes <status-note> --apply`
   - for `truth_fallback`, execute `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_plan --object-type unit --object {unit} --outcome truth_fallback --reason truth_incomplete --notes <status-note> --apply`

## 5. Stop Conditions

1. either a valid active plan file exists for the current candidate truth and records the changed execution surfaces, target paths, retirement targets, and first stable cutover slices, or planning stopped with no consumable active plan artifact because of fallback, bounded blocking, or checkpoint
2. any active plan created in this round passes `Plan Executability`
3. every current-gate acceptance item is either mapped into the active plan or is the reason planning stopped
4. `_status.md` points to the real next step

## 6. Output Contract

1. planning conclusion
2. whether an active plan file was written, updated, or intentionally not created because planning stopped at fallback, bounded blocking, or a checkpoint
3. whether a draft plan file was written, updated, deleted, or intentionally omitted
4. plan binding result
5. candidate intent and selected planning-basis result
6. stable-to-candidate change-surface review result when `stable` exists
7. research preflight result when research preflight was used
8. changed execution surfaces and their target-path result
9. acceptance-item coverage result by `id`
10. retirement-target planning result
11. `handoff validation result`
12. `Plan Executability` result
13. cleanup result when planning fell back to `unit_check`
14. `checkpoint result` when a checkpoint stop was raised
   - when present, it must satisfy the fixed checkpoint fields defined by `specflow/framework/checkpoint_protocol.md`
15. `fallback_reason_code` for fallback, blocking, or checkpoint stops
16. blocking reason and resume signal when planning stayed at `unit_plan` without fallback
17. `_status.md` update result
18. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - when a checkpoint was raised or planning stayed blocked at `unit_plan`, also report `resume signal`
   - if `Next Command=unit_plan`, `why this next step` must explicitly state whether planning is waiting on implementation facts, unresolved direction, or truth writeback

Allowed checkpoint types:

1. `decision`

Allowed `fallback_reason_code` values:

1. `gate_missing`
2. `truth_drift`
3. `binding_drift`
4. `baseline_drift`
5. `rule_drift`
6. `truth_incomplete`
7. `implementation_unknown`
8. `direction_unresolved`

## 7. Non-Goals

1. direct code implementation
2. direct verification
3. rewriting candidate truth
