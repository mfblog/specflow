# Project Standard Create Flow

## 1. Purpose

This internal flow creates a project-local standard file and registers it in `docs/project_standards/_registry.md` when the user expresses intent that the project needs a new project-local standard.

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
2. "Add a project-local output/reporting rule."
3. "We need a project-local decision or escalation rule."
4. "Generate a standard file for this project rule."

Additional rules:

1. the user does not need to know the internal flow name
2. the agent may infer this flow from intent
3. if the user intent is still too vague to decide the standard's object, the flow must stop and ask for clarification instead of creating an empty rule shell
4. choose `review_standard` only when the user wants review, closure, or checking rules
5. choose `output_standard` only when the user wants output, reporting, or result-format constraints
6. choose `decision_standard` only when the user wants decision, escalation, or approval constraints

---

## 3. Preconditions

Before execution:

1. read `specflow/framework/docs/agent_guidelines/project_standards_policy.md`
2. check whether `docs/project_standards/_registry.md` exists
3. if the registry file is missing, first identify that state as governance drift instead of silently treating it as "no active project-local standards"
4. if the current task itself is creating or repairing the project-local standards extension surface, the flow may continue only as an explicit repair round that also creates or repairs `docs/project_standards/_registry.md`
5. otherwise, stop and report the governance drift before creating any new project-local standard
6. read the current `docs/project_standards/_registry.md` when it exists or once the current round has created or repaired it
7. identify the target standard type, consumed command, and application scope
8. identify the target `surface`
9. confirm that the requested standard does not conflict with the framework baseline
10. if the task will create or modify governance files, read the git policy first

---

## 4. Procedure

1. identify the project-local governance problem the user wants to formalize
2. if `docs/project_standards/_registry.md` was missing at the start of the round:
   - report that missing file as governance drift first
   - create or repair `docs/project_standards/_registry.md` in the same round before registering any new standard
   - do not continue as if the repository had simply chosen to use no project-local standards
3. choose the smallest supported standard type from:
   - `review_standard` for project-local review, closure, or checking rules
   - `output_standard` for project-local output or reporting constraints
   - `decision_standard` for project-local decision or escalation constraints
4. choose the target `surface`
5. choose the target command or internal flow that must consume it
6. choose a stable `standard_id`
7. create one project-local standard file under `docs/project_standards/`
8. write the standard as direct rules, not as patch notes
9. update `docs/project_standards/_registry.md`
10. if the created standard is intended to tighten an existing command gate, ensure the relevant command documentation already allows consumption of project-local standards; if not, update that governance rule in the same task
11. perform git close-out if required

---

## 5. Output Contract

The output must include:

1. the created or updated standard file path
2. the chosen `standard_id`
3. the chosen `type`
4. the chosen `surface`
5. the chosen `consumed_by`
6. the chosen `applies_to`
7. whether `docs/project_standards/_registry.md` had to be created or repaired in this round
8. the registry update result
9. the git close-out result

---

## 6. Non-Goals

This flow does not:

1. create a new framework baseline rule automatically
2. allow project-local standards to weaken framework gates
3. create unregistered files that affect command behavior
4. replace module candidate closure
