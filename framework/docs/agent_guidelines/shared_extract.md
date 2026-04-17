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
2. rewrite module-side references and boundary explanation
3. remove duplicate formal truth from the source module side
4. trigger `shared_sync` after any shared-truth or binding writeback

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

---

## 4. Procedure

1. confirm the request is really about extracting already-existing module-local formal truth
2. identify the smallest shared object that multiple modules truly depend on
3. create or update the target candidate-layer `shared_contract`
4. rewrite the source module side so the extracted truth is no longer duplicated as module-local formal truth
5. if additional consumer modules already depend on the extracted truth, update their module-side references and explanations as required
6. if duplicate formal truth still remains after extraction, stop and report boundary closure failure
7. after any write to `docs/specs/shared_contracts/**` or any module `shared_contract_refs`, execute `shared_sync` before claiming closure

---

## 5. Stop Conditions

Stop when one of the following is true:

1. the shared extraction is complete, duplicate formal truth is removed, and `shared_sync` has finished reconciliation
2. the request is not really extraction and must be re-routed to another shared flow
3. module-private truth versus shared truth is still not stably separable
4. the request has crossed into `system_constraints_change_proposal` and must stop at a `shared_ops` checkpoint instead of continuing here

---

## 6. Output Contract

The output must include at least:

1. the extracted shared object and why it belongs to `shared_extract`
2. the source module files that originally carried the truth
3. the target shared-contract file written or updated
4. the module-side rewrite result and whether duplicate formal truth was fully removed
5. the `shared_sync` result, including affected modules and fallback if any
6. the git close-out result when governance files or commit-triggering files were changed

---

## 7. Non-Goals

`shared_extract` does not:

1. preserve two formal truths for the same object
2. skip module-side boundary rewrite after creating a shared file
3. leave reconciliation for later after changing shared truth or bindings
4. absorb shared conclusions into `system_constraints`
