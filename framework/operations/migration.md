# spec_flow_migrate

`spec_flow_migrate` is a dedicated operation for updating SpecFlow file shapes to the current framework contracts after those contracts change.

It is not a unit lifecycle command.
It does not decide product behavior, acceptance meaning, rule meaning, implementation logic, or object ownership.

## 1. Entry

Migration authority exists only when the user explicitly invokes the exact `spec_flow_migrate` entry.
The entry may include a narrowing phrase that names the files or project-instance surfaces to inspect. If no narrowing phrase is given, migration scans all governed project-instance surfaces, detects shape differences between current instance files and the current framework contracts (§2), and writes all differences that have a rule-derived target per §3. Surfaces without a rule-derived target are reported unchanged per §9. Migration does not require intermediate user confirmation before writing rule-derived targets — the narrowing phrase is a scope controller, not an execution gate.

Requests that do not explicitly invoke `spec_flow_migrate` must not receive migration write authority by implication.
They must be handled by the route selected for that request, or stopped when the route is unclear.

## 2. Required Reads

Before any write, migration must read the current owner contract for each target surface it may touch.

Layout-sensitive refs use these roots:

1. in `installed_project`, `<template-root>` is `specflow/templates/` and `<tooling-root>` is `specflow/tooling/`
2. in `source_repo`, `<template-root>` is `templates/` and `<tooling-root>` is `tooling/`

The required read surface is the smallest set that proves the target shape:

1. the framework policy file that owns the target file shape
2. the matching template under `<template-root>/**` when the target shape is template-defined
3. `framework/core/repository_mapping.md` when object registration or path ownership shape is touched
4. `framework/core/object_model.md`, `framework/core/status.md`, and `docs/specs/_status.md` when object state rows are touched
5. `framework/process_snapshot_contract.md` when process files or stored process evidence are touched
6. `framework/operations/entry_routing.md` (Entry File Registration section) when a registered entry managed block is touched
7. `framework/tooling_execution_policy.md` and `<tooling-root>/README.md` when existing tooling is used

Repository history, chat agreement, old file examples, and ordinary word meaning are not target-shape sources.

## 3. Target Shape Rule

Every migration edit must have one current rule-derived target.

A target is rule-derived only when the current owner contract states one of the following:

1. a required file path
2. a required table header
3. a required frontmatter field
4. an allowed field value set
5. a required managed-block shape
6. a required process-file field shape
7. a deterministic tooling command or validation contract

If the current owner contract does not define the target shape, migration must stop for that target.
The executor must not invent a target shape from judgment, naming preference, or old repository shape.

## 4. Allowed Writes

Migration may write only file-shape changes whose target is fixed by Section 3.

Allowed writes are limited to:

1. updating framework policy documents or templates so they state the current shape directly
2. updating project-instance tables, frontmatter, status rows, paths, or registered fields to match the current shape
3. updating registered entry managed blocks per `framework/operations/entry_routing.md` (Entry File Registration section)
4. removing obsolete shape fields when the current owner contract says the field is no longer part of the shape
5. rebuilding deterministic derivatives or running validators when the tooling contract already allows that action

An allowed write changes format, location, field shape, or managed framework text.
It must not change the meaning carried by the file.

## 5. Forbidden Writes

Migration must not:

1. change unit behavior truth
2. change acceptance meaning
3. change rule truth or rule binding meaning
4. change implementation logic
5. choose a new object owner or object responsibility
6. fill missing dependency meaning for an existing unit
7. create compatibility aliases, hidden fallback branches, or repair logic
8. preserve old process evidence as a current pass claim after the evidence source changed
9. edit host-owned content outside a registered entry managed block
10. use tooling to make a semantic decision that belongs to governance rules or runtime reasoning

When a required edit would need any forbidden decision, migration must stop and report the correct owner or next action.

## 6. Process State Handling

Process files are evidence, not behavior truth.

When migration changes a file that existing process evidence depends on, migration must not leave that evidence trusted unless the current process contract still validates it.

For affected unit process state, migration must use the current rules in:

1. `framework/process_snapshot_contract.md`
2. `framework/lifecycle/recovery.md`
3. `framework/governance/impact_sync.md`

Migration may change process state only when those rules define the exact writeback or cleanup action.
If the affected process state cannot be invalidated or rerouted by a current rule, migration must stop and report the affected files and the missing rule-defined next action.

## 7. Tooling Boundary

Existing tooling may be used only for mechanical actions already allowed by `framework/tooling_execution_policy.md`.

Migration may use tooling to collect, parse, validate, rebuild, compare, clean up, transition, sync, or render only when the upstream input and writeback target are already fixed.
Tooling must not decide whether a migration target is semantically correct, whether evidence is sufficient, or which owner should receive a truth decision.

If existing tooling cannot perform the needed mechanical action, migration must report a tooling gap instead of changing the tooling contract by implication.

## 8. Blocked Stop

Migration must stop before writing a target when any of the following is true:

1. the user did not explicitly invoke `spec_flow_migrate`
2. the target surface does not have a current owner contract
3. the target shape cannot be derived from a current owner contract
4. the write would require a forbidden decision from Section 5
5. affected process state cannot be invalidated or rerouted by a current rule
6. registered entry managed-block source selection is unclear
7. existing tooling is required but cannot legally perform the needed mechanical action

A blocked stop report must state:

1. the target that could not be migrated
2. the current contracts that were read
3. the missing contract, unclear source, or forbidden decision
4. the files intentionally left unchanged
5. the smallest legal next owner or action

## 9. Migration Report

After a migration run, the user-facing report must state:

1. files changed
2. the current contract used for each changed surface
3. validators or deterministic tooling commands run
4. process state invalidated, rerouted, or left blocked
5. remaining decisions that are outside migration authority

The report must not claim that unit behavior, acceptance meaning, rule meaning, or implementation logic has been validated unless the owning non-migration process actually did that work.
