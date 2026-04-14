# Candidate Plan Command

## 1. Purpose

This command creates or updates the implementation plan for the current candidate.

## 2. Scope

By default it handles:

1. reading the current valid candidate pass gate
2. generating or updating `_plans/{module}.md`
3. keeping plan bindings aligned with the current candidate, current formal global baseline state, and current Shared Appendix snapshot

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=cand_plan`
3. a current valid `docs/specs/_check_result/{module}.md` exists
4. the current candidate still aligns with the current formal global baseline state
5. read any explicitly referenced candidate appendix files and bound Shared Appendix files
6. read the git policy if commit-triggering files may change

## 4. Procedure

1. read the candidate Spec and all required appendix or Shared Appendix files
2. read `s_system_constraints.md` if it exists
3. read `docs/specs/_check_result/{module}.md`
4. verify the pass gate bindings are still valid
5. if the pass gate is invalid, stop immediately and fall back `_status.md` to `cand_check`
6. create or update `docs/specs/_plans/{module}.md`
7. ensure the plan records:
   - implementation tasks
   - progress, blockers, and verification focus for this round
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
8. update `_status.md`:
   - if the candidate is now ready for implementation -> `Next Command=cand_impl`
   - if candidate truth drift was discovered -> `Next Command=cand_check`
9. perform git close-out if required

## 5. Stop Conditions

1. the plan file exists and is bound to the current candidate truth
2. `_status.md` points to the real next step

## 6. Output Contract

1. planning conclusion
2. plan file path
3. plan binding result
4. any fallback reason if the pass gate was invalid
5. git close-out result
6. `_status.md` update result

## 7. Non-Goals

1. direct code implementation
2. direct verification
3. rewriting candidate truth
