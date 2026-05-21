# Unit Check Command

## 1. Purpose

This command checks whether a unit's `candidate` Spec is sufficiently closed to support stable downstream planning and implementation.

It is a review action, not a "store failed review results" action.

By default, closure means all of the following:

1. `progressability`
   - the unit behavior is clear enough to enter `unit_plan`
   - main flow, key protocols, key boundaries, error semantics, and acceptance criteria are strong enough to prevent planning or implementation divergence
2. `content completeness`
   - the candidate has formally acknowledged key behavior truth that affects implementation results
   - key decisions are not left outside the Spec in chat context, README vision, oral consensus, or author memory
3. `candidate design quality`
   - the candidate connects the user goal, first-round scope, selected direction, and acceptance criteria strongly enough to avoid implementing a merely well-formed but poor project design
4. the candidate is still aligned with the current formal global baseline state

## 2. Scope

By default this command reviews:

1. whether `progressability` holds
2. whether `content completeness` holds
3. whether `Rule Alignment` is explicit and internally consistent
4. whether bound Rule relations and body dependencies are consistent
5. whether shared-candidate signals require routing into rule governance or directly reporting a dual-source-of-truth conflict
6. whether the remaining blocker is actually a user-intent clarification or decision-point that must be written back before closure can pass
7. whether the candidate records a coherent current design rather than an over-broad, unresolved, or chat-dependent proposal
8. whether the candidate source fields and evidence appendix requirements from `onboarding_decision_policy.md` are satisfied
9. whether `Testability / Acceptance Criteria` contains explicit acceptance items that satisfy `spec_writing_guide.md` Section 6

### 2.1 Command Read Summary

Read this summary before the detailed rules below.
It is navigation only and does not replace the preconditions, procedure, stop conditions, or output contract.

1. `unit_check` exists to decide whether the current candidate is strong enough to become implementation input.
2. The minimum inputs are the current candidate main Spec, required appendix files, bound Rule files, and current global baseline when relevant.
3. A pass writes the current `_check_result/unit/{unit}.md` pass gate and advances the object to `unit_plan`.
4. A non-pass result must not write a failed check-result file; it either reports a blocker, asks for a checkpoint, or requires candidate-side repair before a fresh full-scope `unit_check` rerun.
5. Evidence appendix text proves what was reviewed; it does not become implementation truth.

### 2.2 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_check`-local entry, output, stop, and fresh-rerun rules.

Process-file writeback and validation for `_check_result/unit/{unit}.md` must follow `specflow/framework/process_snapshot_contract.md` Section 10. When deterministic snapshot validation tooling is available for the current process kind, the matching `snapshot validate-process` command is the mandatory tool-backed validation step before reporting a pass gate or lifecycle advance.

After writing `_check_result/unit/{unit}.md`, run `snapshot validate-process --object-type unit --object {unit} --process check` before reporting `pass`, `allow_next=true`, or `Next Command=unit_plan`. If validation tooling is unavailable or fails, delete no files by manual hash judgment, do not advance `_status.md`, and report the tooling validation gap or the tool-backed validation failure.

`unit_check` is not a "minimum can-move-forward review."
`unit_check pass` always means:

1. the current candidate may enter `unit_plan`
2. the current candidate already contains the key constraints needed as the truth input for implementation in this round

Result semantics for non-pass conclusions are fixed:

1. `blocked`
   - use when the smallest correct next step cannot be completed by executor-side repair alone in the current round
   - the blocker is waiting on user clarification, user decision, or rule-truth closure outside the active command's direct repair surface
   - if the blocker changes behavior truth, the answer must be written back before `unit_check` may pass
2. `fix_required`
   - use when the executor can already identify a concrete truth-side repair inside the current candidate, appendix, or explicit binding surface
   - no extra user choice is needed before that repair work starts
   - after the repair, the unit must return to `unit_check` rather than skipping forward

Authoritative rerun boundary:

This section is the `unit_check`-local elaboration of the centralized authoritative-run and non-authoritative-follow-up rules inherited above.

1. a new formal `unit_check` rerun may be entered either by explicit command syntax or by a later natural-language request that command routing correctly resolves to a fresh full-scope `unit_check` run for the current unit
   - after a prior `unit_check` ended as `blocked` or `fix_required`, that natural-language request must make rerun intent explicit enough to distinguish "rerun `unit_check` now" from "repair the candidate", "continue follow-up work", or "recheck only the reported blocker"
   - generic repair-oriented wording such as "fix it", "continue", "close this up", or equivalent wording does not by itself authorize a fresh authoritative `unit_check` rerun
2. for `unit_check`, a fresh full-scope run means rerunning the command's full mandatory closure surface for the current unit:
   - reread the current candidate main file plus all required appendix and Rule files
   - reread the current formal global baseline input when it exists
   - rerun the framework-baseline closure checks, including `progressability`, `content completeness`, binding checks, and baseline-alignment checks
   - re-judge the overall gate conclusion for the current candidate instead of confirming only the previously reported finding
3. truth repair performed after a `blocked` or `fix_required` result is not itself that rerun
4. any repair-side reassessment or scoped follow-up review performed after such repair is non-authoritative:
   - it may report only whether the reported findings appear resolved within the checked scope
   - it must not be labeled a formal `unit_check pass`
   - it must not write `docs/specs/_check_result/unit/{unit}.md`
   - it must not advance `_status.md` to `unit_plan`
   - checking only the repaired truth fragment, only the previously reported blocker, or any other narrowed review slice does not count as a fresh full-scope `unit_check` rerun

### 2.3 Unit Check Work State

`unit_check` adopts `specflow/framework/slice_work_state_protocol.md` for its command-local slice work state.
This command file owns the adoption details below.

`unit_check` uses one intermediate work-state file while the check is in progress:

```text
docs/specs/_check_work/unit/{unit}.md
```

This state carrier is not a Spec, not behavior truth, and not a pass gate.
It records only the current `unit_check` round's progress, slice status, input fingerprints, finding references, blocked reason, and resume step.

`unit_plan` must not consume `_check_work`.
The only downstream handoff gate from `unit_check` to `unit_plan` remains:

```text
docs/specs/_check_result/unit/{unit}.md
```

The work-state file must be created or refreshed before semantic slice review starts.
The mechanical maintenance commands are:

```text
specflowctl process check-work-init --object-type unit --object {unit}
specflowctl process check-work-validate --object-type unit --object {unit}
specflowctl process check-work-refresh --object-type unit --object {unit}
specflowctl process check-work-touch --object-type unit --object {unit}
```

Tooling may maintain only the mechanical fields allowed by `slice_work_state_protocol.md`: UTC timestamps, baseline slice skeleton, input fingerprints, stale slice marking, and structural validation.
Tooling must not write slice pass judgments, finding content, severity, or the final `pass`, `blocked`, or `fix_required` conclusion.

Command-specific adoption rules:

1. the state carrier is `docs/specs/_check_work/unit/{unit}.md`
2. the required run fields and slice fields are owned by `process_snapshot_contract.md` Section 9
3. baseline slices and cross-check slices are the required catalog in Section 2.6
4. dynamic slices are allowed only under the triggers in Section 2.6
5. cross-check slices are required and must close after their dependent local slices
6. input freshness and stale handling are owned by `process_snapshot_contract.md` Section 9.4
7. slice-set closure can support a `unit_check pass` only when Section 2.6 and the overall command gate rules also pass
8. missing durable truth, unclear ownership, or candidate-layer dependency use must become `blocked` or `fix_required`, not an extra implementation slice

Work-state reuse rules:

1. if no work-state file exists, create one
2. if an open work-state file exists and `last_updated_at` is within 2 hours, refresh stale status and reuse it
3. if an open work-state file is older than 2 hours and not older than 7 days, stop and require an explicit reuse-or-restart decision
4. if an open work-state file is older than 7 days, start a new file
5. if the existing work-state file is closed, malformed, or bound to stale truth that cannot be refreshed, start a new file or report the malformed state according to the tooling result

### 2.4 Logic Chain Closure

`unit_check` must prove the candidate's main logic chain, not only that sections exist.

The required chain is:

```text
user goal -> unit responsibility -> main flow -> boundary protocol -> output artifact -> acceptance -> unit_plan handoff
```

The chain is closed only when all of the following are true:

1. the user or actor goal is explicit enough to know what useful result this unit must provide
2. the unit responsibility explains why this unit, not another owner, owns that result
3. the main flow explains the normal path from entry input to produced result
4. boundary protocols name the public contract, port, adapter, event, store, trace, or support surface used at each boundary
5. output artifacts are named and have a producer, consumer, persistence or reporting boundary, and freshness meaning
6. acceptance items prove the stated responsibility and output artifacts, not only that files or headings exist
7. the candidate contains enough handoff information for `unit_plan` to plan implementation without inventing missing adapters, test entrypoints, outputs, or ownership decisions

Any missing link in this chain is a main logic-chain breakpoint.
A candidate with a main logic-chain breakpoint cannot pass.

### 2.5 Dependency Truth Boundary

`unit_check` checks one target candidate, but it must read the target candidate's formal dependency truth when the candidate's behavior depends on it.

Allowed formal dependency truth for `unit_check pass` is limited to:

1. stable unit truth referenced through `unit_refs`
2. stable Rule truth referenced through `rule_refs` or the formal stable global baseline
3. `docs/specs/repository_mapping.md` when path ownership or support-surface ownership matters
4. `docs/specs/rules/stable/s_g_rule_repository_baseline.md` when the global baseline exists
5. appendix files explicitly referenced by the current target candidate and validated as same-layer appendix inputs

Candidate-layer dependencies outside the target candidate are not formal pass foundations.
If the target candidate depends on another candidate unit, candidate Rule, candidate appendix outside its own explicit appendix set, or chat-only design decision, the result can only be:

1. `blocked`
   - when the dependency must be stabilized, moved to the correct owner, or extracted into a shared Rule before this check can continue
2. `fix_required`
   - when the current candidate can be repaired immediately by removing the dependency, retargeting it to stable truth, or copying the required decision into the current candidate's owned truth surface

The check must not pass by assuming that an upstream candidate will later become correct.

Before semantic slice review starts, `unit_check` must run the computed candidate relation preflight:

```text
specflowctl relation candidate-preflight --object {unit}
```

If this preflight reports that the target is blocked by another current candidate unit, a candidate Rule, or a candidate progression cycle, `unit_check` must stop before writing any pass gate.
The stop result follows the same boundary as above:

1. use `blocked` when the referenced candidate or candidate Rule must advance, stabilize, or be moved to its owner before the target can pass
2. use `fix_required` when the current target can remove the explicit candidate reference, retarget it to stable truth, or move the dependency statement into a non-blocking evidence-only surface without changing behavior truth

The preflight result is mechanical input only.
It does not judge whether the target candidate is complete, correct, or progressable after the relation blocker is gone.

### 2.6 Slice-Based Closure

Generic slice terms and tooling boundaries are defined by `specflow/framework/slice_work_state_protocol.md`.
This section defines only the `unit_check` slice catalog, dynamic-slice triggers, and closure rule.

`unit_check` must execute through baseline slices, cross-check slices, and any required dynamic slices.
The slice catalog organizes review work; it does not replace the closure standard.

Baseline slices are required unless a slice is explicitly marked `skipped_not_applicable` with a concrete reason:

1. `goal_and_responsibility`
   - checks user goal, unit responsibility, first-round scope, non-goals, and owner fit
2. `dependency_truth_surface`
   - checks formal dependency truth, stable-only dependency foundations, appendix inputs, Rule inputs, repository mapping, and global baseline inputs
3. `main_flow_and_state`
   - checks normal flow, state changes, ordering, and lifecycle semantics inside the candidate
4. `boundary_and_protocol`
   - checks public APIs, ports, adapters, stores, events, trace sources, and cross-unit contracts needed for the flow to run
5. `data_artifact_and_output`
   - checks produced artifacts, evidence, reports, logs, traces, snapshots, persistence records, and their consumers
6. `error_edge_and_gap`
   - checks error states, missing dependencies, fallback stops, diagnostic gaps, and owner handoff for failures
7. `acceptance_and_testability`
   - checks acceptance item structure, test entrypoints, runnable status, proof method, and whether acceptance proves the goal
8. `implementation_handoff`
   - checks whether `unit_plan` can plan implementation from the candidate without filling missing design, adapter, output, or test choices

Cross-check slices are required because locally plausible text can still fail to compose:

1. `goal_to_acceptance_convergence`
   - verifies that accepted items prove the declared goal and responsibility
2. `flow_to_boundary_convergence`
   - verifies that every main-flow step has a matching boundary contract or internal owner
3. `dependency_truth_convergence`
   - verifies that dependency truth actually supports the behavior the candidate relies on
4. `output_to_acceptance_convergence`
   - verifies that output artifacts are directly covered by acceptance and downstream handoff

Dynamic slices are required when the check discovers material review work not fully covered by the baseline slices.
The executor must add a dynamic slice when any of the following appears:

1. a new dependency boundary
2. an owner conflict
3. an uncovered flow node
4. a mechanism-style artifact such as a report, trace, run-state, harness output, or adapter contract
5. an acceptance item that cannot prove the stated target
6. a cross-unit or rule relationship that does not converge
7. any other path where a local slice could pass while the whole logic chain still fails

A dynamic slice can only increase coverage.
It must not replace a baseline slice or allow a baseline slice to be skipped without a concrete not-applicable reason.

`unit_check pass` requires:

1. every baseline slice, cross-check slice, and dynamic slice is `passed` or explicitly `skipped_not_applicable`
2. no slice is `pending`, `blocked`, or `stale`
3. no main logic-chain breakpoint remains
4. no candidate-layer dependency is used as a formal pass foundation
5. `_check_result/unit/{unit}.md` records `slice_summary`, `dependency_truth_result`, and `logic_chain_closure_result`

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=unit_check`
3. the unit has `candidate`
4. read explicitly referenced candidate appendix files and bound Rule files
5. if this round may raise a checkpoint, read `specflow/framework/checkpoint_protocol.md`
6. if referenced appendix files have directory drift, fix that first and rerun the pre-check
7. read `specflow/framework/onboarding_decision_policy.md`
8. read `specflow/framework/candidate_intent_policy.md` and the selected intent standard for the current candidate
9. create or refresh `docs/specs/_check_work/unit/{unit}.md` before semantic slice review
10. validate the work-state shape before using it as a resume aid
11. run `specflowctl relation candidate-preflight --object {unit}` before semantic slice review

## 4. Procedure

1. read `docs/specs/units/candidate/c_unit_{unit}.md` plus all required appendix and Rule files
2. if `stable` exists, also read `docs/specs/units/stable/s_unit_{unit}.md` plus required stable appendix files
3. read `docs/specs/rules/stable/s_g_rule_repository_baseline.md` if it exists; otherwise continue with the "no formal global baseline yet" state
4. initialize or refresh `_check_work/unit/{unit}.md`
   - run `process check-work-init` if no current open work-state exists
   - run `process check-work-refresh` before continuing an existing open work-state
   - stop if the work-state reuse age requires an explicit reuse-or-restart decision
5. run `specflowctl relation candidate-preflight --object {unit}`
   - if `may_continue=false`, stop before semantic closure review
   - report `ready_candidates`, `blocked_by`, `candidate_cycles`, and the source files named by the relation result when they exist
   - do not write `_check_result/unit/{unit}.md`
6. read `candidate_intent` from the candidate frontmatter and apply `specflow/framework/candidate_intent_policy.md`
   - missing or unknown `candidate_intent` is `fix_required`
   - missing selected intent standard is a command blocker
   - intent-specific closure criteria are owned by the selected standard file, not by repeated local rules in this command file
7. judge `progressability`
8. judge `content completeness`
9. classify completeness gaps into:
   - `critical`
   - `important`
   - `elaboration`
10. use these fixed completeness review objects:
   - `Behavior Basis Completeness`
   - `Decision Surface Completeness`
   - `Acceptance Basis Completeness`
   - `Content Organization Completeness`
     - the candidate main Spec and its appendix files must satisfy `spec_writing_guide.md` Section 6
     - explanatory and normative content must be separated at the subsection level
     - mixed-paragraph violations are an `important` completeness gap by default
11. complete `Candidate Design Quality` review as part of the framework baseline:
   - the candidate must connect the current user or actor goal to the behavior being proposed
   - the candidate must define the first-round scope and non-goals clearly enough that future capabilities are not silently implemented now
   - the candidate must record one current selected direction when multiple solution options were discussed
   - the candidate must define acceptance criteria that can prove the result is useful, not only that artifacts exist
   - the candidate must not depend on chat context, guidance discussion, README vision, or rejected alternatives for implementation-critical meaning
   - over-broad scope, unresolved direction, unverifiable success, or chat-dependent behavior truth can only result in `blocked` or `fix_required`
12. execute every baseline slice from Section 2.6
   - record the slice status in `_check_work`
   - use `passed` only when the slice's review question is answered from formal inputs
   - use `skipped_not_applicable` only with a concrete reason in `result_summary`
   - use `blocked` when the slice exposes a blocker that prevents a pass conclusion
13. add dynamic slices whenever Section 2.6 requires them, and finish those dynamic slices before final gate judgment
14. execute every cross-check slice from Section 2.6 after the local baseline slices it depends on have been reviewed
15. perform Logic Chain Closure from Section 2.4
16. perform Dependency Truth Boundary from Section 2.5
17. complete the framework-baseline closure checks owned by `unit_check`, including the fixed completeness review objects, `Candidate Design Quality`, baseline, rule, candidate intent, rule-truth checks, slice closure checks, dependency boundary checks, and logic-chain checks
18. process formal global baseline alignment and any candidate-carried global rule proposal:
   - if the formal global baseline exists and the candidate is still compatible, a mechanical update to the current version is allowed
   - if incompatible, the result can only be `blocked` or `fix_required`
   - if a global rule proposal is present, it must clearly state the proposed global rule delta, the reason the current baseline is insufficient, the unit-local implementation/verification impact, and the affected units or rules
   - if those fields are unclear, the result can only be `blocked` or `fix_required`
19. process `rule_refs`:
   - if current behavior depends on Rule truth but bindings are missing or incomplete, the result can only be `blocked` or `fix_required`
   - if bindings exist but the body does not explain which behavior chain reuses them, the result can only be `blocked` or `fix_required`
20. process candidate intent and source fields:
   - `candidate_intent` must be one of the values allowed by `candidate_intent_policy.md`
   - apply the selected intent standard to determine whether `repair_basis`, `Repair Scope`, stable behavior preservation, behavior delta, and evidence requirements are valid for this round
   - `source_basis` must be one of `new_design`, `existing_implementation`, `mixed`, or `replacement`
   - `evidence_appendix_ref` must be present
   - if `source_basis=existing_implementation` or `source_basis=mixed`, `evidence_appendix_ref` must name an existing candidate evidence appendix and that appendix must be read
   - if `source_basis=new_design` or `source_basis=replacement`, `evidence_appendix_ref` must be `none`
   - evidence appendix conflicts or unknowns that still affect selected candidate behavior are critical completeness gaps unless the candidate main Spec explicitly makes a bounded selected rule that no longer depends on them
   - evidence appendix text must not be treated as implementation truth; only the candidate main Spec and bound formal truth may constrain implementation
22. process explicit acceptance items:
   - the candidate must contain a `Testability / Acceptance Criteria` section, or an explicitly equivalent acceptance section title
   - each acceptance item must record `id`, `target`, `verification_surface`, `implementation_surface`, `verification_method`, and `pass_condition`
   - `verification_surface` must use only the fixed values from `spec_writing_guide.md` Section 6
   - vague acceptance text such as "works", "can be replaced", "aligns with design", "supports integration", or equivalent result-only wording is not sufficient unless the required fields make the item directly verifiable
   - for `verification_surface=public_api`, the item must name the public package or exported contract surface, describe an external-consumer style verification method, and require a pass condition that does not import `internal` packages
   - for `verification_surface=integration`, the item must name the runnable integration entrypoint or mark the item as `not_runnable_yet` with a concrete missing-entrypoint reason
   - if an item is marked `not_runnable_yet`, `unit_check` may treat the item as explicitly bounded only when the reason is concrete and the candidate does not use that same item as a current pass claim
   - if a current-gate acceptance item is missing, vague, structurally incomplete, or falsely implied to pass while marked non-runnable, the result can only be `blocked` or `fix_required` with `fallback_reason_code=truth_incomplete`
23. process shared-candidate signals:
   - by default, shared-candidate hints only trigger a suggestion to enter natural-language rule governance
   - if the current required reading range already confirms a dual source of truth, report it directly as a blocking issue with `fallback_reason_code=shared_truth_conflict`
24. determine whether a blocking checkpoint is the correct stop form:
   - use `clarification` when user intent, boundary meaning, or acceptance meaning is still missing from truth
   - use `decision` when multiple materially different directions remain and the user must choose one
25. checkpoint rules:
   - a checkpoint is not `pass`
   - if a checkpoint conclusion changes behavior truth, it must be written back to candidate or appendix before `unit_check` may be rerun
   - do not write `_check_result/unit/{unit}.md` for checkpoint-only stops
26. merge conclusions in this order:
   - `progressability`
   - `content completeness`
   - `Candidate Design Quality`
   - `Dependency Truth Boundary`
   - `Logic Chain Closure`
   - slice closure
   - overall gate conclusion
27. merge rules:
   - if `progressability` fails -> only `blocked` or `fix_required`
   - if any `critical` completeness gap exists -> only `blocked` or `fix_required`
   - if `Candidate Design Quality` fails on scope, selected direction, acceptance usefulness, or chat-dependent truth -> only `blocked` or `fix_required`
   - if a candidate-layer dependency is used as a formal foundation -> only `blocked` or `fix_required`
   - if the main logic chain has a breakpoint -> only `blocked` or `fix_required`
   - if any required slice is `pending`, `blocked`, or `stale` -> only `blocked` or `fix_required`
   - if only `important` or `elaboration` issues remain, `pass` is still possible
28. before writing a pass gate, run `process check-work-refresh` and `process check-work-validate`
   - no required slice may remain stale
   - every required baseline, cross-check, and dynamic slice must be closed
   - if refresh marks a passed slice stale, stop without writing `_check_result`
29. if the result is `pass`, create or update `docs/specs/_check_result/unit/{unit}.md`
   - write the accepted acceptance-item set into the pass gate summary by item `id`, `verification_surface`, and `not_runnable_yet` state
   - write `slice_summary`, `dependency_truth_result`, and `logic_chain_closure_result`
   - then run `snapshot validate-process --object-type unit --object {unit} --process check`
30. if the result is not `pass`, do not write a failed `_check_result/unit/{unit}.md`; delete an old pass gate if it is no longer valid
31. if the result is `blocked` or `fix_required`, close the current `unit_check` run after writing any required findings:
   - any later truth repair belongs to follow-up work, not to a still-open `unit_check`
   - any later repair-side reassessment or scoped follow-up review remains non-authoritative unless a new fresh full-scope `unit_check` run is entered through command routing
   - report the blocker, owner, recommended repair location, and resume signal
32. update `_status.md`:
   - if pass -> `Next Command=unit_plan`
   - otherwise -> `Next Command=unit_check`
   - the deterministic command closure may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_check --object-type unit --object {unit} --outcome <pass|blocked|fix_required|checkpoint> --notes <status-note> --apply`

## 5. Stop Conditions

1. whether the candidate satisfies both `progressability` and `content completeness` is clear
2. whether `Candidate Design Quality` passes or blocks the candidate is clear
3. if the round passes, `_check_result/unit/{unit}.md` holds the pass gate
4. if the round does not pass, no invalid old pass gate remains
5. `_status.md` is updated
6. the explicit acceptance-item set is either accepted as specific enough for downstream planning or reported as a blocking truth gap
7. no repair-side reassessment or scoped follow-up review has been mistaken for a formal `unit_check pass`
8. the work-state file has no stale required slice before a pass gate is written
9. the candidate relation preflight has either passed or has been reported as the reason no pass gate was written
10. blocked slices report blocker, owner, recommended repair location, and resume signal

## 6. Output Contract

The output should include:

1. overall conclusion
2. severity summary
3. formal global baseline alignment result
4. candidate source and evidence appendix result
5. candidate intent result
6. acceptance-item completeness result, including any vague, missing, or non-runnable item findings
7. slice summary:
   - baseline slice statuses
   - cross-check slice statuses
   - dynamic slice statuses and why each dynamic slice was added
8. dependency truth result:
   - stable unit inputs
   - stable Rule inputs
   - repository mapping and global baseline inputs
   - any rejected candidate-layer dependency
9. logic chain closure result:
   - goal
   - responsibility
   - main flow
   - boundary protocol
   - output artifact
   - acceptance
   - handoff
10. the gate conclusion set:
   - `progressability`
   - `content completeness`
   - `Candidate Design Quality`
   - overall gate conclusion
11. whether `Check Result Snapshot` was written back or an old gate was cleaned up
12. `checkpoint result` when a checkpoint stop was raised
   - when present, it must satisfy the fixed checkpoint fields defined by `specflow/framework/checkpoint_protocol.md`
13. `fallback_reason_code` for blocked, fix-required, or checkpoint stops
14. candidate relation preflight result when it blocks the command
14. structured findings when `blocked` or `fix_required`
15. next-step suggestion
16. `_status.md` update result
17. when follow-up work only confirmed local repair or ran a scoped review instead of a new formal rerun, that this result was non-authoritative and did not change lifecycle state
18. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - when a checkpoint was raised, also report `resume signal`
   - if `Next Command=unit_check`, `why this next step` must explicitly state whether the blocker is truth repair, user clarification, or a required decision rather than only repeating that closure is incomplete

When the result is `blocked` or `fix_required`, findings must be structured and must not be reduced to vague summaries.

Severity must use the shared meanings defined in:

1. `specflow/framework/severity_policy.md`

Each finding must explain:

1. background
2. what happened
3. impact
4. best recommendation
5. why that recommendation is best
6. whether it is blocking
7. which constraint layer it belongs to

Allowed checkpoint types:

1. `clarification`
2. `decision`

Allowed `fallback_reason_code` values:

1. `truth_incomplete`
2. `baseline_drift`
3. `rule_drift`
4. `shared_truth_conflict`
5. `governance_drift`

Candidate source-field and evidence-appendix blockers must use `truth_incomplete`.
The output must name the exact source or evidence condition in the natural-language explanation after the standardized code.
Do not introduce source-basis-specific or evidence-appendix-specific `fallback_reason_code` values in this command.

## 7. Non-Goals

1. directly generating a plan
2. directly entering code implementation
3. creating, updating, or deleting an independent stable `g_` rule candidate file

## 8. Examples

```md
unit_check:ai
```
