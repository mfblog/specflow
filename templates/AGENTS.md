## Host Instructions

Content outside the managed block below belongs to the host repository.

Keep repository-specific rules outside the managed block. `specFlow` tooling may update only the managed block.



<!-- SPECFLOW:BEGIN -->
## specFlow Addendum

### 1. What specFlow Is

This repository uses specFlow to manage development work.

specFlow maintains project documents that record accepted design, behavior, boundaries, acceptance criteria, shared rules, and code ownership. Design, development, modification, testing, and verification work should proceed around those documents.

This does not mean every code edit must change a spec document. A request must enter the specFlow flow first only when it changes documented project truth, or when the current documents are not clear enough to choose one correct implementation result.

### 2. Spec Document Types

Spec documents have two types:

1. `unit`
   A unit is one governed engineering responsibility. It may describe a feature, module, service, or end-to-end result. A unit is not the same thing as a directory, and directory shape alone must not define it.
2. `rule`
   A rule is shared truth that multiple units may need to follow. A rule carries reusable constraints, prohibitions, default requirements, or shared process requirements.

### 3. Spec Document Layers

Spec documents have two layers:

1. `stable`
   Stable is the accepted current project truth. Implementation should conform to stable documents.
2. `candidate`
   Candidate is proposed next project truth. Candidate truth must be checked, planned, implemented, and verified before it can be promoted to stable truth.

### 4. State Files

specFlow has two important state files:

1. `docs/specs/_status.md`
   Records whether each unit is currently on the stable or candidate layer, and records the only legal next lifecycle command for that unit.
2. `docs/specs/repository_mapping.md`
   Records ownership between units or rules, spec files, and implementation paths. It answers which formal object owns a path. It does not answer the next lifecycle step.

### 5. Command Format

specFlow commands use this format:

```text
command:{unit}
```

Examples:

```text
unit_new:agent_runtime
unit_impl:agent_runtime
```

Exact commands have priority over natural-language routing. When a user request exactly matches a command, read the linked owner file and follow that owner. Do not reinterpret an exact command as an ordinary implementation request.

Before any lifecycle action, implementation proposal, reconciliation plan, test-repair plan, or repo-tracked file edit, select the active owner through the Command Index or First Read rules.

### 6. Command Index

When the user request exactly matches one of these commands, read the linked owner file first and follow that file.

#### Unit Lifecycle Commands

Before reading any lifecycle command owner file except `unit_advance:{unit}`, first read:

```text
specflow/framework/lifecycle/overview.md
```

| Command | Owner file | Purpose |
|---|---|---|
| `unit_init:{unit}` | `specflow/framework/lifecycle/unit_init.md` | Capture an existing accepted capability as first stable truth. |
| `unit_new:{unit}` | `specflow/framework/lifecycle/unit_new.md` | Create the first candidate truth for a new unit. |
| `unit_fork:{unit}` | `specflow/framework/lifecycle/unit_fork.md` | Fork candidate truth from existing stable truth for a change or repair. |
| `unit_check:{unit}` | `specflow/framework/lifecycle/unit_check.md` | Check whether candidate truth is clear enough for planning. |
| `unit_plan:{unit}` | `specflow/framework/lifecycle/unit_plan.md` | Create or update the implementation plan from checked candidate truth. |
| `unit_impl:{unit}` | `specflow/framework/lifecycle/unit_impl.md` | Implement according to the current plan. |
| `unit_verify:{unit}` | `specflow/framework/lifecycle/unit_verify.md` | Verify implementation against candidate truth. |
| `unit_promote:{unit}` | `specflow/framework/lifecycle/unit_promote.md` | Promote verified candidate truth to stable truth. |
| `unit_stable_verify:{unit}` | `specflow/framework/lifecycle/unit_stable_verify.md` | Check whether current implementation still conforms to stable truth. |
| `unit_advance:{unit}` | `specflow/framework/advance_policy.md` | Automatically advance through the next legal command recorded in `_status.md`. |

#### Framework Commands

| Command | Owner file | Purpose |
|---|---|---|
| `spec_flow_review` | `specflow/framework/governance/review.md` | Run the default scoped governance review. |
| `spec_flow_review:full` | `specflow/framework/governance/review.md`, then `specflow/framework/spec_flow_review.md` | Run the full governance review. |
| `spec_flow_design_review` | `specflow/framework/governance/review.md`, then `specflow/framework/spec_flow_design_review.md` | Run the default full-scope design-baseline review. |
| `spec_flow_migrate` | `specflow/framework/operations/migration.md` | Run the specFlow migration flow. |

### 7. Development Loop

The normal unit development loop is:

```text
unit_new / unit_fork -> unit_check -> unit_plan -> unit_impl -> unit_verify -> unit_promote
```

For stable implementation alignment checks, use:

```text
unit_stable_verify
```

Lifecycle state may advance only through legal command closure. Do not manually edit `_status.md` as a substitute for the lifecycle flow.

### 8. Natural-Language Requests

If the user request does not exactly match a specFlow command, first decide whether it changes documented project truth.

Documented project truth includes:

1. behavior
2. interface or protocol
3. field meaning
4. default value
5. validation rule
6. error semantics
7. state transition
8. unit responsibility boundary
9. acceptance criteria
10. rule content
11. rule binding
12. repository mapping path ownership

If the request changes any of these items, or if current documents are not clear enough to choose one correct implementation result, do not edit implementation files directly. Route the request into specFlow first.

If the request is only an implementation-side cleanup, refactor, test addition, logging or observability improvement, performance optimization, or a small repair of an implementation deviation already defined by current spec truth, and it does not change the project truth listed above, it may proceed through direct implementation change.

### 9. First Read

When no exact command matches, read only the first matching owner in this order:

1. If the request asks for formal truth creation or change, no formal truth, behavior, protocol, boundary, acceptance, rule, ownership, lifecycle, lifecycle state, Next Command, stable/candidate state, unit phase, repository mapping, guidance, skipping `_status.md` or owner checks, reconciliation, audit, alignment, or gap-review, read:

   ```text
   specflow/framework/operations/entry_routing.md
   ```

2. If the request may change field meaning, schema fields, output fields, fixture fields, contract-like log fields, or downstream compatibility, read:

   ```text
   specflow/framework/operations/entry_routing.md
   ```

   This does not apply when the user explicitly limits the work to internal non-semantic implementation support.

3. If the request is limited to implementation-side code, tests, configs, prompts, fixtures, integration scripts, or other repo-tracked implementation files, and no exact lifecycle command is already active, read:

   ```text
   specflow/framework/operations/implementation_change.md
   ```

4. For every other request that may affect specFlow, read:

   ```text
   specflow/framework/operations/entry_routing.md
   ```

After the first owner routes the request, continue only through the routed owner. Do not create a replacement flow.

### 10. Pre-Action Rules

Before editing any implementation file, prove that the active owner allows the implementation edit.

Before changing behavior truth, acceptance truth, object ownership, rule truth, global rules, lifecycle state, process evidence, or repository mapping, prove that the active owner allows that write.

Resolve path ownership, object boundaries, unit state, and implementation permission from the durable documents named by the active owner. Do not guess from directory shape, code shape, or chat.

Testing, debugging, review, and exploration may inspect or verify. They do not authorize mutation by themselves. If exploration discovers a required change to behavior, protocol, boundary, acceptance, rule, lifecycle state, ownership, or implementation permission, stop and return to the legal owner.

Do not close an advancing gate from self-assessment. When the active Context Card requires independent evaluation, the process evidence must contain a valid independent reviewer receipt.

### 11. No Custom Flow

Do not create a custom reconciliation, audit, alignment, or gap-review flow to replace the recorded `Next Command`, active Context Card, or operation owner.

If the user describes work with those words, still route the request through the legal owner first.

Do not rename an implementation plan as a review, reconciliation, or audit to avoid truth writeback, command preconditions, independent evaluation, or forbidden writes.

### 12. Hard Stops

Stop instead of guessing when any of these are true:

1. the user intent or target object is unclear
2. path ownership, object boundary, current state, or support-surface ownership is unclear
3. behavior, acceptance, boundary, shared rule, or system truth exists only in chat and has not been written into durable truth
4. implementation permission is not proven
5. rule truth, global rule, lifecycle, repository mapping, or process ownership is unclear
6. a prerequisite command, truth writeback, checkpoint, or verification gate is required first
7. spec, routing, lifecycle, implementation-change, review, migration, rule-governance, or entry-sync rules conflict

### 13. Required Output

For any specFlow route, report the user-facing answer first and keep traceability details separate.

The user-facing answer should use ordinary project language when possible and state:

1. current state
2. next action
3. why that action is legal
4. expected result
5. remaining gap

The user-facing answer must not require the user to understand internal object-family names, command names, lifecycle state names, owner file names, or governance-flow names.

The traceability note may name the entry shape, first owner, routed owner, allowed action, forbidden action, files changed, and stop reason. It must not be required for the user to understand the answer.

### 14. Rule Locations

Detailed routing, lifecycle, implementation-change, migration, governance review, rule-governance, repository mapping, guidance, onboarding, and entry-sync rules live under `specflow/framework/`.

Project truth inputs live under `docs/specs/`.

Framework-root relative paths in routed files use `framework/...` as the logical framework root. In installed projects, resolve them under `specflow/framework/...`; project refs such as `docs/specs/...` remain repository-root relative.
<!-- SPECFLOW:END -->
