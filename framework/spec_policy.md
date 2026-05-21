# Spec Policy

This file defines the durable object model for specFlow.

specFlow has three formal object families:

1. `unit`
2. `rule`
3. `repository_mapping`

No other formal lifecycle object is recognized. In particular, `scenario` is not a specFlow object, has no command lifecycle, and must not be created, migrated, advanced, or validated as a supported object type.

## 1. Unit

A `unit` is one independently governed engineering responsibility.

A `unit` may describe a local capability, a service slice, a workflow result, or a complete user-visible result chain. The framework does not require a separate chain object above `unit`. If a team needs end-to-end proof, it should model that proof as a `unit` whose responsibility is the end-to-end result.

Unit main Spec files use these paths:

1. stable unit: `docs/specs/units/stable/s_unit_{unit}.md`
2. candidate unit: `docs/specs/units/candidate/c_unit_{unit}.md`

Unit appendix files use these paths:

1. stable appendix: `docs/specs/units/stable/appendix/s_unit_{unit}_{name}.md`
2. candidate appendix: `docs/specs/units/candidate/appendix/c_unit_{unit}_{name}.md`

Each current-layer unit main Spec must define its formal rule bindings in frontmatter:

```yaml
rule_refs: none
```

or:

```yaml
rule_refs:
  - s_b_rule_example@1.0.0
```

Each current-layer unit main Spec must also define formal unit dependencies in frontmatter:

```yaml
unit_refs: none
```

or:

```yaml
unit_refs:
  - s_unit_agent@0.6.0
```

`unit_refs` has these rules:

1. it may reference only stable unit refs
2. it records dependency only
3. it does not grant permission to modify the referenced unit
4. when a unit body depends on another unit's formal behavior, that dependency must be listed in `unit_refs`
5. when a referenced stable unit is promoted to a new stable version, every current-layer unit still referencing the old stable version must be rerouted to the legal revalidation entry before closure is claimed

## 2. Rule

A `rule` is a shared constraint.

Rules are not lifecycle command targets. Rule governance files may create, bind, extract, sync, or retarget rules, but rule consumers are derived only from current-layer unit `rule_refs`.

Rule files use these paths:

1. stable rule: `docs/specs/rules/stable/s_{g_or_b}_rule_{id}.md`
2. candidate rule: `docs/specs/rules/candidate/c_{g_or_b}_rule_{id}.md`

Rule scope is resolved from rule truth:

1. `rule_scope: global` or an id beginning with `g_rule_` means repository-wide rule
2. `rule_scope: bound` or an id beginning with `b_rule_` means bound shared rule

Rule files must not store consumer lists. `bound_objects` is not the source of rule consumers. The consumer graph is always reconstructed from current-layer unit frontmatter `rule_refs`.

`promotion_owner_unit` remains valid rule frontmatter. It identifies the unit responsible for promoting or owning a rule promotion decision. It does not create a consumer binding.

## 3. Repository Mapping

`repository_mapping` is the durable registry for path and object ownership.

The Object Registry table has exactly this header:

```md
| kind | id | registration_state | implementation_paths | spec_files | responsibility |
```

`kind` may be only:

1. `unit`
2. `rule`

`registration_state` may be:

1. `planned`
2. `landed`

`scope` is not a registry column. Rule global or bound status is resolved from rule truth, not from the registry.

## 4. Status

`docs/specs/_status.md` tracks only formal unit lifecycle rows.

Valid object type:

1. `unit`

The status file must not contain supported `scenario` lifecycle rows. Tooling must reject `object-type=scenario`.

## 5. Dependency Order

The normal dependency direction is:

```text
repository_mapping -> unit -> rule
stable global rule -> unit and rule
rule -> unit
unit -> unit through stable-only unit_refs
```

This means:

1. repository mapping decides path ownership
2. unit truth owns behavior responsibility
3. rule truth owns reusable constraints
4. unit-to-unit dependency is explicit and stable-only
5. no hidden lifecycle object may be inferred from a multi-step workflow

## 6. Candidate Relation Graph

Candidate advancement order is a computed relation, not a manually maintained field.

The computed candidate relation graph may read only these inputs:

1. `docs/specs/_status.md`
2. current-layer candidate unit main Specs
3. same-layer unit appendix files explicitly referenced by a current-layer candidate unit main Spec
4. `unit_refs`
5. `rule_refs`
6. Markdown `.md` links
7. explicit version refs in the forms `c_unit_{unit}@{version}`, `c_b_rule_{rule}@{version}`, and `c_g_rule_{rule}@{version}`

The graph builder must not infer candidate order from prose alone.
If a document says in natural language that one candidate waits for another candidate, the document must also contain one of the explicit reference forms above before tooling may compute an order.

The graph has these edge meanings:

1. stable dependency edge
   - a current-layer unit depends on stable unit or stable Rule truth through `unit_refs` or `rule_refs`
   - stable dependency edges are formal pass foundations only when the referenced stable truth exists and is current enough for the command being run
2. candidate progression edge
   - a current-layer candidate unit main Spec or non-evidence same-layer appendix explicitly references another current candidate unit or candidate Rule
   - the referencing candidate waits for the referenced candidate unit or candidate Rule before it may pass `unit_check` or be automatically advanced
3. reference-only edge
   - an evidence appendix explicitly references a candidate unit or candidate Rule
   - the edge is displayed for traceability but does not block candidate advancement

Cycle semantics are fixed:

1. a cycle made only from stable dependency edges is diagnostic only and does not by itself block a candidate preflight
2. a cycle that contains candidate progression edges blocks every candidate unit in that cycle
3. a blocked candidate must not receive a `unit_check` pass gate and must not be entered by `unit_advance:{unit}`

## 7. Process Evidence

Unit process files may record:

1. `unit_appendix_snapshot`
2. `unit_snapshot`
3. `rule_snapshot`

`unit_snapshot` records the resolved stable unit dependencies listed in `unit_refs`. If it is present, tooling must validate it against current truth. New snapshots should include it when `unit_refs` is non-empty.

Process evidence does not replace truth. It proves what was checked in one command round.
