# Spec Init Command

## 1. Purpose

This command creates the first `stable` Spec for a historical module.

Goals:

1. capture the module's already-effective formal behavior
2. create the module's first formal truth file
3. register the module in `docs/specs/_status.md`

## 2. Scope

By default this command handles:

1. first-time governance onboarding of a historical module
2. modules that already have implementation and stable behavior but are not yet inside the Spec system
3. creation of the first `stable`

It does not handle:

1. creating a new module
2. forking a new candidate from existing `stable`
3. creating `candidate` directly

## 3. Preconditions

Before execution:

1. complete the required pre-checks from `spec_policy.md`; if the module is not registered yet, at minimum confirm there are no conflicting old status or leftover process files
2. the target module name is explicit
3. the module is not yet in `docs/specs/_status.md`
4. the goal is to capture current truth, not define future design
5. if the task also touches global baseline, shared mechanisms, or exceptions, read `docs/specs/system/stable/s_system_constraints.md`
6. if the module involves technical choices, shared infrastructure, cross-module reuse, global exceptions, or system-level constraint relationships, the first `stable` must include `Global Constraint Alignment` or an equivalent section
7. if the task changes `stable`, `_status.md`, or other commit-triggering governance files, read the git policy first
8. if the round creates, updates, or deletes any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, read `specflow/framework/docs/agent_guidelines/shared_sync.md` first

## 4. Procedure

1. summarize the module's already-effective behavior baseline
2. if needed, read `s_system_constraints.md` as an upstream input
3. create `docs/specs/stable/s_{module}.md`
4. ensure the file covers:
   - `Context & Motivation`
   - `Terminology`
   - `Data Structures / Protocols`
   - `State Machine / Business Flow`
   - `Edge Cases & Error Handling`
   - `Testability / Acceptance Criteria`
5. if needed, add `Global Constraint Alignment` with at least:
   - `system_constraints_stable_ref`
   - `shared_contract_refs`
   - `shared_mechanism_reuse_summary`
   - `global_constraint_exceptions`
6. do not introduce `system_constraints_change_proposal` into the first `stable`; that field belongs only to module `candidate`
7. if the round changed Shared Contract bindings or shared files, update the affected Shared Contract `bound_modules`
8. update `docs/specs/_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=spec_fork`
9. if the round changed any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, run `shared_sync` after `_status.md` has been updated, even when no additional affected module is known yet
10. perform git close-out if required by policy

## 5. Stop Conditions

1. the first `stable` exists
2. `_status.md` registration is complete
3. Shared Contract side effects, if any, are closed
4. the command does not automatically open a candidate round

## 6. Output Contract

1. onboarding judgment
2. created file path
3. whether `Global Constraint Alignment` was required and why
4. `_status.md` update result
5. Shared Contract reconciliation result when the round changed shared truth or bindings
6. git close-out result
7. next-step suggestion

## 7. Non-Goals

1. creating the first `candidate`
2. jumping directly into implementation
3. redesigning the module

## 8. Example

```md
spec_init:module_ai
```
