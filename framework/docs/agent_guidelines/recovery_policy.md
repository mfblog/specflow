# Recovery Policy

## 1. Purpose

This file defines the centralized recovery rules used when a Spec Flow command cannot safely continue after state mutation or handoff invalidation.

It answers five questions:

1. what "recovery" means in Spec Flow
2. which recovery cases are centralized here
3. how incomplete promotion recovery must work
4. which repository state may be claimed after recovery
5. which standardized `fallback_reason_code` applies

This file does not replace command-local stop conditions.
It defines the shared recovery baseline that commands must follow.

---

## 2. Scope

This policy covers two recovery classes:

1. candidate-side recovery after handoff invalidation
2. incomplete promotion recovery after `cand_promote` has already started mutating repository state

Boundary:

1. normal smallest-step fallback still comes from:
   - `specflow/framework/docs/agent_guidelines/candidate_handoff_contract.md`
   - the active command file
2. this policy adds the shared repository-state recovery rules that commands must use while performing that fallback

---

## 3. Core Terms

### 3.1 Recovery

`recovery` means:

1. stop claiming that the current command has completed successfully
2. restore the repository to a state that the next legal command can safely consume
3. update `_status.md` so the next step is explicit

### 3.2 Recovery Baseline

`recovery baseline` means the exact pre-mutation snapshot a command keeps so it can restore files if the command is interrupted mid-round.

It is not:

1. a module Spec
2. a process file
3. a new lifecycle stage

### 3.3 Incomplete Promotion Recovery

`incomplete promotion recovery` means the special recovery path used when `cand_promote` has already begun mutating files for promotion, but the round cannot be safely closed as a completed promotion.

Use standardized `fallback_reason_code=promotion_recovery`.

---

## 4. Candidate-Side Recovery Baseline

For candidate-side commands that only consume upstream artifacts and then discover drift:

1. use the smallest fallback step defined by the active command and the candidate handoff contract
2. delete or invalidate the outdated process files required by that fallback
3. update `_status.md` to that smallest actionable step
4. do not invent extra recovery states

Plain meaning:

1. ordinary candidate-chain drift recovery usually needs cleanup plus `_status.md` fallback
2. it does not need a second state machine

---

## 5. Incomplete Promotion Recovery

### 5.1 Required Recovery Baseline Before Mutation

Before `cand_promote` makes its first file mutation, it must capture a recovery baseline covering every file the round may overwrite or delete, including at minimum:

1. the current module row in `docs/specs/_status.md`
2. `docs/specs/candidate/c_{module}.md`
3. current-round candidate appendix files for that module
4. current-round `_check_result/{module}.md`
5. current-round `_plans/{module}.md`
6. current-round `_verify_result/{module}.md`
7. `docs/specs/stable/s_{module}.md` if it already existed before promotion
8. `docs/specs/system/stable/s_system_constraints.md` if the round may mutate it
9. any Shared Appendix files the round may mutate, promote, absorb, or delete

Rules:

1. the baseline must preserve exact file bytes after read
2. the baseline may live in memory or in a temporary executor-owned artifact
3. it must remain available until promotion success is fully closed
4. it is not a formal repository artifact and should not be committed

### 5.2 When Recovery Is Required

Incomplete promotion recovery is required when both are true:

1. `cand_promote` has already mutated at least one promotion target file or deletion target file
2. the command cannot still complete the full promotion round safely

Examples:

1. the command was interrupted after writing the new `stable`
2. shared-file handling became blocked after some promotion writes already happened
3. cleanup started but did not finish

### 5.3 Recovery Procedure

When incomplete promotion recovery is triggered, the command must:

1. stop claiming promotion success
2. restore every mutated or deleted file covered by the recovery baseline back to its exact pre-mutation content or existence state
3. delete any new file created by this interrupted promotion round that did not exist in the recovery baseline
4. restore the module row in `docs/specs/_status.md` to candidate semantics:
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=cand_check`
5. keep `Stable=yes|no` consistent with the pre-round module state from the recovery baseline
6. if the round touched Shared Appendix files or `s_system_constraints.md`, restore those files before claiming recovery complete
7. after repository restoration, treat existing candidate-side process files as no longer safe for reuse and delete:
   - `_check_result/{module}.md`
   - `_plans/{module}.md`
   - `_verify_result/{module}.md`

Plain meaning:

1. restore files first
2. then force the module back to candidate closure
3. then require a fresh `cand_check`

### 5.4 Recovery Result

After incomplete promotion recovery completes, the only safe claim is:

1. promotion did not complete
2. repository semantics are restored to the candidate round
3. the module must restart from `cand_check`

The command must not claim:

1. that the new `stable` is formally active
2. that prior verify evidence still safely covers the recovered repository state
3. that promotion can resume from a step later than `cand_check`

---

## 6. Standardized Reason Code

This policy adds one standardized candidate-side recovery code:

1. `promotion_recovery`
   - use only when `cand_promote` had already started mutating repository state and the round had to be restored back to candidate semantics

Other fallback or drift cases must keep using the existing standardized codes from the candidate handoff contract.

---

## 7. Relationship To Other Files

This policy works together with:

1. `specflow/framework/docs/agent_guidelines/candidate_handoff_contract.md`
2. `specflow/framework/docs/agent_guidelines/commands/cand_promote.md`
3. `specflow/framework/docs/agent_guidelines/git_policy.md`

Priority rules:

1. the active command file decides whether a fallback or recovery path is triggered
2. the candidate handoff contract decides the smallest legal fallback step for normal handoff invalidation
3. this file defines the shared repository-restoration rules once recovery is required

---

## 8. Non-Goals

This file does not:

1. create new module lifecycle stages
2. define a general rollback system for arbitrary code changes outside the active command scope
3. replace git-history policy
