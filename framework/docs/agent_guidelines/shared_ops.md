# Shared Ops Command

## 1. Purpose

`shared_ops:{natural-language request}` is the only user-facing command entry for shared-truth governance.

It exists because shared work is intent-driven rather than file-name-driven.
Users usually know what they want to do, but not which internal shared flow should own that work.

It answers five questions:

1. whether the request really belongs to shared-truth governance
2. which internal shared flow should handle it
3. whether the request can be handled by one standard shared flow
4. whether the request must fall into `shared_escape`
5. whether the command must stop at a checkpoint instead of continuing automatically
6. whether a multi-step shared request still has unfinished formal follow-up

This file defines the routing and stop rules for `shared_ops`.
It does not replace module commands.

---

## 2. Command Shape

The user-facing command shape is:

```text
shared_ops:{natural-language request}
```

Examples:

```text
shared_ops:我一开始就要设计一个给 agent 和 assistant 共用的结构化输出 fallback 共享契约
shared_ops:把 module_ai 和 module_memory 里共用的 app config topology 抽成 shared contract
shared_ops:module_skill 需要复用 shared_app_config_topology
shared_ops:我刚改了 structured_output_fallback，帮我检查影响哪些模块
```

Rules:

1. the suffix after `shared_ops:` is free-form natural language
2. the request should describe the user's intent, not force an internal chain name
3. users should not be asked to choose among `shared_new`, `shared_extract`, `shared_bind`, `shared_sync`, or `shared_escape`
4. old direct user-facing entries such as `shared_contract_extract_review` and `shared_contract_reconcile` are retired and must not be presented as the preferred interface

---

## 3. Scope

By default `shared_ops` handles only cross-module shared-truth governance.

It may route into one of four standard internal flows:

1. `shared_new`
2. `shared_extract`
3. `shared_bind`
4. `shared_sync`

It may also route into:

5. `shared_escape`

It does not:

1. replace module command chains
2. replace `cand_check`, `cand_plan`, `cand_impl`, `cand_verify`, or `cand_promote`
3. create an independent `system_constraints` command chain
4. allow the executor to invent an ad hoc fifth standard shared flow outside the routing rules here

---

## 4. Preconditions

Before routing a `shared_ops` request:

1. read `specflow/framework/docs/agent_guidelines/spec_policy.md`
2. read `specflow/framework/docs/agent_guidelines/command_policy.md`
3. read `docs/specs/_status.md` when the request names existing formal modules
4. resolve each named existing module's current layer from `_status.md` before reading its main Spec
5. read the current relevant module candidate or stable files after current-layer resolution
6. if the request names modules that do not yet have current-layer Spec files, do not block on that absence before routing
7. read the relevant `shared_contract` files if the request names shared truth directly
8. read `docs/specs/system/stable/s_system_constraints.md` when the request may cross the boundary into global-default-rule promotion

The executor must not route by keyword alone when the named files already show a different formal situation.

---

## 5. Routing Rules

### 5.1 Route To `shared_new`

Use `shared_new` only when the request clearly means:

1. the user wants to design shared truth from the start
2. that truth is intended to exist independently rather than first living in one module appendix
3. the request is not mainly about reusing an already-existing `shared_contract`

Typical signals:

1. "一开始就要设计成共享"
2. "先规划一个多个模块共用的正式协议"
3. "这部分本来就不属于单模块"

### 5.2 Route To `shared_extract`

Use `shared_extract` only when the request clearly means:

1. truth already exists inside one or more modules
2. that truth should now be extracted into one independent `shared_contract`
3. the main task is the boundary extraction itself

Typical signals:

1. "抽取"
2. "提取"
3. "从模块里拿出来"
4. "现在多个模块都在写这一份真相"

### 5.3 Route To `shared_bind`

Use `shared_bind` only when the request clearly means:

1. a `shared_contract` already exists
2. a module now needs to consume it
3. the main task is binding and module-side explanation, not re-designing the shared truth itself

Typical signals:

1. "复用"
2. "绑定"
3. "接入已有 shared contract"

### 5.4 Route To `shared_sync`

Use `shared_sync` only when the request clearly means:

1. a `shared_contract` changed
2. the user wants to know which modules are affected
3. the main task is state fallback, snapshot invalidation, or impact closure

Typical signals:

1. "改了 shared 后检查影响"
2. "哪些模块要回退"
3. "同步 shared 改动后的状态"

### 5.5 Route To `shared_escape`

Use `shared_escape` when the request cannot be stably routed into exactly one standard shared flow.

This is mandatory, not optional.

`shared_escape` must be used when at least one of the following holds:

1. one request simultaneously hits more than one standard shared flow and the action order matters to formal truth
2. the request is really re-drawing the boundary between module-private truth and shared truth
3. the request simultaneously involves shared restructuring and `system_constraints_change_proposal`
4. one existing `shared_contract` needs to be split, merged, renamed, or structurally reorganized
5. current repository truth is insufficient to stably judge which part belongs to shared and which part stays module-private

---

## 6. Procedure

1. confirm the request really belongs to cross-module shared-truth governance
2. resolve the relevant repository truth before routing:
   - use `_status.md` to resolve current layer for any named existing formal module
   - read named `shared_contract` files when shared truth is named directly
   - read `s_system_constraints.md` when the request may cross the shared/system boundary
3. test whether the request belongs to exactly one of `shared_new`, `shared_extract`, `shared_bind`, or `shared_sync`
4. if exactly one standard shared flow applies, route to that flow
5. if routing is not stable, enter `shared_escape`
6. if the routed flow changes shared truth or module shared bindings, do not claim closure until required reconciliation through `shared_sync` is complete
7. if the request crosses into `system_constraints_change_proposal`, stop through `shared_escape` and raise a checkpoint instead of inventing a shared-side continuation
8. if `shared_escape` emitted a `remaining_steps_contract`, do not claim `shared_ops` closure until every listed step has finished under that contract

## 7. Internal Flow Contracts

The routing target decides the immediate next behavior.

Routing targets must follow these formal files:

1. `shared_new` -> `specflow/framework/docs/agent_guidelines/shared_new.md`
2. `shared_extract` -> `specflow/framework/docs/agent_guidelines/shared_extract.md`
3. `shared_bind` -> `specflow/framework/docs/agent_guidelines/shared_bind.md`
4. `shared_sync` -> `specflow/framework/docs/agent_guidelines/shared_sync.md`
5. `shared_escape` -> `specflow/framework/docs/agent_guidelines/shared_escape.md`

Fixed closure rules:

1. if `shared_new` or `shared_extract` writes `docs/specs/shared_contracts/**`, it must not claim closure until `shared_sync` has completed
2. if `shared_bind` changes any module `shared_contract_refs`, it must not claim closure until `shared_sync` has completed
3. if a routed request crosses into `system_constraints_change_proposal`, the shared flow must stop through `shared_escape` and raise a `shared_ops` checkpoint rather than inventing a shared-side continuation
4. no internal shared flow may guess the module current layer without resolving it from `_status.md` first when the named module already exists
5. no internal shared flow may modify module `stable` truth directly; if a shared request needs module truth writeback and the target module is currently at `stable`, the flow must stop at a `shared_ops` checkpoint and require `spec_fork:{module}` first
6. if `shared_escape` emits a `remaining_steps_contract`, finishing only the first routed flow does not close `shared_ops`

---

## 8. Stop Conditions

Stop when one of the following is true:

1. the request has been stably routed into one internal shared flow and that flow has completed its own closure requirements
2. the request has been decomposed by `shared_escape`, every step listed in `remaining_steps_contract` has finished, and all closure requirements of the final step are complete
3. the request has been routed into `shared_escape` and a checkpoint has been raised
4. the request is outside shared-truth governance and must return to module-side truth handling before resume

## 9. Escape And Checkpoint Rules

### 9.1 `shared_escape`

`shared_escape` is not a catch-all executor freedom zone.
Its job is to decompose a complex shared request into smaller valid actions or stop safely.

It must:

1. identify the smallest action units in the current request
2. try to decompose the request into a sequence of standard shared flows
3. stop immediately if the order of that sequence would itself change formal truth
4. raise a checkpoint instead of guessing

Allowed checkpoint types:

1. `clarification`
2. `decision`

### 9.2 Mandatory Checkpoint Conditions

A checkpoint is mandatory when any one of the following holds:

1. the same truth has two or more plausible formal landing points
2. the boundary between shared truth and module-private truth is unstable
3. the boundary between shared truth and `system_constraints_change_proposal` is unstable
4. the execution order of multiple shared actions would change the resulting formal truth
5. current repository truth is insufficient to support a stable decomposition

### 9.3 Shared Checkpoint Output

A `shared_ops` checkpoint must follow `specflow/framework/docs/agent_guidelines/checkpoint_protocol.md`.

Fixed rules:

1. set `command=shared_ops`
2. set `module` to the formal module name only when the current stop is truly about exactly one module
3. otherwise set `module=none`
4. `required_writeback_target` may point to one or more shared-contract files, module candidate files, or appendix files when those are the truth targets that must be updated before resume
5. `resume_next_step` must be the smallest legal follow-up, which is normally rerunning `shared_ops` after the required truth writeback
6. when the checkpoint exists because one or more target modules are still at `stable`, `required_writeback_target` must point to the future module candidate main file set rather than the current stable file set

A `shared_ops` checkpoint must also report at least:

1. the complex intent recognized from the request
2. why automatic continuation is unsafe
3. which boundary, decomposition, or decision point requires user input
4. the recommended action sequence if the user confirms one direction

If the stop reason is a cross-boundary move into `system_constraints_change_proposal`, the checkpoint must additionally report:

1. which formal module candidate must receive the writeback
2. that chat-only agreement is not durable truth
3. that `resume_next_step` is rerunning `shared_ops` only after the module candidate truth has been updated

---

## 10. Output Contract

The output must include at least:

1. the recognized intent from the user request
2. the routed target flow and why that flow owns the request
3. the repository truth inputs used to make the routing decision
4. whether the request required direct module truth writeback and whether any target module first had to stop for `spec_fork:{module}`
5. whether reconciliation through `shared_sync` was required and whether it has completed
6. if `shared_escape` emitted a `remaining_steps_contract`, that contract and the current completion position
7. if routing was unstable, the `shared_escape` result or checkpoint
8. the git close-out result when governance files or commit-triggering files were changed

## 11. Boundary Against Other Objects

### 11.1 Boundary Against Module Commands

1. `module` keeps the full lifecycle command chain
2. `shared_ops` handles only cross-module shared-truth governance
3. if the real task is still single-module candidate closure, do not route into `shared_ops`
4. if the needed truth writeback target is a module currently at `stable`, `shared_ops` must stop for module-side `spec_fork` rather than writing stable directly

### 11.2 Boundary Against `system_constraints_change_proposal`

1. `system_constraints_change_proposal` remains inside module candidate truth
2. `shared_ops` must not create an independent system command chain
3. if the user's true intent has already become "promote a global default rule," `shared_ops` must report that the task has crossed out of shared-only governance
4. `shared_ops` may read `system_constraints_change_proposal` as boundary input, but must not replace the module-side proposal flow

---

## 12. Non-Goals

`shared_ops` does not:

1. replace module lifecycle commands
2. guarantee that every shared request can continue without a checkpoint
3. allow the executor to guess through unstable shared/system/module boundaries
4. define a new independent lifecycle object parallel to modules
5. treat a partially executed multi-step sequence as a closed shared request
