# Spec New Command

## 1. Purpose

This command creates the first `candidate` Spec for a brand-new module.

Goals:

1. define the first complete candidate design
2. establish the starting point of the candidate chain
3. register the module in `docs/specs/_status.md`

## 2. Scope

By default this command handles:

1. first-time project initiation for a new module
2. modules that do not yet have any formally effective version
3. creation of the first `candidate`
4. initialization of `system_constraints_stable_ref` and `shared_contract_refs`

## 3. Preconditions

1. complete the required pre-checks
2. the target module name is explicit
3. the module is not yet in `_status.md`
4. the goal is future design first, not capturing current truth first
5. if the round will create, update, or delete any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, read `shared_sync.md`
6. if `_status.md` or other commit-triggering governance files will change, read the git policy first

## 4. Procedure

1. if `s_system_constraints.md` exists, read it as the current formal global baseline; otherwise continue with the "no formal global baseline yet" state
2. define the new module's goals, boundaries, protocols, and main flow
3. create `docs/specs/candidate/c_{module}.md`
4. initialize `frontmatter.version` to `0.1.0`
5. ensure the file covers the core sections of a formal Spec
6. initialize `Global Constraint Alignment`:
   - `system_constraints_stable_ref=s_system_constraints@<current_version>` if the formal global baseline exists, otherwise `none`
   - `shared_contract_refs=none`
   - `shared_mechanism_reuse_summary`
   - `global_constraint_exceptions`
   - `system_constraints_change_proposal`
7. if the round changed Shared Contract bindings or shared files, update the corresponding `bound_modules`
8. update `_status.md`:
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=cand_check`
9. if the round changed any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, run `shared_sync` after `_status.md` has been updated, even when no additional affected module is known yet
10. perform git close-out if required

## 5. Stop Conditions

1. the first `candidate` exists
2. `_status.md` registration is complete
3. Shared Contract side effects, if any, are closed
4. the command does not automatically continue into implementation

## 6. Output Contract

1. initiation judgment
2. created file path
3. initialized candidate version
4. initialized formal global baseline reference or `none`
5. `_status.md` update result
6. Shared Contract reconciliation result when the round changed shared truth or bindings
7. git close-out result
8. remaining closure items

## 7. Non-Goals

1. creating the first formal `stable`
2. capturing historical behavior
3. automatically entering `cand_impl`
4. creating an independent `system_constraints` candidate file

## 8. Example

```md
spec_new:module_executor
```
