# Unit New Command

## 1. Purpose

This command creates the first `candidate` Spec for a brand-new unit.

Goals:

1. define the first complete candidate design
2. establish the starting point of the candidate chain
3. register the unit in `docs/specs/_status.md`

## 2. Scope

By default this command handles:

1. first-time project initiation for a new unit
2. units that do not yet have any formally effective version
3. creation of the first `candidate`
4. initialization of `source_basis`, `evidence_appendix_ref`, `system_constraints_ref`, and `shared_contract_refs`

It does not:

1. invent a shared/unit boundary when the first candidate still depends on shared truth that is not yet formalized
2. write `shared_contract_refs=none` as a placeholder when the new unit already depends on shared truth

### 2.1 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_new`-local entry, output, and stop rules.

## 3. Preconditions

1. complete the required pre-checks
2. the target unit name is explicit
3. the unit is not yet in `_status.md`
4. the goal is future design first, not capturing current truth first
5. read `specflow/framework/onboarding_decision_policy.md` and decide the first candidate's `source_basis` and `evidence_appendix_ref`
6. if the first candidate uses `source_basis=existing_implementation` or `source_basis=mixed`, prepare the required evidence appendix in the same round
7. if the first candidate depends on shared truth that is not yet formalized as `shared_contract`, or if the shared/unit boundary is still unstable, do not start `unit_new`; resolve that shared truth through natural-language shared governance first
8. if the first candidate reuses already-existing shared truth, read the relevant `shared_contract` files before writing `shared_contract_refs`
9. if the round will create, update, or delete any unit `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, read `shared_sync.md`
10. if the round may update `bound_objects` or remove intentional-unbound retention fields from a touched Shared Contract file, read every current-layer unit main file needed to derive the real repository-wide binding set of each touched Shared Contract from `shared_contract_refs`
11. if `_status.md` or other commit-triggering governance files will change, read the git policy first

## 4. Procedure

1. if `system_constraints.md` exists, read it as the current formal global baseline; otherwise continue with the "no formal global baseline yet" state
2. decide whether the first candidate already reuses existing shared truth:
   - if no, the round may initialize `shared_contract_refs=none`
   - if yes, the round must bind that shared truth explicitly in the first candidate instead of using `none`
3. define the new unit's goals, boundaries, protocols, and main flow
4. create `docs/specs/units/candidate/c_unit_{unit}.md`
5. initialize `frontmatter.version` to `0.1.0`
6. initialize `frontmatter.source_basis` and `frontmatter.evidence_appendix_ref` according to `onboarding_decision_policy.md`
7. if `source_basis=existing_implementation` or `source_basis=mixed`, create the evidence appendix named by `evidence_appendix_ref`; if `source_basis=new_design` or `source_basis=replacement`, write `evidence_appendix_ref=none`
8. ensure the file covers the core sections of a formal Spec
9. initialize `Global Constraint Alignment`:
   - `system_constraints_ref=system_constraints@<current_version>` if the formal global baseline exists, otherwise `none`
   - write `shared_contract_refs=none` only when the first candidate does not yet reuse shared truth
   - if the first candidate already reuses existing shared truth, write the explicit `shared_contract_refs` set using the Shared Contract binding contract from `specflow/framework/spec_policy.md` Section 6.1 and explain that reuse in the candidate body in the same round
   - `shared_mechanism_reuse_summary`
   - `global_constraint_exceptions`
   - `system_constraints_change_proposal`
10. if the round changed Shared Contract bindings or touched Shared Contract files:
   - derive the real repository-wide binding set of each touched Shared Contract from current-layer unit `shared_contract_refs` plus this round's prepared target-unit candidate writeback
   - if current repository truth is insufficient to derive that touched real binding set safely, stop and reroute through natural-language shared governance from current repository truth instead of guessing
   - update `bound_objects` only as declarative metadata so each touched Shared Contract file matches the real binding set implied by that repository-wide binding view plus this round's prepared target-unit writeback
   - the deterministic metadata writeback may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> shared reconcile-bound-objects --units {unit}` and additional `--shared-refs` / `--shared-ids` filters when the active flow has already identified them
   - if a touched Shared Contract file now has one or more formal bound units after this round, remove or stop carrying any `unbound_retention`, `unbound_retention_reason`, and `unbound_retention_owner` fields from that resulting bound file state in the same round
11. update `_status.md`:
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=unit_check`
   - the deterministic row writeback may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> status set-object --type unit --object {unit} --stable no --candidate yes --active-layer candidate --next-command unit_check --notes <status-note> --create`
12. if the round changed any unit `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, run `shared_sync` after `_status.md` has been updated, even when no additional affected unit is known yet
   - if any touched shared file changed only in `bound_objects` during this round, pass execution-local `bound_objects_only_shared_file_refs` with the exact file refs for those files
   - the deterministic reconciliation part may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> shared sync-impact --shared-refs <shared-ref> --units {unit}` or the corresponding `--shared-ids` form, and at least one shared trigger input must already be known before this deterministic execution starts
13. perform git close-out if required

## 5. Stop Conditions

1. the first `candidate` exists
2. `_status.md` registration is complete
3. any first-round shared binding required by the candidate has been written explicitly instead of being left as placeholder `none`
4. Shared Contract side effects, if any, are closed
5. the command does not automatically continue into implementation
6. if repository truth was insufficient to close shared-truth binding metadata safely, the command stopped and rerouted through natural-language shared governance instead of guessing

## 6. Output Contract

1. initiation judgment
2. created file path
3. initialized candidate version
4. initialized `source_basis`
5. initialized `evidence_appendix_ref` and evidence appendix write result when required
6. initialized formal global baseline reference or `none`
7. initialized explicit Shared Contract binding set or confirmed `shared_contract_refs=none`
8. whether the command had to stop and reroute through natural-language shared governance because repository truth was insufficient to close shared-truth binding metadata safely
9. `_status.md` update result
10. Shared Contract reconciliation result when the round changed shared truth or bindings
11. git close-out result
12. remaining closure items
13. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - report `round conclusion`, `current state`, `next step`, `why this next step`, and `next-stage entry gap`
   - `current state` must explicitly confirm `Active Layer=candidate` and `Next Command=unit_check`
   - `next-stage entry gap` must explicitly confirm that entry into the later different command `unit_check` is already satisfied after `unit_new` closes

## 7. Non-Goals

1. creating the first formal `stable`
2. capturing historical behavior
3. automatically entering `unit_impl`
4. creating an independent `system_constraints` candidate file
5. using `shared_contract_refs=none` to postpone required shared-truth closure

## 8. Example

```md
unit_new:executor
```
