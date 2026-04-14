# Candidate Promote Command

## 1. Purpose

This command promotes the specified module's `candidate` into the new `stable`.

## 2. Scope

By default it handles:

1. promoting the candidate version into the formal version
2. updating state files
3. cleaning this round's candidate and process files
4. updating `s_system_constraints.md` when needed

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=cand_promote`
3. a latest valid `_verify_result/{module}.md` still covers the current candidate, current implementation, and current formal global baseline state
4. implementation alignment is complete and no blocking verification issue remains
5. the candidate's `system_constraints_stable_ref` matches the current formal global baseline state
6. read required candidate appendix files and bound Shared Appendix files, and decide how each one will be handled after promotion
7. read the git policy before promotion

## 4. Procedure

1. read and re-check the latest `_verify_result/{module}.md`
2. read `docs/specs/candidate/c_{module}.md` and all required appendix files
3. validate the full binding relation of `_verify_result/{module}.md`
4. if `_verify_result/{module}.md` is invalid, identify the reason and stop immediately:
   - if code changed after verification -> fall back to `cand_verify`
   - if implementation drift against candidate exists -> fall back to `cand_impl`
   - if candidate truth or formal global baseline changed -> fall back to `cand_check`
5. continue only when bindings, coverage, and gate fields all remain valid
6. confirm that candidate `frontmatter.version` is the new `stable` version for this round
7. if `promotion_to_system_stable=with_module`, absorb `proposed_system_constraints_updates` into `docs/specs/system/stable/s_system_constraints.md`
8. if `shared_appendix_refs` is not empty, make a forced decision for each bound shared item:
   - migrate to `docs/specs/shared/stable/s_shared_xxx.md`
   - absorb the stable conclusion into `s_system_constraints.md`
   - absorb the stable conclusion into module `stable` and delete the shared appendix file
   - if none of those can be completed now, stop promotion
9. generate or update `docs/specs/stable/s_{module}.md`
10. if current-round candidate appendix files exist, in the same promotion round either:
   - migrate retained content to `docs/specs/stable/appendix/` or an equivalent dedicated subdirectory
   - absorb the content into `docs/specs/stable/s_{module}.md`
   - delete candidate appendix files no longer needed
11. do not delete `docs/specs/candidate/c_{module}.md` until `_status.md` has already been updated to `Candidate=no`
12. update `_status.md` to the promoted stable state
13. only after that update may physical deletion happen:
   - `docs/specs/candidate/c_{module}.md`
   - current-round candidate appendix files
   - `_check_result/{module}.md`
   - `_plans/{module}.md`
   - `_verify_result/{module}.md`
14. if the command is interrupted after promotion internals started but before final cleanup finished, treat the repository as an "incomplete promotion recovery state" and recover to candidate semantics instead of claiming success
15. if other modules were affected by Shared Appendix changes but not directly closed here, run `shared_flow_reconcile`
16. perform git close-out if required

## 5. Stop Conditions

1. promotion succeeded or a blocking reason is explicit
2. `_status.md` is updated
3. this round's candidate cleanup is complete
4. if verification became invalid, the command stopped and `_status.md` fell back appropriately
5. if the command entered incomplete-promotion recovery state, candidate semantics were restored and the module can restart from `cand_check`

## 6. Output Contract

1. promotion conclusion
2. formal version confirmation result
3. file and state update result
4. `system_constraints` linked-promotion result
5. cleanup result
6. fallback reason if verification became invalid
7. recovery-state explanation if incomplete promotion occurred
8. git close-out result
9. follow-up state explanation

## 7. Non-Goals

1. redesigning `candidate`
2. replacing implementation verification with promotion
3. automatically opening the next candidate round
4. maintaining an independent `system_constraints` candidate file
