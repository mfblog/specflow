# Lifecycle Overview

The unit lifecycle is available when the selected adoption mode or user request needs formal unit governance.
Reader-only, implementation-only, single-unit-trial, and unit-check-only entry modes are defined in `framework/core/adoption_modes.md`; they do not require every project or task to run the full lifecycle by default.

The formal unit lifecycle stays explicit:

```text
unit_new / unit_fork -> unit_check -> unit_plan -> unit_impl -> unit_verify -> unit_promote
```

`unit_check` remains a formal command. It validates candidate truth readiness. `unit_plan` creates the implementation handoff from a valid check result.

Each lifecycle file is a Context Card. A selected lifecycle command reads the overview, then the matching card, then only that card's required context and triggered on-demand expansions.

Lifecycle authority follows `framework/core/lifecycle_authority.md`: advancing state requires current valid evidence, required independent evaluation receipt, deterministic validation, and successful `command close`.

## Command Forms

A lifecycle Context Card may be selected in either of two ways:

1. the request exactly states one of the command forms below
2. `framework/operations/entry_routing.md` resolves a natural-language request to one of the existing command forms below from current durable truth

Only the existing exact command forms may select a lifecycle Context Card:

```text
unit_init:{unit}
unit_new:{unit}
unit_fork:{unit}
unit_check:{unit}
unit_plan:{unit}
unit_impl:{unit}
unit_verify:{unit}
unit_promote:{unit}
unit_stable_verify:{unit}
```

Do not invent scenario commands, command aliases, or object-type shortcuts.
Requests that use `scenario_*`, `scenario_advance:{id}`, or `object-type=scenario` must stop with a removed-lifecycle report.

## Unit Commands

| Command | Purpose |
|---|---|
| `unit_init:{unit}` | Capture an existing accepted capability as first stable truth |
| `unit_new:{unit}` | Create the first candidate for a new unit |
| `unit_fork:{unit}` | Fork a candidate from existing stable truth |
| `unit_check:{unit}` | Decide whether candidate truth is clear enough for planning |
| `unit_plan:{unit}` | Create or update the implementation plan from checked truth |
| `unit_impl:{unit}` | Implement according to the active plan |
| `unit_verify:{unit}` | Verify implementation against candidate truth |
| `unit_promote:{unit}` | Promote verified candidate truth to stable |
| `unit_stable_verify:{unit}` | Check implementation alignment against stable truth |

## Dependencies

Unit dependency truth is versioned at stable boundaries.

1. Candidate units may depend on current stable unit versions or current candidate truth when the active Context Card permits it.
2. Stable unit promotion must not silently change another unit's consumed stable version.
3. When a stable unit version changes, run `framework/governance/impact_sync.md` for every unit that references the prior version.
4. When a dependency change invalidates downstream process evidence, recover through `framework/lifecycle/recovery.md`.

## Rule Consumption

Lifecycle commands must respect current applicable rules:

1. global rules apply to every current-layer unit unless the rule defines an explicit exception.
2. bound rules apply only when the unit frontmatter includes the rule in `rule_refs`.
3. rule consumer lists are derived from unit truth, not stored in rule files.
4. rule changes, bindings, and topology changes route through `framework/governance/rule_system.md` and `framework/governance/rules/*.md`.

## Hard Gates

1. `unit_plan` consumes a valid `_check_result/unit/{unit}.md`.
2. `unit_impl` and `unit_verify` consume valid check and plan evidence.
3. `unit_promote` consumes valid verify evidence.
4. `unit_stable_verify` advancing outcomes consume valid `_stable_verify_result/unit/{unit}.md`.
5. Advancing check, plan, verify, and stable-verify evidence must contain a valid independent reviewer receipt.
6. Command close is the only lifecycle advancement authority for `_status.md`.
7. A prior non-advancing result does not block later progression when current evidence validates again and command close succeeds.

## Shared Execution Gates

Before mutating lifecycle truth or process files:

1. read the active lifecycle Context Card and its Required Context.
2. prove current `_status.md` names the command as legal for the unit.
3. validate any consumed process file with the matching deterministic tooling when tooling is available.
4. capture the recovery baseline required by `framework/lifecycle/recovery.md` before the first mutation.
5. use `docs/specs/repository_mapping.md` when path ownership, object ownership, or implementation surface ownership matters.

After mutation:

1. run the required deterministic validation.
2. generate the required independent evaluation request before any advancing check, plan, verify, or stable-verify outcome.
3. run required independent evaluation from that request.
4. close through command close before claiming `_status.md` advancement.
5. run `framework/governance/impact_sync.md` when unit truth, rule truth, global rules, dependencies, or repository mapping may affect another unit.

## Automatic Progression

This overview is not the execution owner for `unit_advance:{unit}`.
`unit_advance:{unit}` is governed by `framework/advance_policy.md`, including relation candidate-preflight, blocked candidate stops, and candidate cycle stops.
At summary level, automatic progression may enter only commands already recorded as the next legal command in `_status.md`, and must stop when user intent is required for candidate creation, stable verification judgment, or unclear ownership.
