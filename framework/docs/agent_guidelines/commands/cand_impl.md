# Candidate Implement Command

## 1. Purpose

This command advances code implementation according to the current `candidate` and `_plans/{module}.md`.

## 2. Scope

By default it handles:

1. implementing according to plan
2. adding necessary tests or verification actions
3. writing progress back into `_plans/{module}.md`
4. consuming the `cand_plan -> cand_impl` handoff only when gate and plan bindings both still hold

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=cand_impl`
3. a current valid `docs/specs/_check_result/{module}.md` exists
4. a current valid `docs/specs/_plans/{module}.md` exists
5. the candidate still aligns with the current formal global baseline state
6. read required candidate appendix files and bound Shared Appendix files
7. read the git policy before implementation work

## 4. Procedure

1. read the candidate Spec and all required appendix or Shared Appendix files
2. read `s_system_constraints.md` if it exists
3. read the current `_check_result/{module}.md`
4. read the current `_plans/{module}.md`
5. validate all required bindings of the pass gate and plan file according to the candidate handoff contract
6. if any binding is invalid, stop immediately and fall back `_status.md` to `cand_check`
7. if `system_constraints_stable_ref` no longer matches the current formal global baseline state, stop immediately and fall back to `cand_check`
8. only when both pass gate and plan are still valid may implementation continue
9. implement in the order defined by the plan
10. run necessary verification, or record clearly what could not be run
11. write completion status, blockers, and verification results back into `_plans/{module}.md`
12. update `_status.md`:
   - if implementation is ready for verification -> `Next Command=cand_verify`
   - if implementation is still blocked -> keep `Next Command=cand_impl`
   - if candidate truth or formal global baseline drift means closure must restart -> `Next Command=cand_check`
13. perform git close-out if required

## 5. Stop Conditions

1. the plan has advanced as far as feasible in this round
2. the plan file has been written back
3. `_status.md` points to the real next executable step
4. if the pass gate or plan became invalid, implementation was stopped and `_status.md` was fallen back to `cand_check`

## 6. Output Contract

1. implementation progress result
2. tests or verification run, or explicit gaps
3. plan write-back result
4. `handoff validation result`
5. `fallback_reason_code` when the pass gate or plan was invalid
6. fallback reason if the pass gate or plan was invalid
7. git close-out result
8. `_status.md` update result

Allowed checkpoint types:

1. none

Allowed `fallback_reason_code` values:

1. `gate_missing`
2. `truth_drift`
3. `binding_drift`
4. `baseline_drift`
5. `shared_appendix_drift`

## 7. Non-Goals

1. rewriting candidate truth
2. replacing verification with implementation
3. advancing an independent `system_constraints` state machine
