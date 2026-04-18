# Shared Extract Flow

## 1. Purpose

`shared_extract` is the internal flow for extracting already-existing module truth into one independent `shared_contract`.

It answers four questions:

1. whether multiple modules really depend on the same formal truth now
2. which part of current module-local truth should move into one shared object
3. how module-side truth must be rewritten so duplicate formal truth no longer remains
4. how the repository must be reconciled after the shared extraction lands

This is not a user-facing command entry.
The user reaches it through `shared_ops:{natural-language request}`.

---

## 2. Scope

By default it handles requests where shared truth already exists inside one or more modules and now needs to be extracted.

It may:

1. create or update a candidate-layer `shared_contract`
2. rewrite module candidate-side references and boundary explanation
3. remove duplicate formal truth from the source module candidate side
4. update the target shared file's declarative `bound_modules` metadata so it matches the real binding set after extraction writeback
5. trigger `shared_sync` after any shared-truth or binding writeback
6. stop at a `shared_ops` checkpoint when any source or consumer module is currently at `stable`

It does not:

1. design new shared truth from scratch when no module-local source truth exists
2. bind a module to an already-stable shared truth as the only task
3. replace module lifecycle commands
4. promote shared truth into `system_constraints`

---

## 3. Preconditions

Before execution:

1. read `specflow/framework/docs/agent_guidelines/spec_policy.md`
2. read `specflow/framework/docs/agent_guidelines/command_policy.md`
3. read `docs/specs/_status.md` for every named existing module
4. resolve each named module's current layer from `_status.md` before reading its main Spec
5. read the source module current-layer main files and any explicitly referenced appendix truth involved in the extraction
6. read any relevant existing `shared_contract` files that may overlap the target truth
7. read `docs/specs/system/stable/s_system_constraints.md` when the request may cross into project-wide default-rule promotion
8. if any involved module is currently at `stable`, also read `specflow/framework/docs/agent_guidelines/commands/spec_fork.md`
9. if the round may create, update, or delete any module `shared_contract_refs` value or any file under `docs/specs/shared_contracts/**`, read `specflow/framework/docs/agent_guidelines/shared_sync.md` first

---

## 4. Procedure

1. confirm the request is really about extracting already-existing module-local formal truth
2. identify the smallest shared object that multiple modules truly depend on
3. if any involved module current layer is `stable`, do not modify that module `stable` directly:
   - raise a blocking `shared_ops` checkpoint with `type=prerequisite_action`
   - require `spec_fork:{module}` for each such module before extraction continues
   - set `required_writeback_target` to the corresponding module candidate main file set because chat-only agreement does not create legal extraction targets
4. create or update the target candidate-layer `shared_contract`
5. rewrite the source module candidate side so the extracted truth is no longer duplicated as module-local formal truth
6. if additional consumer modules already depend on the extracted truth, update their module candidate-side references and explanations as required
7. update the target shared file's `bound_modules` only as declarative metadata so it matches the real binding set implied by module-side `shared_contract_refs`
8. if duplicate formal truth still remains after extraction, stop and report boundary closure failure
9. after any write to `docs/specs/shared_contracts/**` or any module `shared_contract_refs`, execute `shared_sync` before claiming closure

---

## 5. Stop Conditions

Stop when one of the following is true:

1. the shared extraction is complete, duplicate formal truth is removed, and `shared_sync` has finished reconciliation
   - the target shared file `bound_modules` metadata must already match the real module-side binding set
2. the request is not really extraction and must be re-routed to another shared flow
3. one or more involved modules are currently at `stable` and the flow has raised a `shared_ops` checkpoint for `spec_fork` first
4. module-private truth versus shared truth is still not stably separable
5. the request has crossed into `system_constraints_change_proposal` and must stop at a `shared_ops` checkpoint instead of continuing here

---

## 6. Output Contract

The output must include at least:

1. the extracted shared object and why it belongs to `shared_extract`
2. which involved modules were already at `candidate` and which had to stop for `spec_fork`
3. the source module files that originally carried the truth
4. the target shared-contract file written or updated, or the checkpoint result when extraction could not legally start yet
5. the module candidate-side rewrite result and whether duplicate formal truth was fully removed
6. the target shared file `bound_modules` reconciliation result
7. the `shared_sync` result, including affected modules and fallback if any
8. the git close-out result when governance files or commit-triggering files were changed

---

## 7. Non-Goals

`shared_extract` does not:

1. preserve two formal truths for the same object
2. skip module-side boundary rewrite after creating a shared file
3. leave reconciliation for later after changing shared truth or bindings
4. modify module `stable` truth directly
5. absorb shared conclusions into `system_constraints`
