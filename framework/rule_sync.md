# Rule Sync

`rule_sync` is the internal rule-governance flow that computes downstream unit impact after rule truth or rule binding changes.

It is the rule-specific impact discovery layer. Once the affected unit set is fixed, generic fallback and cleanup are handed to `impact_sync`.

## 1. Scope

`rule_sync` may:

1. resolve the changed or explicitly in-scope rule refs and rule ids
2. rebuild the bound-rule consumer graph from current-layer unit `rule_refs`
3. determine which current-layer units are affected by the rule change
4. interpret rule-specific execution-local exceptions
5. pass the fixed affected unit set and resolved exceptions to `impact_sync`
6. use the deterministic tooling command `specflowctl rule sync-impact` after scope and exception inputs are known

`rule_sync` must not:

1. rewrite rule truth
2. rewrite unit truth
3. update `docs/specs/repository_mapping.md`
4. decide a rule boundary or topology plan
5. replace `rule_escape`
6. use `bound_objects` as consumer truth

## 2. Required Reads

Before impact is computed, read:

1. `specflow/framework/spec_policy.md`
2. `specflow/framework/spec_writing_guide.md`
3. `specflow/framework/command_policy.md`
4. `specflow/framework/impact_sync_policy.md`
5. `docs/specs/repository_mapping.md`
6. `docs/specs/_status.md`
7. every in-scope rule file
8. every current-layer unit main Spec needed to rebuild the bound-rule consumer graph from `rule_refs`

If the caller changed rule truth, unit bindings, or the rule object map, that writeback must already be present before `rule_sync` computes impact.

If the rule object map changed, `docs/specs/repository_mapping.md` must already contain the intended current truth before `rule_sync` starts.

## 3. Consumer Source

Stable global rules are default inputs for every current-layer unit.
When a changed or explicitly in-scope rule ref or rule id resolves to a stable global rule, the affected unit set is every current-layer unit row in `docs/specs/_status.md`.

For bound shared rules, the only formal consumer source is current-layer unit frontmatter `rule_refs`.

Rule files must not provide consumer truth. `bound_objects` is ignored as a consumer source and must not be reconciled.

## 4. Execution-Local Inputs

The caller may provide:

1. `rule_refs`
   - exact changed or in-scope refs such as `s_b_rule_runtime_model@0.4.0` or `s_g_rule_repository_baseline@1.1.0`
2. `rule_ids`
   - changed or in-scope rule ids when exact refs are not enough by themselves
   - a rule id that resolves to a stable global rule selects the all-current-unit impact path
3. `units`
   - an optional narrowing set after at least one rule trigger is known
4. `deleted_rule_refs`
   - exact bound shared rule refs for Rule files deleted by the caller after the caller already proved from current-layer unit `rule_refs` that those refs have no current consumers
5. `current_stable_landing_unit`
   - the unit whose stable truth was written in the same round
6. `stable_landing_rule_refs`
   - the exact stable rule refs written by that same stable landing round
7. `retargeted_units`
   - candidate units retargeted in the same stable landing round from the old candidate rule ref to the listed stable rule refs

`current_stable_landing_unit` is valid only together with `stable_landing_rule_refs`.

`retargeted_units` may be used only when the caller selected exact old and new rule refs through `rule_refs`, and every retargeted unit is currently candidate.

`deleted_rule_refs` is a terminal-deletion no-impact proof input.
For each deleted ref, `rule_sync` must verify that the ref is not present under `docs/specs/rules/**` and is not referenced by any current-layer unit `rule_refs`.
If any deleted ref still exists as a Rule file or still has a current-layer unit consumer, the no-impact path must fail.

`rule_sync` must not invent execution-local inputs that the caller did not prove.

## 5. Procedure

1. Load the in-scope rule files and record their exact refs.
2. Validate that `docs/specs/repository_mapping.md` is current enough for the in-scope rule object map. If it is missing or conflicting, stop and return control to `rule_escape`.
3. Read `_status.md` and every needed current-layer unit main Spec.
4. Rebuild the real bound shared rule consumer graph from unit `rule_refs`.
5. For `deleted_rule_refs`, verify the terminal no-impact condition:
   - the deleted ref is no longer present under `docs/specs/rules/**`
   - no current-layer unit frontmatter `rule_refs` contains the deleted ref
   - when every input is only `deleted_rule_refs`, close with affected candidate units `none`, affected stable units `none`, and no `impact_sync` fallback
6. Derive the affected unit set:
   - include every current-layer unit when a selected rule ref or rule id resolves to a stable global rule
   - include units that currently bind a changed exact rule ref
   - include units that currently bind a changed rule id when the change applies across that id's current relevant refs
   - include units explicitly retargeted by a same-round stable landing
   - do not include a sibling rule layer only because it has the same `rule_id`, except for the stable global rule all-unit path above
7. Apply only the proven execution-local exceptions:
   - stable landing self-exemption for the exact `current_stable_landing_unit` and exact `stable_landing_rule_refs`
   - explicit candidate fallback for validated `retargeted_units`
8. Convert the final result into `impact_sync` input:
   - final invalidating rule refs
   - final affected candidate units
   - final affected stable units
   - final stable-landing exceptions
9. Hand the fixed result to `impact_sync`.
10. When using tooling, run `specflowctl rule sync-impact` with the exact `--rule-refs`, `--rule-ids`, or `--deleted-rule-refs` and any already-proven exception flags.

If repository truth is insufficient, return control to `rule_escape` without performing fallback cleanup. A caller that already mutated truth must follow its own post-mutation recovery rule or caller-owned blocked transition before rerouting. If the caller has no such post-mutation rule, it must stop before mutation instead of leaving mutated truth without an owner.

## 6. Fallback Result

Affected candidate units fall back according to the reason proven by `impact_sync_policy.md`.

For rule truth or binding drift, the candidate unit falls back to `unit_check`.

Affected stable units route to `unit_stable_verify`.

## 7. Rejection

`rule_sync` must reject:

1. scenario consumers
2. scenario paths
3. scenario commands
4. `object-type=scenario`
5. any attempt to use `bound_objects` as consumer truth

## 8. Output Contract

The output must report:

1. the rule refs or rule ids treated as changed or in scope
2. the affected candidate units
3. the affected stable units
4. whether repository mapping truth was sufficient
5. every execution-local exception applied
6. every retargeted unit validated for explicit fallback
7. every deleted rule ref verified as terminal no-impact
8. whether control passed to `impact_sync`
9. whether control closed as no-impact
10. whether control returned to `rule_escape`
