# Unit Entry Commands

This file covers three entry commands: `unit_init:{unit}`, `unit_new:{unit}`, and `unit_fork:{unit}`.

| Command | Purpose |
|---------|---------|
| `unit_init` | Record an existing capability directly as first stable truth (no candidate layer needed) |
| `unit_new` | Create the first candidate truth for a brand-new unit |
| `unit_fork` | Branch a candidate round from existing stable truth for changes or repairs |

## Input

- `docs/specs/_status.md` (if the target unit may already be registered)
- `docs/specs/repository_mapping.md` (if path ownership or registration must be confirmed)
- `docs/specs/units/stable/s_unit_{unit}.md` + stable-layer appendices (unit_fork only)

## Requirements Per Command

### unit_init
The existing accepted capability must be explicit enough to write stable truth without choosing new behavior, acceptance, or ownership.

### unit_new
Candidate truth must be explicit enough to write the first candidate Spec and its source fields.

### unit_fork
- Current stable truth is the baseline for the candidate round
- `candidate_intent` must be determined (`change` or `repair`)
- If a valid stable verify result exists:
  - `controlled_repair_required` → write `repair`
  - `controlled_change_required` → write `change`
  - `aligned` → no specific intent required
- Every stable-layer appendix must have a corresponding same-named candidate-layer appendix

## Not Allowed

- Modify implementation files
- Manually modify lifecycle state
- Modify rule truth or global rules
- Modify other units' truth
- Modify stable-layer truth during `unit_new` / `unit_fork`
- Modify candidate-layer truth during `unit_init`
- Introduce behavior, acceptance, ownership, or rules not yet decided in the Required Context

## Note

`unit_check` is the required follow-up quality gate for all three commands. After unit_new and unit_fork complete, Next Command is set to `unit_check`.

## How to End

| Command | Success Result | Next Step |
|---------|---------------|-----------|
| `unit_init` | `stable_created` | `unit_fork` |
| `unit_new` | `candidate_created` | `unit_check` |
| `unit_fork` | `candidate_created` | `unit_check` |

Close through `command close`. Before closing, ensure all writes are complete and no unresolved rule-governance or ownership issues remain.
