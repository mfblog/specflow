# Candidate Plan Command

## 1. Purpose

This command creates or updates the implementation plan for the current candidate.

## 2. Scope

By default it handles:

1. reading the current valid candidate pass gate
2. generating or updating `_plans/{module}.md`
3. keeping plan bindings aligned with the current candidate, current formal global baseline state, and current Shared Appendix snapshot
4. stopping at a structured decision checkpoint only when key implementation direction is still not locked

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=cand_plan`
3. a current valid `docs/specs/_check_result/{module}.md` exists
4. the current candidate still aligns with the current formal global baseline state
5. read any explicitly referenced candidate appendix files and bound Shared Appendix files
6. if this round may raise a checkpoint, read `specflow/framework/docs/agent_guidelines/checkpoint_protocol.md`
7. read the git policy if commit-triggering files may change

## 4. Procedure

1. read the candidate Spec and all required appendix or Shared Appendix files
2. read `s_system_constraints.md` if it exists
3. read `docs/specs/_check_result/{module}.md`
4. verify the pass gate bindings are still valid
5. if the pass gate is invalid, stop immediately and fall back `_status.md` to `cand_check`
6. determine whether a `decision` checkpoint is required:
   - only use it when key implementation direction is still not locked
   - if the unresolved decision changes behavior truth, boundary truth, or acceptance truth, the resume path must go back to `cand_check` after writeback
   - do not treat the checkpoint as permission to continue without that writeback
7. create or update `docs/specs/_plans/{module}.md`
8. ensure the plan records:
   - implementation tasks
   - progress, blockers, and verification focus for this round
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
9. update `_status.md`:
   - if the candidate is now ready for implementation -> `Next Command=cand_impl`
   - if candidate truth drift was discovered -> `Next Command=cand_check`
10. perform git close-out if required

## 5. Stop Conditions

1. the plan file exists and is bound to the current candidate truth
2. `_status.md` points to the real next step

## 6. Output Contract

1. planning conclusion
2. plan file path
3. plan binding result
4. `handoff validation result`
5. `checkpoint result` when a checkpoint stop was raised
   - when present, it must satisfy the fixed checkpoint fields defined by `specflow/framework/docs/agent_guidelines/checkpoint_protocol.md`
6. `fallback_reason_code` for fallback or checkpoint stops
7. any fallback reason if the pass gate was invalid
8. git close-out result
9. `_status.md` update result

Allowed checkpoint types:

1. `decision`

Allowed `fallback_reason_code` values:

1. `gate_missing`
2. `truth_drift`
3. `binding_drift`
4. `baseline_drift`
5. `shared_appendix_drift`
6. `truth_incomplete`

## 7. Non-Goals

1. direct code implementation
2. direct verification
3. rewriting candidate truth
