# Shared New Flow

## 1. Purpose

`shared_new` is the internal flow for creating shared truth from the start, or opening the next candidate-layer round for an already-independent shared object.

It answers four questions:

1. whether the target truth should exist independently as `shared_contract`
2. whether that truth is really cross-module shared truth rather than one module's private appendix
3. which candidate-layer `shared_contract` file should carry that truth now, including when a stable-layer sibling file already exists
4. how the repository must be reconciled after that shared truth is created or updated

This is not a user-facing command entry.
The user reaches it through `shared_ops:{natural-language request}`.

---

## 2. Scope

By default it handles requests where shared truth should be authored as an independent shared object, whether that object is being created for the first time or reopened at the candidate layer for a new shared round.

It may:

1. create a new candidate-layer `shared_contract`
2. create or update a candidate-layer `shared_contract` for a `shared_contract_id` that already has a stable-layer file
3. update an existing candidate-layer `shared_contract` that is still the same shared object
4. record expected future landing points in planning text when consumer modules do not yet have current-layer candidates
5. trigger `shared_sync` after shared truth writeback

It does not:

1. extract already-written module-local truth out of an existing module body
2. bind one module to an already-existing `shared_contract` as the main task
3. replace module command chains
4. promote shared truth into `system_constraints`

---

## 3. Preconditions

Before execution:

1. read `specflow/framework/docs/agent_guidelines/spec_policy.md`
2. read `specflow/framework/docs/agent_guidelines/command_policy.md`
3. read `docs/specs/_status.md` when the request names existing formal modules
4. resolve every named existing module's current layer from `_status.md` before reading its main Spec
5. read any current-layer module main files already involved in the request
6. read any relevant existing `shared_contract` files if the request names or overlaps them
7. read `docs/specs/system/stable/s_system_constraints.md` when the request may cross into project-wide default-rule promotion
8. if the round may create, update, or delete any file under `docs/specs/shared_contracts/**`, read `specflow/framework/docs/agent_guidelines/shared_sync.md` first

If the request names modules that do not yet have current-layer Spec files and the user intent is explicitly "design shared truth first", do not block on that absence.

---

## 4. Procedure

1. confirm the request is really about independent shared authoring, including creating shared truth from the start or opening the next candidate-layer round for an already-independent shared object, rather than `shared_extract` or `shared_bind`
2. inspect existing module truth and existing shared truth to ensure the target truth is not already formalized elsewhere as duplicate formal truth
3. decide the target shared object boundary:
   - one shared object per shared file
   - do not merge unrelated shared topics into one file
4. if the request is to continue evolving an already-independent shared object that currently has only a stable-layer file, create or update the sibling candidate-layer `shared_contract` for the same `shared_contract_id`
5. otherwise create or update the target candidate-layer `shared_contract`
6. if no consumer module formally binds the shared truth yet:
   - keep `bound_modules=none`
   - record expected future consumers only as planning text in the shared file body
7. if the same truth still remains duplicated as formal module truth elsewhere, stop and report that boundary closure is incomplete
8. after any write to `docs/specs/shared_contracts/**`, execute `shared_sync` before claiming closure, even when the affected-module set is currently empty

---

## 5. Stop Conditions

Stop when one of the following is true:

1. the target candidate-layer `shared_contract` has been written and required reconciliation through `shared_sync` is complete
2. the request is not really independent shared authoring or next-round opening and must be re-routed to another shared flow
3. duplicate formal truth remains in module-local files and boundary closure has not been completed
4. the request is really a pure module retarget or shared impact-check request and must be re-routed to another shared flow
5. the request has crossed into `system_constraints_change_proposal` and must stop at a `shared_ops` checkpoint instead of continuing here

---

## 6. Output Contract

The output must include at least:

1. the recognized shared object and why it belongs to `shared_new`
2. the target shared-contract file written or updated
3. whether the round created the first candidate-layer file for an already-existing stable-layer shared object
4. whether any named modules already bind that truth formally
5. whether duplicate module-local formal truth was found
6. the `shared_sync` result, including whether any modules were affected
7. the git close-out result when governance files or commit-triggering files were changed

---

## 7. Non-Goals

`shared_new` does not:

1. guess through unstable boundaries
2. invent formal module bindings before `shared_contract_refs` exists
3. leave reconciliation for later after changing shared truth
4. absorb shared conclusions into `system_constraints`
