# Impact Sync Policy

Impact sync reroutes units when shared truth changes.

The only supported command-target object is `unit`.

## 1. Inputs

Impact sync may be triggered by changes to:

1. rule truth
2. repository mapping truth
3. stable global rule truth
4. a promoted stable unit referenced by another unit's `unit_refs`

## 2. Rule Consumers

Stable global rules are repository-wide defaults.
When stable global rule truth changes, every current-layer unit listed in `docs/specs/_status.md` is an affected unit.

Bound shared rule consumers are derived only from current-layer unit frontmatter `rule_refs`.

Rule files do not own consumer lists.

## 3. Unit Dependencies

When a stable unit is promoted, impact sync must find every current-layer unit whose `unit_refs` still references the previous stable version.

Those dependent units must be rerouted to the legal revalidation entry before closure is claimed.
The deterministic tooling entry for this promoted-stable-unit path is `specflowctl unit release-version --unit <unit> --from-ref <old-stable-unit-ref> --to-ref <new-stable-unit-ref>`.

## 4. Reroute Rules

Candidate unit drift:

1. truth or binding drift -> `unit_check`
2. plan drift -> `unit_plan`
3. evidence drift -> `unit_verify`

Impact sync must not infer implementation drift from snapshot mismatch.

When an active command or recovery path has proven implementation deviation while current truth, check gate, and plan still stand, it must use `implementation_layer` and route to `unit_impl`.

Stable unit drift:

1. route to `unit_stable_verify`

## 5. Rejection

Impact sync must not process scenario objects, scenario process files, or `scenario_*` next commands.
