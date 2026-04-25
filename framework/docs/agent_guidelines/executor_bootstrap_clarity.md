# Executor Bootstrap Clarity Standard

## 1. Purpose

This file defines the framework-level documentation clarity standard for `specFlow` governance documents.

It answers one question:

1. can an executor that does not already know `specFlow` read the relevant documents, build the correct concepts, choose the correct first action, and know when to stop

This is a governance review standard.
It is not a project-local standard.
It applies to the framework baseline itself.

Plain meaning:

1. a document may be internally correct and still fail this standard if it depends on unstated `specFlow` knowledge
2. a document may link to another rule file, but the link must be explicit enough that the executor knows why to read it
3. a document must not rely on ordinary software-engineering meanings for project-specific terms such as `unit`, `scenario`, `candidate`, `stable`, `shared_contract`, or `checkpoint`

---

## 2. Review Target

This standard reviews documents that define or route governance behavior.

It applies at minimum to:

1. entry index managed blocks
2. routing policy files
3. command policy files
4. command files
5. shared-governance flow files
6. review policy files
7. process-state contract files
8. tooling execution contract files when they are in the current review scope

It does not review business truth by default.
It reviews whether governance documents can bootstrap a correct executor.

---

## 3. Bootstrap Reader Model

The review must use this reader model:

1. the reader is a capable executor
2. the reader has no prior `specFlow` concept memory
3. the reader has the current user request and the relevant repository files
4. the reader will follow explicit links and read required policy files
5. the reader must not infer hidden rules from repository history, previous conversations, or familiar meanings of similar engineering terms

If a rule only works for a reader who already knows the intended `specFlow` meaning, the rule fails this standard.

---

## 4. Required Check Objects

A review under this standard must check these six objects.

### 4.1 Concept Bootstrap

The document must either define project-specific terms before relying on them or link to the file that defines them.

At minimum, when relevant, the document must not leave these terms to ordinary interpretation:

1. `Spec`
2. `unit`
3. `scenario`
4. `stable`
5. `candidate`
6. `_status.md`
7. `repository_mapping`
8. `shared_contract`
9. `shared_ops`
10. `checkpoint`
11. `implementation_change_policy`

### 4.2 Entry And Owner Resolution

The document must let a new executor decide which policy, command, or governance flow owns the current request.

Required clarity:

1. exact entry shapes must be distinguished from intent fragments
2. natural-language requests must not be forced into one exclusive intent category
3. internal flow names must not be presented as user choices unless they are valid user-facing entries
4. object ownership must be derived from source-of-truth files, not directory shape or naming habit

### 4.3 Action Boundary

The document must state what the executor may do and what it must not do.

Required clarity:

1. whether the flow may edit truth
2. whether the flow may edit implementation
3. whether the flow may advance lifecycle state
4. whether the flow may write process files
5. whether the flow may commit or must consult `git_policy.md`
6. which actions belong to another command or governance flow

### 4.4 Stop And Escalation Boundary

The document must state when the executor must stop instead of guessing.

Required clarity:

1. missing user intent
2. missing object boundary
3. ambiguous truth landing point
4. unstable shared or system boundary
5. implementation request that may require truth writeback
6. conflicting policy rules
7. missing prerequisite command or writeback target

### 4.5 Cross-Document Dependency Closure

The document must make required dependencies explicit.

Required clarity:

1. which file must be read first
2. which source-of-truth file supplies current object state
3. which policy controls a delegated decision
4. which downstream flow must be rerun after a stop or checkpoint
5. which file owns the durable writeback target

If a document depends on another file but does not say why or when to read it, the dependency is not closed.

### 4.6 Output And Resume Contract

The document must define what the executor reports at the end of the action or stop.

Required clarity:

1. success or non-pass result meaning
2. current state
3. next legal step
4. why that next step is legal
5. checkpoint fields when checkpointing is allowed
6. resume signal and resume path
7. whether a temporary contract is execution-local or durable truth

---

## 5. Finding Rules

A failure under this standard is a governance finding when it can cause a new executor to:

1. start from code when truth writeback is required
2. choose the wrong command or governance flow
3. treat a non-command object as a command target
4. guess `unit` or `scenario` ownership from directory shape
5. leave a checkpoint or resume path unclear
6. treat chat-only agreement as durable truth
7. skip required downstream reconciliation
8. narrow a default review scope without explicit user instruction

The finding must identify:

1. the document that fails the standard
2. the missing concept, owner, action boundary, stop boundary, dependency, or output contract
3. the executor mistake this could cause
4. the smallest repair that would let a new executor proceed correctly

---

## 6. Passing Rule

A document passes this standard only when a new executor can answer these questions from the document and its explicit links:

1. what object or flow is being governed
2. which terms must not be guessed
3. which file to read first
4. what action is allowed
5. what action is forbidden
6. when to stop
7. what to report
8. how to resume

If one of these answers is required for correct execution and is missing, the document does not pass this standard.

---

## 7. Relationship To Reviews

`spec_flow_review` consumes this standard as a framework-baseline governance standard.

Rules:

1. default `spec_flow_review` must report an executor-bootstrap clarity result
2. a narrowed review that includes routing, command, checkpoint, shared governance, process-state, or entry-index behavior must include this standard for the in-scope files
3. a pass claim for in-scope governance documents must not ignore executor-bootstrap clarity failures

`spec_flow_design_review` may use this standard as evidence for human operability, but this file does not change that review's scoring model by itself.

---

## 8. Non-Goals

This standard does not:

1. require every document to repeat every `specFlow` concept
2. forbid explicit links to authoritative policy files
3. replace `spec_policy.md`, `command_policy.md`, or `natural_language_routing.md`
4. review whether the governance design is worth using
5. create a project-local standard surface under `docs/project_standards/`
