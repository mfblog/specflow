# Candidate Promote Command

## 1. Purpose

This command promotes the specified module's `candidate` into the new `stable`.

## 2. Scope

By default it handles:

1. promoting the candidate version into the formal version
2. updating state files
3. cleaning this round's candidate and process files
4. updating `s_system_constraints.md` when a closed module-carried global proposal is ready
5. consuming the `cand_verify -> cand_promote` handoff only when verification still covers the current round

## 3. Preconditions

1. complete required pre-checks
2. `_status.md` says `Next Command=cand_promote`
3. a latest valid `_verify_result/{module}.md` still covers the current candidate, current implementation, and current formal global baseline state
4. implementation alignment is complete and no blocking verification issue remains
5. the candidate's `system_constraints_stable_ref` matches the current formal global baseline state
6. read required candidate appendix files and bound Shared Contract files, and decide how each one will be handled after promotion
7. read `specflow/framework/docs/agent_guidelines/recovery_policy.md` before promotion
8. if the round may create, update, or delete any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, read `specflow/framework/docs/agent_guidelines/shared_sync.md` before promotion
9. read the git policy before promotion

## 4. Procedure

1. read and re-check the latest `_verify_result/{module}.md`
2. read `docs/specs/candidate/c_{module}.md` and all required appendix files
3. validate the full binding relation of `_verify_result/{module}.md` according to the candidate handoff contract
4. if `_verify_result/{module}.md` is invalid, identify the reason and stop immediately:
   - if code changed after verification:
     - delete `_verify_result/{module}.md`
     - fall back to `cand_verify`
   - if implementation drift against candidate exists:
     - delete `_verify_result/{module}.md`
     - fall back to `cand_impl`
   - if another required binding of `_verify_result/{module}.md` no longer matches the current round:
     - delete `_check_result/{module}.md`
     - delete `_plans/{module}.md`
     - delete `_verify_result/{module}.md`
     - use `fallback_reason_code=binding_drift` and fall back to `cand_check`
   - if bound Shared Contract truth, layer, version, or snapshot drifted:
     - delete `_check_result/{module}.md`
     - delete `_plans/{module}.md`
     - delete `_verify_result/{module}.md`
     - use `fallback_reason_code=shared_contract_drift` and fall back to `cand_check`
   - if candidate truth or formal global baseline changed:
     - delete `_check_result/{module}.md`
     - delete `_plans/{module}.md`
     - delete `_verify_result/{module}.md`
     - fall back to `cand_check`
5. continue only when bindings, coverage, and gate fields all remain valid
6. before the first file mutation, capture the recovery baseline required by `recovery_policy.md`
7. confirm that candidate `frontmatter.version` is the new `stable` version for this round
8. if the module candidate contains a closed `system_constraints_change_proposal` that this round has implemented and verified, absorb the promoted conclusion into `docs/specs/system/stable/s_system_constraints.md`
9. if `shared_contract_refs` is not empty, decide for each bound shared item:
   - if it should remain an independent cross-module truth after promotion, promote it into `docs/specs/shared_contracts/stable/`
   - if part of its conclusion has become a project-wide default rule, also absorb that specific conclusion into `s_system_constraints.md`
   - do not absorb a Shared Contract into module `stable` merely because promotion happened
   - do not treat promotion itself as a reason to delete a still-needed Shared Contract
   - if the required post-promotion truth shape is still unclear, stop promotion
10. generate or update `docs/specs/stable/s_{module}.md`
11. if current-round candidate appendix files exist, in the same promotion round either:
   - migrate retained content to `docs/specs/stable/appendix/` or an equivalent dedicated subdirectory
   - absorb the content into `docs/specs/stable/s_{module}.md`
   - delete candidate appendix files no longer needed
12. do not delete `docs/specs/candidate/c_{module}.md` until `_status.md` has already been updated to `Candidate=no`
13. update `_status.md` to the promoted stable state:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=spec_fork`
14. only after that update may physical deletion happen:
   - `docs/specs/candidate/c_{module}.md`
   - current-round candidate appendix files
   - `_check_result/{module}.md`
   - `_plans/{module}.md`
   - `_verify_result/{module}.md`
15. if the command is interrupted after promotion internals started but before final cleanup finished, run incomplete promotion recovery according to `recovery_policy.md` instead of claiming success
16. if the round changed any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, run `shared_sync` only after `_status.md` already reflects the promoted stable layer, even when no additional affected module is known yet
17. perform git close-out if required

## 5. Stop Conditions

1. promotion succeeded or a blocking reason is explicit
2. `_status.md` is updated to:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=spec_fork`
3. this round's candidate cleanup is complete
4. if verification became invalid, the command stopped and `_status.md` fell back appropriately
5. if the command entered incomplete-promotion recovery state, candidate semantics were restored and the module can restart from `cand_check`

## 6. Output Contract

1. promotion conclusion
2. formal version confirmation result
3. file and state update result
4. `system_constraints` linked-promotion result
5. Shared Contract reconciliation result when the round changed shared truth or bindings
6. cleanup result
7. `handoff validation result`
8. fallback cleanup result when verification became invalid before promotion could start
9. `fallback_reason_code` if verification became invalid
10. fallback reason if verification became invalid
11. `fallback_reason_code=promotion_recovery` when incomplete promotion recovery occurred
12. recovery-state explanation if incomplete promotion occurred
13. git close-out result
14. follow-up state explanation
   - when promotion succeeds, the follow-up state must explicitly confirm:
     - `Stable=yes`
     - `Candidate=no`
     - `Active Layer=stable`
     - `Next Command=spec_fork`

Allowed checkpoint types:

1. none

Allowed `fallback_reason_code` values:

1. `truth_drift`
2. `binding_drift`
3. `baseline_drift`
4. `shared_contract_drift`
5. `implementation_deviation`
6. `evidence_incomplete`
7. `promotion_recovery`

## 7. Non-Goals

1. redesigning `candidate`
2. replacing implementation verification with promotion
3. automatically opening the next candidate round
4. maintaining an independent `system_constraints` candidate file
