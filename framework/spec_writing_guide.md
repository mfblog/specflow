# Spec Writing Guide

This guide defines the writing contract for specFlow delivery documents.

Files under `specflow/` are framework and delivery documents and are written in English.

Files under `docs/` are project communication documents and are written in Chinese unless a specific delivery artifact requires otherwise.

## 1. Formal Spec Paths

Unit main Specs:

| kind | layer | path |
|---|---|---|
| unit | stable | `docs/specs/units/stable/s_unit_{id}.md` |
| unit | candidate | `docs/specs/units/candidate/c_unit_{id}.md` |

Rule Specs:

| kind | layer | path |
|---|---|---|
| rule | stable | `docs/specs/rules/stable/s_{g_or_b}_rule_{id}.md` |
| rule | candidate | `docs/specs/rules/candidate/c_{g_or_b}_rule_{id}.md` |

`docs/specs/scenarios/**` is not a supported formal Spec path.

## 2. Unit Frontmatter

Each unit main Spec must include these fields:

```yaml
id: {unit}
layer: stable|candidate
version: x.y.z
unit_refs: none
rule_refs: none
```

`unit_refs` may also be a YAML list of stable unit refs:

```yaml
unit_refs:
  - s_unit_agent@0.6.0
```

`rule_refs` may also be a YAML list of exact rule refs:

```yaml
rule_refs:
  - s_b_rule_example@1.0.0
```

Candidate unit Specs must also record the candidate source fields required by the active unit command, such as `candidate_intent`, `source_basis`, `repair_basis`, and `evidence_appendix_ref` when that command requires them.

## 3. Unit Dependencies

`unit_refs` means the current unit depends on the referenced stable unit's formal behavior.

It does not mean:

1. the current unit may edit the referenced unit
2. the referenced unit is part of the current unit's ownership
3. the dependency can point to a candidate unit

If the body says the unit relies on another unit's official behavior, `unit_refs` must list that stable unit ref.

## 4. Rule References

`rule_refs` is the only formal consumer binding for rules.

Rules:

1. `rule_refs` must be in frontmatter
2. no formal `rule_refs` list should be duplicated in the body
3. refs must use exact layer and version
4. refs must be sorted lexically when written as a list
5. `rule_refs: none` means the unit binds no rule

Rule files must not record consumer truth through `bound_objects`.

## 5. Rule Frontmatter

Each rule Spec must include:

```yaml
rule_id: {rule}
rule_scope: global|bound
layer: stable|candidate
rule_version: x.y.z
```

`promotion_owner_unit` may be present when one unit owns the promotion decision.

## 6. Acceptance Criteria

Each current-layer unit main Spec must include a `Testability / Acceptance Criteria` section or an explicitly equivalent acceptance section.

The section must include structured acceptance items:

```yaml
acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for demo behavior.
    pass_condition: Demo behavior passes the declared checks.
    not_runnable_yet: no
```

The acceptance item ids are used by process evidence. Changing ids invalidates existing process files.

## 7. Appendix Files

Appendix files are support truth for one unit.

They do not replace the unit main Spec.

Appendix files must be explicitly linked from the current-layer unit main Spec or named by the relevant frontmatter field.

## 8. Process Snapshots

Process files may include:

1. `unit_appendix_snapshot`
2. `unit_snapshot`
3. `rule_snapshot`

`unit_snapshot` records resolved stable unit dependencies from `unit_refs`.

`rule_snapshot` records resolved rule dependencies from `rule_refs`.

Snapshots prove what one command reviewed. They do not create or replace formal truth.
