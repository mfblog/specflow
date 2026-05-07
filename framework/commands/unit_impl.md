# Unit Implement Command

## 1. Purpose

This command advances code implementation according to the current `candidate` and `_plans/active/{unit}.md`.

## 2. Scope

By default it handles:

1. implementing according to plan slices
2. adding necessary tests or verification actions
3. dynamically confirming which legacy dependencies have already stopped being required
4. writing progress back into `_plans/active/{unit}.md`
5. consuming the `unit_plan -> unit_impl` handoff only when gate and plan bindings both still hold
6. preserving the active plan's acceptance item coverage while writing implementation progress

### 2.1 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_impl`-local entry, output, and stop rules.

Process-file consumption and writeback for `_check_result/unit/{unit}.md` and `_plans/active/{unit}.md` must follow `specflow/framework/process_snapshot_contract.md` Section 9. When deterministic snapshot validation tooling is available for the current process kind, the matching `snapshot validate-process` command is the mandatory tool-backed validation step before treating either process file as consumable, reporting the implementation handoff as valid, or advancing lifecycle state.

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=unit_impl`
3. a current valid `docs/specs/_check_result/unit/{unit}.md` exists
4. a current valid `docs/specs/_plans/active/{unit}.md` exists
5. the candidate still aligns with the current formal global baseline state
6. read required candidate appendix files and bound Rule files

## 4. Procedure

1. read the candidate Spec and all required appendix or Rule files
2. read `s_g_rule_repository_baseline.md` if it exists
3. read the current `_check_result/unit/{unit}.md`
4. read the current `_plans/active/{unit}.md`
5. validate all required bindings of the pass gate and plan file according to the candidate handoff contract
   - this includes validating that the active plan still covers the current candidate acceptance item `id` set
6. when a required appendix is an evidence appendix, treat it only as reviewed evidence covered by the pass gate; it must not supply implementation requirements, acceptance criteria, or behavior rules
7. if handoff validation fails, classify the failure through `recovery_policy.md` Section 4 before cleanup:
   - if the check gate no longer covers current truth or bindings, use `truth_layer`, delete the unit candidate-side process chain, and fall back `_status.md` to `unit_check`
   - if only the active plan is missing, malformed, not tool-valid, or missing acceptance coverage while the check gate still covers current truth, use `plan_layer`, delete `_plans/draft/{unit}.md`, `_plans/active/{unit}.md`, and `_verify_result/unit/{unit}.md` if present, then set `_status.md` to `unit_plan`
   - if only the check gate process shape is malformed while current truth and bindings still match, use `gate_layer`, delete `_check_result/unit/{unit}.md`, and set `_status.md` to `unit_check`
8. only when both pass gate and plan are still valid may implementation continue
9. implement slice by slice in the order defined by the current plan unless the plan itself declares a dependency-safe different order
10. for each slice, use the recorded objective, file scope, dependencies, verification action, and done condition as the execution boundary
11. do not collapse a blocked slice into a vague whole-unit status; record clearly which slice is complete, blocked, or still pending
12. treat a slice as truly advanced only when at least one of the following is now true:
   - an execution surface has been cut over to its target path
   - a named retirement target is now confirmed as no longer required
13. when implementation discovers additional legacy paths, legacy helpers, legacy patches, or legacy wrappers that were not yet fully modeled:
   - keep the work in `unit_impl` if the discovery only deepens implementation facts
   - write the discovery back into the current active plan under the round's implementation-progress sections
   - do not restart the round from `unit_plan` unless the discovery proves candidate truth itself is insufficient
14. if implementation discovers that the active plan's convergence target cannot stand without a new behavior or boundary decision:
   - stop treating the issue as implementation-only
   - fall back to `unit_check`
15. run necessary verification for the slices advanced in this round, or record clearly what could not be run
16. write slice completion status, blockers, verification results, and retirement progression back into `_plans/active/{unit}.md`
17. ensure the active plan write-back records at minimum:
   - `Takeover Progress`
   - `Retirement Progress`
   - `Acceptance Item Progress` for any acceptance item affected by advanced slices
   - `Newly Confirmed Legacy`
   - `Residual Legacy Dependencies`
   - for each advanced slice: `execution_surface`, `cutover_result`, `retirement_result`, and `verification_note`
18. update `_status.md`:
   - if implementation is ready for verification -> `Next Command=unit_verify`
   - if implementation is still blocked -> keep `Next Command=unit_impl`
   - if candidate truth or formal global baseline drift means closure must restart -> `Next Command=unit_check`

## 5. Stop Conditions

1. the current plan slices have advanced as far as feasible in this round, and each claimed advance names either a target-path cutover or a retirement confirmation
2. the plan file has been written back
3. `_status.md` points to the real next executable step
4. if the pass gate or plan became invalid, or implementation discovered that candidate truth itself must be re-closed, implementation was stopped and `_status.md` was fallen back to `unit_check`

## 6. Output Contract

1. implementation progress result
2. slice progress result
3. tests or verification run, or explicit gaps
4. takeover-progress result
5. retirement-progress result
6. plan write-back result
7. blocked-slice result when implementation could not finish the current plan round
8. `handoff validation result`
9. cleanup result when implementation stopped through layered recovery
10. `fallback_reason_code` and `failure_layer` when implementation stopped through layered recovery
11. fallback reason when implementation stopped through layered recovery
12. `_status.md` update result
13. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - `current state` must explicitly confirm the written `Active Layer` and `Next Command`
   - if `Next Command=unit_impl`, `why this next step` must explicitly state that implementation progressed but candidate closure has not yet reached the `unit_verify` entry condition
   - `next-stage entry gap` must name the unfinished implementation, verification, closure, or retirement surfaces that still block `unit_verify`

Allowed checkpoint types:

1. none

Allowed `fallback_reason_code` values:

1. `gate_missing`
2. `truth_drift`
3. `binding_drift`
4. `baseline_drift`
5. `rule_drift`
6. `truth_incomplete`
7. `gate_layer`
8. `plan_layer`

## 7. Non-Goals

1. rewriting candidate truth
2. replacing verification with implementation
3. advancing an independent stable `g_` rule state machine
