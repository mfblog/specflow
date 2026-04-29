# Spec Flow Migrate

## 1. Purpose

`spec_flow_migrate` migrates the current project instance after the repository has received newer `specFlow` framework rules.

It answers five questions:

1. whether current project-instance files can be consumed by the current framework rules
2. which project-instance shape problems can be updated mechanically
3. which process-state files must be invalidated after migration
4. which problems require user judgment or another upstream action before migration can continue
5. what the executor must report after migration, partial migration, or a blocked migration

Plain input `spec_flow_migrate` means full project-instance migration unless the user explicitly narrows the target surface.

This flow is not a standard `unit` or `scenario` command.
It does not use `{command}:{object}` syntax.
It does not create a lifecycle object, replace `spec_flow_review`, or add compatibility aliases for old project-instance shapes.

---

## 2. Migration Target

Project-instance migration targets only the files that must match the current framework contracts before normal `specFlow` routing, commands, review, and tooling can consume the project.

The default target surface is:

1. project-instance truth and state under `docs/specs/**`
2. template-governed process contract files under `docs/specs/_check_result/**`, `docs/specs/_plans/**`, and `docs/specs/_verify_result/**`
3. `docs/specs/_status.md`
4. `docs/specs/repository_mapping.md`
5. `docs/specs/system_constraints.md`
6. registered entry index managed blocks in `AGENTS.md`, `GEMINI.md`, and `CLAUDE.md`

Migration may read framework rules, templates, and tooling contracts as inputs.
Migration must not rewrite framework rules, command files, tooling source, or template files.

If framework-managed files are missing, internally inconsistent, or not updated to the intended target framework version, `spec_flow_migrate` must stop with a `prerequisite_action` checkpoint instead of inferring the target framework from repository history, chat context, or a remote source.

---

## 3. Required Read Surface

Before any migration writeback, read:

1. this file
2. `specflow/framework/checkpoint_protocol.md`
3. `specflow/framework/spec_flow_review.md` Section 2.9 for the project-instance compatibility boundary
4. `specflow/framework/spec_policy.md` for truth-file ownership and binding rules
5. `specflow/framework/command_policy.md` for command ownership and lifecycle boundaries
6. `specflow/framework/process_snapshot_contract.md` for process snapshot shape and invalidation rules
7. `specflow/framework/recovery_policy.md` for fallback cleanup and next-command targets
8. `specflow/framework/entry_index_registry.md` before changing registered entry files
9. `specflow/framework/tooling_execution_policy.md` before using any governance tooling
10. `docs/specs/repository_mapping.md`
11. `docs/specs/_status.md`
12. `docs/specs/system_constraints.md` when it exists
13. the template-side process and state contracts under `specflow/templates/docs/specs/**` that correspond to the project files being migrated
14. the template entry files under `specflow/templates/` when registered entry managed blocks are in scope

When the user explicitly narrows the migration target, read only the subset needed to prove that narrowed migration and its downstream invalidation effects.

---

## 4. Compatibility Scan

The first executable step is a compatibility scan.

The scan must classify every discovered issue into exactly one of these classes:

1. `mechanical_update`
   - the current framework rules or templates define one exact target shape
   - the update does not choose business behavior, object ownership, acceptance meaning, shared-truth ownership, or system-constraint meaning
2. `process_invalidation`
   - a process file or status row can no longer prove the gate, plan, verification, or active-layer state it previously claimed
   - the affected object and fallback target are mechanically determined by current `_status.md`, `process_snapshot_contract.md`, and `recovery_policy.md`
3. `blocked_decision`
   - more than one target meaning is possible, or a business, ownership, acceptance, shared, or system decision is needed
4. `blocked_prerequisite`
   - migration cannot continue until a concrete upstream file, framework update, mapping writeback, or command result exists
5. `out_of_scope`
   - the issue is business-truth correctness, implementation correctness, product design quality, or another concern outside project-instance format migration

The scan must not treat old shape as valid only because older framework versions accepted it.
The target is the current framework rule set in the repository.

---

## 5. Allowed Writeback

`spec_flow_migrate` may write only `mechanical_update` and `process_invalidation` results.

Allowed mechanical updates include:

1. adding, renaming, or reordering required fields when the new field value is mechanically derivable from existing project truth or the current template contract
2. converting tables or frontmatter to the current required shape when every row maps one-to-one
3. replacing registered entry managed blocks with the managed block from the current matching template entry file
4. updating template-governed process README files from current templates
5. deleting or invalidating process files that cannot remain consumable under current snapshot rules
6. updating `_status.md` only when the object, active layer, and next legal command are mechanically determined by current command and recovery rules

Forbidden writeback:

1. do not change unit, scenario, shared-contract, repository-mapping, or system-constraint business meaning
2. do not add fallback logic, compatibility aliases, legacy command names, or dual-format reader rules
3. do not preserve a stale `_check_result`, active plan, `_verify_result`, or status claim by editing its snapshot fields to match new files
4. do not invent `version`, `shared_version`, `source_basis`, `evidence_appendix_ref`, `Next Command`, or binding values when the current project truth does not determine them
5. do not infer object ownership from directory shape when `docs/specs/repository_mapping.md` is missing or unclear
6. do not change implementation-side files
7. do not create or modify `specflow/tooling` source or a `specflowctl migrate` command as part of this flow

If one file contains both mechanically migratable shape and unresolved business meaning, update only the independent mechanical part when doing so cannot hide or change the blocked meaning.
Otherwise stop before writing that file.

---

## 6. Process-State Invalidation

Migration must invalidate process state whenever migrated truth or support files make an existing process claim unprovable under the current framework.

Invalidation rules:

1. if a current `unit` candidate's consumable process state becomes invalid, delete its candidate-side process files and set the unit's next legal command to `unit_check`
2. if a current `scenario` candidate's consumable process state becomes invalid, delete its candidate-side process files and set the scenario's next legal command to `scenario_check`
3. if a current `unit` stable alignment claim becomes invalid, do not delete candidate-side files solely for that stable drift; set the unit's next legal command to `unit_stable_verify`
4. if a current `scenario` stable alignment claim becomes invalid, do not delete candidate-side files solely for that stable drift; set the scenario's next legal command to `scenario_stable_verify`
5. if a process file cannot be tied to one current object and one current layer, do not guess; classify it as `blocked_decision` or `blocked_prerequisite`

Process-state invalidation must follow `specflow/framework/recovery_policy.md`.
Snapshot comparison must follow `specflow/framework/process_snapshot_contract.md`.

---

## 7. Checkpoints

`spec_flow_migrate` may stop through these checkpoint types:

1. `clarification`
   - use when the user narrowed migration scope but the target surface is ambiguous
2. `decision`
   - use when two or more migration mappings are possible and the choice changes durable project truth
3. `prerequisite_action`
   - use when a required framework file, project truth file, repository mapping update, or upstream command result must exist before migration can continue

Checkpoint fields must follow `specflow/framework/checkpoint_protocol.md`.

For checkpoints raised by this flow:

1. `command` must be `spec_flow_migrate`
2. `unit` must be `none`
3. `required_writeback_target` must name the concrete project truth, support file, or upstream action target when the answer affects durable state
4. `resume_next_step` must be rerunning `spec_flow_migrate` from current repository truth unless a more specific prerequisite action is named

---

## 8. Post-Migration Check

After every migration writeback round, run the compatibility scan again over the migrated surface.

The post-migration check must prove:

1. every written file now matches the current framework shape that governed that writeback
2. every process-state invalidation required by changed truth or support files has been applied
3. registered entry managed blocks are consistent
4. remaining problems are classified as `blocked_decision`, `blocked_prerequisite`, or `out_of_scope`
5. no standard command lifecycle result is claimed solely because migration completed

Migration completion does not advance `unit` or `scenario` lifecycle gates.
Any object whose check, plan, verification, or stable alignment was invalidated must re-enter its next legal command after migration.

---

## 9. Output Contract

The final or stop report must include:

1. migration scope
2. current framework target files used as migration authority
3. compatibility scan summary
4. files changed
5. process files deleted or invalidated
6. `_status.md` rows changed
7. registered entry managed-block result
8. remaining blocked decisions or prerequisites
9. out-of-scope findings, if any
10. post-migration check result
11. final conclusion:
    - `migrated`
    - `partially_migrated_blocked`
    - `blocked_no_change`

When no file changed, report `files changed: none`.
When no blocker remains, report `remaining blockers: none`.

---

## 10. Non-Goals

This flow does not:

1. review business truth correctness
2. review whether the current governance design is worthwhile
3. replace `spec_flow_review` or `spec_flow_design_review`
4. create a new `unit`, `scenario`, or `shared_contract`
5. execute implementation work
6. add old-format compatibility behavior
7. introduce new tooling commands
8. infer migration rules from Git history, release notes, or chat-only decisions
