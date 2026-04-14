# Checkpoint Protocol

## 1. Purpose

This file defines the structured checkpoint protocol used by `specflow` candidate-side commands.

It answers five questions:

1. what a checkpoint is
2. which checkpoint types are allowed
3. which fields every checkpoint must carry
4. how checkpoint results must be resumed
5. what a checkpoint must never replace

This is a direct rule document for executors.

---

## 2. What A Checkpoint Is

A checkpoint is a structured communication stop raised by a command when the command cannot close correctly without a small amount of human input or human judgment.

It is not:

1. a standard command
2. an independent state machine
3. a process file
4. a behavior source of truth

Plain meaning:

1. the active command is still responsible for the workflow state
2. the checkpoint only formalizes the stop reason, the required user response, and the legal resume path

---

## 3. Allowed Checkpoint Types

Only these checkpoint types are allowed:

1. `clarification`
   - use when the missing blocker is user intent, scope boundary, or acceptance meaning that is not yet formally written into current truth
2. `decision`
   - use when more than one materially different direction remains viable and the user must choose one
3. `human_verify`
   - use when automation is insufficient to close confidence and a small amount of human effect judgment is still required

Type rules:

1. `clarification` and `decision` are allowed only where the active command file explicitly permits them
2. `human_verify` is allowed only where the active command file explicitly permits it
3. executors must not invent additional checkpoint types

---

## 4. Fixed Checkpoint Fields

Every checkpoint must include all of the following fields:

1. `type`
2. `blocking`
3. `command`
4. `module`
5. `question_or_action`
6. `why_blocking`
7. `required_writeback_target`
8. `resume_signal`
9. `resume_next_step`

Field meanings:

1. `type`
   - one of the allowed checkpoint types
2. `blocking`
   - always records whether the current command is fully blocked pending user input
3. `command`
   - the active command that raised the checkpoint
4. `module`
   - the formal module name
5. `question_or_action`
   - the exact input or verification the user must provide
6. `why_blocking`
   - the minimal explanation of why the command cannot close correctly yet
7. `required_writeback_target`
   - where the checkpoint conclusion must be written before command resume, or `none` only when no truth writeback is required
8. `resume_signal`
   - what user response counts as the checkpoint being answered
9. `resume_next_step`
   - the smallest legal next step after the checkpoint is satisfied

---

## 5. Resume Rules

Checkpoint resume must follow these rules:

1. the active command does not become `pass` merely because a checkpoint was raised
2. the active command must re-judge its gate conditions after resume instead of assuming the checkpoint answer fixed everything
3. if the checkpoint conclusion changes behavior truth, boundary truth, or acceptance truth, that conclusion must be written back to the current candidate or required appendix before resume
4. executors must not treat chat-only conclusions as durable truth
5. `resume_next_step` must be the smallest legal step, not the most convenient step

Writeback rules:

1. for `clarification`, `required_writeback_target` should normally be the current candidate main file or a required appendix file
2. for `decision`, `required_writeback_target` should normally be the current candidate main file or a required appendix file when the decision affects behavior truth
3. for `human_verify`, `required_writeback_target` may be `none` only when the checkpoint concerns final confidence rather than truth repair

---

## 6. Boundary Rules

Checkpoint usage is intentionally narrow.

Executors must not use checkpoints to:

1. bypass candidate truth writeback
2. hide an implementation bug as if it were only a user decision
3. replace required automated work with manual work
4. create a second workflow outside the command chain
5. keep asking open-ended preference questions that do not materially affect the active command

Additional type boundaries:

1. `clarification` is not a substitute for missing technical investigation
2. `decision` is not a substitute for executor reluctance to recommend the best option
3. `human_verify` is not a substitute for tests, static checks, or implementation verification that the executor could have performed

---

## 7. Command Relationship

The checkpoint protocol works together with:

1. `specflow/framework/docs/agent_guidelines/command_policy.md`
2. `specflow/framework/docs/agent_guidelines/candidate_handoff_contract.md`
3. the active command file

Priority rules:

1. the active command file decides whether that command may raise a checkpoint and which types are legal
2. `command_policy.md` decides command-level progression and fallback rules
3. this file defines the common structure and resume semantics of checkpoints

---

## 8. Non-Goals

This file does not:

1. create new standard commands
2. define behavior truth for any module
3. authorize executors to keep truth in chat instead of files
4. replace `_check_result`, `_plans`, or `_verify_result`
