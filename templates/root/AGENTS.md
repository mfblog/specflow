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
This addendum is the bootstrap guide for an executor that does not already know `specFlow`.
It gives only the concepts and routing rules needed to choose the first authoritative policy file.
Detailed lifecycle, file, and command rules live in the policy files linked below.

### 1. What specFlow Is

`specFlow` is a governance flow for changing project truth and implementation in the right order.

Core rule:

1. formal truth is written in Specs and registered governance files
2. commands and routing flows act on that truth
3. implementation must not invent behavior that the current truth does not already allow

### 2. Core Terms Must Not Be Guessed

These terms are project-specific.
Do not interpret them by ordinary software-engineering habit before reading the relevant policy.

1. `Spec`
   - a source-of-truth file, not a normal explanation document
2. `unit`
   - a governed object, not automatically a directory, package, service, or module name
3. `scenario`
   - an end-to-end trigger-to-outcome chain, not a unit and not direct implementation ownership
4. `stable`
   - accepted truth
5. `candidate`
   - current proposed truth for a change round
6. `_status.md`
   - a state index, not behavior truth
7. `repository_mapping.md`
   - the truth for path ownership, object boundaries, and repository structure
8. `shared_contract`
   - shared truth reused across formal objects
9. `shared_ops`
   - an internal shared-governance router, not a user-facing command
10. `implementation_change_policy.md`
   - the mandatory gate before any direct implementation-side modification

### 3. Entry Shapes

First classify only the request shape.
Do not treat these shapes as the user's full intent.

1. Exact standard command
   - the user gives explicit `unit` or `scenario` command syntax defined by `command_policy.md`
   - read `specflow/framework/docs/agent_guidelines/command_policy.md`
   - then read the matching file under `specflow/framework/docs/agent_guidelines/commands/`
2. Exact governance review entry
   - the user gives `spec_flow_review` or `spec_flow_design_review`
   - read the matching review policy directly
3. Natural-language request
   - every non-exact request that asks for project design, implementation, verification, promotion, explanation, governance, repository mapping, shared truth, system constraints, or project-local standards
   - read `specflow/framework/docs/agent_guidelines/natural_language_routing.md`

If a request is outside these areas, follow the host agent rules.

### 4. Intent Fragments And Mandatory Gates

A natural-language request may contain several intent fragments at the same time.
Do not force it into one exclusive category.

Common fragments include:

1. implementation work
   - creating, modifying, or deleting repo-tracked code, tests, config, migrations, build scripts, or other implementation-side files
2. `unit` truth
3. `scenario` truth
4. repository mapping or path ownership
5. shared truth or Shared Contract binding
6. system constraints or global defaults
7. governance review
8. explanation-only work

Mandatory gate:

1. if any fragment asks for implementation-side modification, read `specflow/framework/docs/agent_guidelines/implementation_change_policy.md` before any file edit
2. if that policy returns `truth_writeback_required` or `boundary_unclear`, do not start from code
3. if the request has both implementation and truth fragments, route truth first unless the policy proves the implementation is already allowed by current truth

### 5. Hard Stops

1. Do not guess behavior by bypassing source-of-truth files under `docs/specs/`.
2. If you are unsure whether a change is a behavior change, treat it as a behavior change.
3. Behavior changes must not start from code.
4. Do not guess `unit` or `scenario` boundaries from directory shape alone; use `docs/specs/repository_mapping.md`.
5. Do not treat chat-only agreement as durable truth.
6. Do not ask the user to choose internal shared flow names; natural-language routing enters shared governance when needed.
7. Plain `spec_flow_review` and `spec_flow_design_review` use their default scopes unless the user explicitly narrows them.
8. Registered entry index files must keep their managed blocks consistent before commit.
9. For commit requirements and exceptions, read `specflow/framework/docs/agent_guidelines/git_policy.md`.
10. When Spec, command, routing, and git rules conflict, stop and return to the relevant policy file instead of guessing.

### 6. Where Details Live

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
