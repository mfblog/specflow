# Output Standard

specFlow answers should put the user-facing result first and keep internal trace details separate.

This is the ordinary output standard for framework routing, commands, governance flows, and guidance surfaces unless a narrower owner defines stricter local wording.

## User-Facing Result

When applicable, report:

1. current state.
2. completed action.
3. next action.
4. reason for the next action.
5. remaining gap.

Use ordinary project language before internal command names.

## Human Stops

When work cannot continue without user input or a user action, stop with a plain-language stop report.

The stop report must state:

1. what is blocking progress.
2. the one answer, decision, verification, or prerequisite action needed from the user.
3. why the work cannot close correctly without it.
4. where execution resumes after the answer or action.
5. what still cannot be claimed complete.

Internal result names such as `checkpoint`, `decision_checkpoint`, or `human_verify` may appear only as trace details. They do not require a fixed field template.

## Execution Note

Execution notes may include active entry file, command, process files, status updates, and stop reason. They must not be required to understand the answer.

## Command Close-Out

When a lifecycle command, rule-governance flow, or migration step mutates durable truth or process state, the close-out must state:

1. the durable files changed.
2. the validation or review evidence used.
3. the resulting next legal command or route.
4. any affected downstream unit or rule sync work.
5. any claim that cannot yet be made.

Do not claim lifecycle advancement, rule closure, migration completion, or stable alignment unless the owning gate has accepted the evidence.
