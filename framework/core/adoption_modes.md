# Adoption Modes

Adoption modes define how a project can start using specFlow without committing to the full lifecycle on day one.

They are user entry choices. They are not lifecycle states, not process schema, not harness commands, and not a mode-selection flag for `specflowctl init`.

`specflowctl init` installs the shared project skeleton for every mode. It does not force a project to adopt the full lifecycle immediately.

## Modes

| Mode | Entry | Allowed Scope | Stop Boundary |
|---|---|---|---|
| `reader-only` | Start `specflow-reader` or inspect durable truth manually | Read existing status, truth, repository mapping, and process evidence | Stop before lifecycle commands, process evidence writes, status changes, implementation edits, promotion, stable verification, or governance review |
| `implementation-only` | Natural language request routed through `framework/operations/implementation_change.md` | Change code or tests that already fit written formal truth | Stop when behavior, boundary, acceptance, rule, ownership truth, or lifecycle state must change |
| `single-unit-trial` | Name one unit and use only the lifecycle steps needed for that unit | Use `unit_init`, `unit_new`, `unit_fork`, `unit_check`, `unit_plan`, `unit_impl`, and `unit_verify` for one unit while the rest of the repository stays outside specFlow | Stop before promotion, stable verification, rule governance, or governance review unless the user explicitly asks |
| `unit-check-only` | Run `unit_check:{unit}` against candidate truth | Decide whether a candidate Spec is clear enough to plan from | Stop after pass, blocked, or fix-required check evidence; do not require plan, implementation, verification, promotion, stable verification, or governance review |

## Shared Rules

The selected mode must be visible in the user-facing close-out when the mode limits what the agent will do next.

Mode limits do not weaken hard boundaries:

1. do not edit truth, process evidence, status, implementation, repository mapping, or rules unless the active Context Card or operation policy allows it.
2. do not self-approve an advancing gate that requires independent evaluation.
3. do not treat read-only inspection, implementation-only work, or unit-check-only work as stable acceptance.
4. do not introduce a new process file, lifecycle state, lifecycle command, harness command, or process schema field for adoption mode selection.

When a request exceeds the selected mode, stop at the smallest legal next step and explain the mode boundary in plain language.

Promotion, stable verification, rule governance, and governance review remain available as explicit later choices. They are not default requirements for using specFlow in an adoption mode.
