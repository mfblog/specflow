## Host Instructions

If the current command or governance flow explicitly consumes project-local standards, follow only the registered files selected by `docs/project_standards/_registry.md` for the current `surface`, `consumed_by`, and `applies_to` scope.

Content outside the managed block below belongs to the host repository.

Keep repository-specific rules outside the managed block. `specFlow` tooling may update only the managed block.

## Mermaid Communication Rules

- When explaining a Mermaid diagram, refer to nodes by the exact visible text shown in the diagram. Do not refer to nodes only as hidden IDs such as `A`, `B`, or `C`.
- If short node identifiers are needed for repeated references, include the identifier in the visible node label as well, for example `B["B. Module Main Spec"]`, and use that same visible name in the prose.

<!-- SPECFLOW:BEGIN -->
## specFlow Addendum

`specFlow` treats Specs and registered governance documents as the source of truth.
This addendum is only an entry index.
It tells the executor when to enter `specFlow` and which policy file owns the detailed rules.

### 1. When To Use specFlow

Use `specFlow` rules when the request involves any of the following:

1. natural-language project design, implementation, verification, promotion, or governance work
2. explicit `unit` or `scenario` command syntax
3. governance review entries such as `spec_flow_review` or `spec_flow_design_review`
4. direct modification of repo-tracked code, tests, Specs, governance files, or implementation-side files
5. repository mapping, path ownership, object boundaries, Shared Contract truth, shared-governance routing, system constraints, or project-local standards

If a request is outside these areas, follow the host agent rules.

### 2. First Files To Read

Choose the first policy file by request shape:

1. Natural-language `specFlow` request:
   - read `specflow/framework/docs/agent_guidelines/natural_language_routing.md`
2. Explicit `unit` or `scenario` command:
   - read `specflow/framework/docs/agent_guidelines/command_policy.md`
   - then read the matching file under `specflow/framework/docs/agent_guidelines/commands/`
3. Direct implementation request:
   - read `specflow/framework/docs/agent_guidelines/implementation_change_policy.md` before any code or implementation-side edit
4. Governance review:
   - `spec_flow_review` -> `specflow/framework/docs/agent_guidelines/spec_flow_review.md`
   - `spec_flow_design_review` -> `specflow/framework/docs/agent_guidelines/spec_flow_design_review.md`
5. Repository mapping, shared governance, system constraints, git handling, or project-local standards:
   - start from the relevant policy listed in Section 4

### 3. Hard Rules

1. Do not guess behavior by bypassing source-of-truth files under `docs/specs/`.
2. If you are unsure whether a change is a behavior change, treat it as a behavior change.
3. Behavior changes must not start from code.
4. Direct implementation requests must be classified through `implementation_change_policy.md`; `truth_writeback_required` and `boundary_unclear` must not start from code.
5. Do not guess `unit` or `scenario` boundaries from directory shape alone; use `docs/specs/repository_mapping.md`.
6. `shared_ops` is an internal shared-governance router, not a user-facing command.
7. Plain `spec_flow_review` and `spec_flow_design_review` use their default scopes unless the user explicitly narrows them.
8. Registered entry index files must keep their managed blocks consistent before commit.
9. For commit requirements and exceptions, read `specflow/framework/docs/agent_guidelines/git_policy.md`.
10. When Spec, command, routing, and git rules conflict, stop and return to the relevant policy file instead of guessing.

### 4. Where Details Live

Use these files for the detailed rules:

1. `specflow/framework/docs/agent_guidelines/natural_language_routing.md`
   - natural-language intent closure, routing, missing intent, and multi-step request handling
2. `specflow/framework/docs/agent_guidelines/command_policy.md`
   - standard command forms, command families, lifecycle order, and shared gate rules
3. `specflow/framework/docs/agent_guidelines/spec_policy.md`
   - Spec objects, layers, source-of-truth boundaries, and file ownership rules
4. `specflow/framework/docs/agent_guidelines/implementation_change_policy.md`
   - direct implementation classification and diversion rules
5. `specflow/framework/docs/agent_guidelines/repository_mapping_policy.md`
   - repository structure and path-ownership governance
6. `docs/specs/repository_mapping.md`
   - current project structure truth
7. `docs/specs/system_constraints.md`
   - current global technical baseline and system constraints
8. `specflow/framework/docs/agent_guidelines/shared_ops.md`
   - internal shared-governance routing after natural-language routing reaches shared work
9. `specflow/framework/docs/agent_guidelines/git_policy.md`
   - commit boundaries and git handling
10. `specflow/framework/docs/agent_guidelines/entry_index_registry.md`
   - registered entry files and managed-block sync rules
<!-- SPECFLOW:END -->
