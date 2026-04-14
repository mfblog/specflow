# Project Standard Create Flow

## 1. Purpose

This internal flow creates a project-local standard file and registers it in `docs/project_standards/_registry.md` when the user expresses intent that the project needs a new project-local review standard.

It answers four questions:

1. when the agent may create a project-local standard
2. what minimum information the created standard must contain
3. how the new standard is registered
4. how this flow avoids inventing hidden rules outside the formal registry

This flow is not a standard module command, is not in `{command}:{module}` form, and is not a user-facing command name.
It is an internal agent flow that may be triggered from user intent.

---

## 2. Trigger

This flow may be used when the user clearly expresses intent such as:

1. "Create a project-specific review standard."
2. "Add a local rule for our project."
3. "We need our own review guideline on top of the framework."
4. "Generate a standard file for this project rule."

Additional rules:

1. the user does not need to know the internal flow name
2. the agent may infer this flow from intent
3. if the user intent is still too vague to decide the standard's object, the flow must stop and ask for clarification instead of creating an empty rule shell

---

## 3. Preconditions

Before execution:

1. read `specflow/framework/docs/agent_guidelines/project_standards_policy.md`
2. read `docs/project_standards/_registry.md` if it exists
3. identify the target standard type, consumed command, and application scope
4. identify the target `surface`
5. confirm that the requested standard does not conflict with the framework baseline
6. if the task will create or modify governance files, read the git policy first

---

## 4. Procedure

1. identify the project-local review problem the user wants to formalize
2. choose the smallest supported standard type from:
   - `review_standard`
   - `output_standard`
   - `decision_standard`
3. choose the target `surface`
4. choose the target command or internal flow that must consume it
5. choose a stable `standard_id`
6. create one project-local standard file under `docs/project_standards/`
7. write the standard as direct rules, not as patch notes
8. update `docs/project_standards/_registry.md`
9. if the created standard is intended to tighten an existing command gate, ensure the relevant command documentation already allows consumption of project-local standards; if not, update that governance rule in the same task
10. perform git close-out if required

---

## 5. Output Contract

The output must include:

1. the created or updated standard file path
2. the chosen `standard_id`
3. the chosen `type`
4. the chosen `surface`
5. the chosen `consumed_by`
6. the chosen `applies_to`
7. the registry update result
8. the git close-out result

---

## 6. Non-Goals

This flow does not:

1. create a new framework baseline rule automatically
2. allow project-local standards to weaken framework gates
3. create unregistered files that affect command behavior
4. replace module candidate closure
