# Repository Mapping

`docs/specs/repository_mapping.md` is the durable registry for object ownership and implementation paths.
It does not define lifecycle state. Lifecycle state lives in `docs/specs/_status.md`.

## Registry Contract

The Object Registry table must use exactly this header:

```md
| kind | id | registration_state | implementation_paths | spec_files | responsibility |
```

Allowed `kind` values are:

1. `unit`
2. `rule`

`scenario` is not a valid `kind`.

Allowed `registration_state` values are:

1. `planned`
2. `landed`

There is no `scope` column.

## Unit Rows

A unit row records:

1. the unit id
2. whether its implementation ownership is planned or landed
3. the implementation path set, or `none`
4. the unit Spec file set
5. the unit responsibility in one readable sentence

For `landed` rows, each implementation path must exist.
For `planned` rows, `implementation_paths` must be `none`.

## Rule Rows

A rule row records:

1. the rule id
2. whether the rule is planned or landed
3. any implementation path set owned by the rule mechanism, or `none`
4. the rule Spec file set, or `none`
5. the shared constraint responsibility

Rule bound or global scope is resolved from rule frontmatter `rule_scope` or by id prefix:

1. `g_rule_` means global
2. `b_rule_` means bound

The registry must not use a `scope` column for this.

## Path Ownership

Repository mapping answers this question:

```text
Which formal owner is allowed to speak for this path?
```

It does not answer:

1. which unit is currently active
2. which unit can be promoted
3. which rule consumers exist
4. which dependencies are current

Those answers come from `_status.md`, unit frontmatter `rule_refs`, and unit frontmatter `unit_refs`.

## Validation

Validation must check:

1. the Object Registry header is exact
2. every registry row has the correct number of cells
3. `kind` is only `unit` or `rule`
4. `registration_state` is only `planned` or `landed`
5. planned objects do not claim implementation paths
6. landed implementation paths exist
7. declared Spec files exist unless explicitly `none`
8. every formal unit row in `_status.md` has an Object Registry row

Validation must reject old `scenario` registry rows rather than migrating them.

## Usage

Read repository mapping before changing ownership boundaries, creating new implementation paths, or deciding whether a request is implementation-only. Do not use directory shape as ownership proof when mapping truth is available.

## Write Boundaries

A repository mapping change may update only ownership and path registration truth.

If a task also changes unit behavior, rule truth, global baseline rules, or unit dependencies, that truth change must be routed through the corresponding owner before implementation is claimed complete.
