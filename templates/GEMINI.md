## Host Instructions

Content outside the managed block below belongs to the host repository.

Keep repository-specific rules outside the managed block. `specFlow` tooling may update only the managed block.



<!-- SPECFLOW:BEGIN -->
## specFlow Addendum

Use this lightweight entry procedure for requests that belong to `specFlow`.
The entry only routes the request. The routed lifecycle or operation file is the active Context Card and defines the current required context, allowed writes, forbidden writes, on-demand expansions, independent evaluation, and close requirements.

### 1. First Read

1. If the request exactly matches a standard lifecycle command (`unit_init:{unit}`, `unit_new:{unit}`, `unit_fork:{unit}`, `unit_check:{unit}`, `unit_plan:{unit}`, `unit_impl:{unit}`, `unit_verify:{unit}`, `unit_promote:{unit}`, `unit_stable_verify:{unit}`), read `specflow/framework/lifecycle/overview.md`, then the matching Context Card under `specflow/framework/lifecycle/`. Entry commands share `unit_init_new_fork.md`.
2. If the request exactly matches `unit_advance:{unit}`, read `specflow/framework/advance_policy.md`.
3. For every other `specFlow` request, read `specflow/framework/operations/entry_routing.md` first.

After routing, read only the active Context Card's required context. Enter its on-demand expansions only when their trigger appears.
Framework-root relative paths in routed files use `framework/...` as the logical framework root. In installed projects, resolve them under `specflow/framework/...`; project refs such as `docs/specs/...` remain repository-root relative.

### 2. Authority Boundary

1. Do not edit truth, process evidence, lifecycle status, rules, repository mapping, or implementation files until the active Context Card or operation explicitly allows that write.
2. Do not close an advancing gate from self-assessment. When the active Context Card requires independent evaluation, the process evidence must contain a valid independent reviewer receipt.
3. Do not guess project terms, object ownership, or lifecycle state from directory shape or chat. Read the durable source named by the active Context Card.

### 3. Active Surface

The active layered surface is:

1. `specflow/framework/core/`
2. `specflow/framework/lifecycle/`
3. `specflow/framework/operations/`
4. `specflow/framework/governance/`

### 4. Required Report

For any `specFlow` route, report the user-facing answer first and keep traceability details separate.

The user-facing answer must state the current state, next action, reason, expected result, and remaining gap in plain project-structure language when they apply.
It must not require the user to understand internal object-family names, command names, lifecycle state names, policy-file names, or governance-flow names.

The execution note may name the entry shape, active Context Card, files changed, next legal step, and stop reason.
It must not be required for the user to understand the answer.
<!-- SPECFLOW:END -->
