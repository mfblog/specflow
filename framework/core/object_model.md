# Object Model

This file defines the durable specFlow object model used by active entry files.

## Object Families

specFlow has three formal object families:

1. `unit` - one independently governed engineering responsibility.
2. `rule` - a reusable shared constraint.
3. `repository_mapping` - the durable registry for path and object ownership.

`scenario` is not a supported formal object type.

## Unit

A unit owns behavior truth, implementation planning, implementation work, and verification evidence for one engineering responsibility. A unit may describe a local capability, a service slice, or a complete user-visible result chain.

Unit truth lives in:

| Layer | Main Spec | Appendix |
|---|---|---|
| stable | `docs/specs/units/stable/s_unit_{unit}.md` | `docs/specs/units/stable/appendix/s_unit_{unit}_{name}.md` |
| candidate | `docs/specs/units/candidate/c_unit_{unit}.md` | `docs/specs/units/candidate/appendix/c_unit_{unit}_{name}.md` |

Unit frontmatter records identity, layer, version, `unit_refs`, and `rule_refs`. Candidate units also record `candidate_intent`, `source_basis`, and any required evidence or repair fields.

## Rule

Rules carry shared constraints.

- `g_rule_` rules are global and apply to every current-layer unit.
- `b_rule_` rules apply only to units that explicitly list them in `rule_refs`.

Bound rule consumers are derived from current-layer unit `rule_refs`; rule files must not store consumer lists.

## Repository Mapping

`docs/specs/repository_mapping.md` maps formal objects to implementation paths and ownership responsibilities. Path ownership must be read from this file when it matters; agents must not infer ownership from directory shape alone.

## Process Evidence

Process files record what a command checked in one round. They are evidence, not behavior truth.

Downstream lifecycle gates consume only tool-valid process evidence:

1. `_check_result` for `unit_check` pass evidence.
2. `_plans/active` for `unit_plan` handoff.
3. `_verify_result` for `unit_verify` evidence.
4. `_stable_verify_result` for `unit_stable_verify` advancing evidence.
