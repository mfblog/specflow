# Shared New Flow

## 1. Purpose

`shared_new` is the internal flow for creating shared truth from the start.

It answers four questions:

1. whether the target truth should exist independently as `shared_contract`
2. whether that truth is really cross-module shared truth rather than one module's private appendix
3. which candidate-layer `shared_contract` file should carry that truth now
4. how the repository must be reconciled after that shared truth is created or updated

This is not a user-facing command entry.
The user reaches it through `shared_ops:{natural-language request}`.

---

## 2. Scope

By default it handles only requests where shared truth should be designed independently from the start.

It may:

1. create a new candidate-layer `shared_contract`
2. update an existing candidate-layer `shared_contract` that is still the same shared object
3. record expected future landing points in planning text when consumer modules do not yet have current-layer candidates
4. trigger `shared_sync` after shared truth writeback

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

If the request names modules that do not yet have current-layer Spec files and the user intent is explicitly "design shared truth first", do not block on that absence.

---

## 4. Procedure

1. confirm the request is really architecture-first shared creation rather than `shared_extract` or `shared_bind`
2. inspect existing module truth and existing shared truth to ensure the target truth is not already formalized elsewhere as duplicate formal truth
3. decide the target shared object boundary:
   - one shared object per shared file
   - do not merge unrelated shared topics into one file
4. create or update the target candidate-layer `shared_contract`
5. if no consumer module formally binds the shared truth yet:
   - keep `bound_modules=none`
   - record expected future consumers only as planning text in the shared file body
6. if the same truth still remains duplicated as formal module truth elsewhere, stop and report that boundary closure is incomplete
7. after any write to `docs/specs/shared_contracts/**`, execute `shared_sync` before claiming closure, even when the affected-module set is currently empty

---

## 5. Stop Conditions

Stop when one of the following is true:

1. the target candidate-layer `shared_contract` has been written and required reconciliation through `shared_sync` is complete
2. the request is not really architecture-first shared creation and must be re-routed to another shared flow
3. duplicate formal truth remains in module-local files and boundary closure has not been completed
4. the request has crossed into `system_constraints_change_proposal` and must stop at a `shared_ops` checkpoint instead of continuing here

---

## 6. Output Contract

The output must include at least:

1. the recognized shared object and why it belongs to `shared_new`
2. the target shared-contract file written or updated
3. whether any named modules already bind that truth formally
4. whether duplicate module-local formal truth was found
5. the `shared_sync` result, including whether any modules were affected
6. the git close-out result when governance files or commit-triggering files were changed

---

## 7. Non-Goals

`shared_new` does not:

1. guess through unstable boundaries
2. invent formal module bindings before `shared_contract_refs` exists
3. leave reconciliation for later after changing shared truth
4. absorb shared conclusions into `system_constraints`
