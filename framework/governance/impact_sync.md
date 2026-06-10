# Impact Sync

Impact sync reconciles downstream units after unit, rule, global rule, or repository mapping truth changes.

It owns consumer discovery, freshness classification, and fallback routing for affected units.

## Triggers

Run impact sync when:

1. a stable unit version changes and another current-layer unit references the prior version.
2. a rule is created, changed, promoted, retired, renamed, merged, split, or rebound.
3. a stable global rule changes or gains an explicit exception.
4. repository mapping changes path ownership, object registration, implementation path registration, or support-surface boundaries used by current truth.
5. a governance flow cannot prove that downstream unit evidence remains current.

## Inputs

Use the smallest durable truth that can prove affected consumers:

1. changed rule or global rule truth.
2. changed repository mapping entries.
3. promoted stable unit reference and release version.
4. current-layer unit frontmatter and dependency fields.
5. current process evidence only after the affected unit set is known.

Do not infer consumers from implementation directories alone.

## Consumer Discovery

Rule consumers are derived from current-layer unit frontmatter:

1. `g_rule_` files apply to every current-layer unit unless the global rule itself defines an explicit exception.
2. `b_rule_` files apply only to units whose `rule_refs` include that rule.
3. rule files must not store consumer lists.

Stable unit dependency consumers are derived from current-layer dependency fields and release-version references.

Repository mapping consumers are derived from object, implementation path, and support-surface registrations that overlap the changed mapping entry.

## Fallback Reason Classification

Use the canonical fallback reason codes from `framework/lifecycle/recovery.md`.
`current`, `plan_drift`, and `implementation_deviation` are additional codes used for
agent-internal routing decisions and are not defined in `recovery.md`:

1. `current` - no process or implementation evidence is invalidated.
2. `truth_drift` - candidate behavior, boundary, or acceptance truth must be rewritten or rechecked.
3. `binding_drift` - a current unit or rule binding no longer matches current truth.
4. `baseline_drift` - a captured dependency or baseline no longer matches current truth.
5. `rule_drift` - a rule snapshot no longer matches the current rule.
6. `truth_incomplete` - required candidate truth is missing or incomplete.
7. `gate_missing` - required check gate evidence is missing or invalid.
8. `plan_drift` - candidate truth remains current, but the plan no longer validates.
9. `implementation_deviation` - implementation no longer satisfies current truth.
10. `evidence_incomplete` - candidate verification evidence is missing or invalid.
11. `stable_verify_invalid` - stable verification evidence is missing or invalid.

When classification is uncertain, use the earliest proven invalidated layer and its canonical reason code.

## Fallback Routing

Use `framework/lifecycle/recovery.md` for the actual process-file deletion and next-command update.

1. `truth_drift`, `binding_drift`, `baseline_drift`, `rule_drift`, and `truth_incomplete` return affected candidate units to `unit_check`.
2. `gate_missing` returns affected candidate units to `unit_check`.
3. `plan_drift` and `implementation_deviation` are handled agent-internally; no SpecFlow command reroute is needed.
4. `evidence_incomplete` returns affected candidate units to `unit_verify`.
5. `stable_verify_invalid` routes affected stable units to `unit_stable_verify`.
6. Stable units invalidated by `binding_drift` or `rule_drift` route to `unit_stable_verify` without rewriting stable truth.
7. Stable truth changes that require a new unit version route through `unit_fork:{unit}`.

## Stable Unit Release Handoff

When an already-existing stable unit version is published:

1. Current candidate consumers may be mechanically retargeted from the exact old `unit_refs` entry to the exact new `unit_refs` entry and must fall back to `unit_check`.
2. Current stable consumers must not have stable truth rewritten by release-version tooling. Remove stale `unit_stable_verify` evidence when present and route the stable consumer to `unit_stable_verify`.
3. If the stable consumer needs to accept the new dependency ref as stable truth, that later truth change routes through `unit_fork:{unit}` and the owning unit lifecycle.

## Rule Sync Handoff

Rule-governance flows notify `framework/governance/rules/rule_sync.md` of changed rule refs.
`rule_sync` computes affected consumers from rule refs and current-layer unit frontmatter, then applies fallback routing through this file and `framework/lifecycle/recovery.md`.

## Removed Scenario Lifecycle

Requests that use `scenario_*`, `scenario_advance:{id}`, or `object-type=scenario` are not impact-sync work.
Stop and report that scenario lifecycle support has been removed.
