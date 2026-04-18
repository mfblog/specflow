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

It does not:

1. invent a shared/module boundary when the first candidate still depends on shared truth that is not yet formalized
2. write `shared_contract_refs=none` as a placeholder when the new module already depends on shared truth

## 3. Preconditions

1. complete the required pre-checks
2. the target module name is explicit
3. the module is not yet in `_status.md`
4. the goal is future design first, not capturing current truth first
5. if the first candidate depends on shared truth that is not yet formalized as `shared_contract`, or if the shared/module boundary is still unstable, do not start `spec_new`; resolve that shared truth through `specflow/framework/docs/agent_guidelines/shared_ops.md` first
6. if the first candidate reuses already-existing shared truth, read the relevant `shared_contract` files before writing `shared_contract_refs`
7. if the round will create, update, or delete any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, read `shared_sync.md`
8. if `_status.md` or other commit-triggering governance files will change, read the git policy first

## 4. Procedure

1. if `s_system_constraints.md` exists, read it as the current formal global baseline; otherwise continue with the "no formal global baseline yet" state
2. decide whether the first candidate already reuses existing shared truth:
   - if no, the round may initialize `shared_contract_refs=none`
   - if yes, the round must bind that shared truth explicitly in the first candidate instead of using `none`
3. define the new module's goals, boundaries, protocols, and main flow
4. create `docs/specs/modules/candidate/c_{module}.md`
5. initialize `frontmatter.version` to `0.1.0`
6. ensure the file covers the core sections of a formal Spec
7. initialize `Global Constraint Alignment`:
   - `system_constraints_stable_ref=s_system_constraints@<current_version>` if the formal global baseline exists, otherwise `none`
   - write `shared_contract_refs=none` only when the first candidate does not yet reuse shared truth
   - if the first candidate already reuses existing shared truth, write the explicit `shared_contract_refs` set and explain that reuse in the candidate body in the same round
   - `shared_mechanism_reuse_summary`
   - `global_constraint_exceptions`
   - `system_constraints_change_proposal`
8. if the round changed Shared Contract bindings or shared files, update the corresponding `bound_modules`
9. update `_status.md`:
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=cand_check`
   - the deterministic row writeback may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> status set-module --module {module} --stable no --candidate yes --active-layer candidate --next-command cand_check --notes <status-note> --create`
10. if the round changed any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, run `shared_sync` after `_status.md` has been updated, even when no additional affected module is known yet
   - the deterministic reconciliation part may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> shared sync-impact --modules {module}` and additional `--shared-refs` / `--shared-ids` filters when the active flow has already identified them
11. perform git close-out if required

## 5. Stop Conditions

1. the first `candidate` exists
2. `_status.md` registration is complete
3. any first-round shared binding required by the candidate has been written explicitly instead of being left as placeholder `none`
4. Shared Contract side effects, if any, are closed
5. the command does not automatically continue into implementation

## 6. Output Contract

1. initiation judgment
2. created file path
3. initialized candidate version
4. initialized formal global baseline reference or `none`
5. initialized explicit Shared Contract binding set or confirmed `shared_contract_refs=none`
6. `_status.md` update result
7. Shared Contract reconciliation result when the round changed shared truth or bindings
8. git close-out result
9. remaining closure items

## 7. Non-Goals

1. creating the first formal `stable`
2. capturing historical behavior
3. automatically entering `cand_impl`
4. creating an independent `system_constraints` candidate file
5. using `shared_contract_refs=none` to postpone required shared-truth closure

## 8. Example

```md
spec_new:module_executor
```
