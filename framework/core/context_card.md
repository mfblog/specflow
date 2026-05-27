# Context Card Standard

A Context Card is the progressive-disclosure contract for a routed specFlow step.

The entry file chooses a card. The card then decides what the executor may read, write, defer, and close. Commands should link owner files instead of repeating global policy.

## Path Resolution

Context Cards may use `framework/...` refs for framework-owned files. These refs are framework-root relative:

1. in installed projects, resolve `framework/...` under `specflow/framework/...`.
2. in the specFlow source repository, resolve `framework/...` under local `framework/...`.

Project truth refs such as `docs/specs/...` remain repository-root relative.

Context Cards use `<tooling-root>/...` for governance tooling command refs:

1. in installed projects, resolve `<tooling-root>/...` under `specflow/tooling/...`.
2. in the specFlow source repository, resolve `<tooling-root>/...` under local `tooling/...`.

## Required Sections

Every lifecycle Context Card must contain these sections in this order:

```text
Required Context
Allowed Writes
Forbidden Writes
On-Demand Expansions
Independent Evaluation
Close Requirements
```

## Required Context

List only the durable files needed to execute the current command.

Do not require the executor to read flat policy, recovery, migration, governance, rule topology, or design-review files by default. Put those files under On-Demand Expansions with concrete triggers.

## Allowed Writes

List the exact files or file families the current command may create or update.

If a write is conditional, name the condition. Anything not named here is forbidden for this command.

## Forbidden Writes

Name hard boundaries for the current command.

At minimum, forbid writes to truth, process evidence, status, repository mapping, rules, and implementation unless the command's Allowed Writes section explicitly permits them.

## On-Demand Expansions

List optional owner files and the trigger that makes each file relevant.

The executor enters an expansion only when the trigger appears in the current work. Returning from an expansion does not expand the default context for future commands.

## Independent Evaluation

State whether the command has an advancing gate that requires an independent reviewer receipt.

If required, link `framework/core/independent_evaluation.md` and name the process evidence file that must carry the receipt.

## Close Requirements

List deterministic close checks and command-close requirements.

If process evidence is consumed by `command close`, the card must name the matching `snapshot validate-process` invocation.
