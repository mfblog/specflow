# Rule Topology

`rule_topology` is the internal rule-governance flow for changing the relationship between rule files and their unit consumers.

It is used when the rule structure itself must change, such as splitting, merging, replacing, retiring, or intentionally keeping an unbound rule.

### Entry Condition

This flow is valid only when the relationship between existing rule files and their unit consumers must change structurally (split, merge, replace, retire, or intentionally unbound). It is not for creating a new rule or binding a unit to an existing rule — use `rule_new` or `rule_bind` respectively.

## 1. Scope

`rule_topology` may:

1. split one rule object into multiple rule objects
2. merge multiple rule objects into one rule object
3. rename, replace, or retire rule objects
4. create, update, or delete candidate rule files required by the topology plan
5. delete a stable rule file only when current repository truth proves it is already unbound and terminal cleanup is legal
6. rewrite candidate unit `rule_refs` and body-level rule consumption explanation
7. write intentional unbound-retention fields for a touched rule that must remain independently authored
8. update `docs/specs/repository_mapping.md` when the rule object map changes
9. run `rule_sync` after topology writeback

`rule_topology` must not:

1. replace `rule_new` for first-time independent rule authoring
2. replace `rule_extract` for simple extraction of unit-local truth
3. replace `rule_bind` for one unit binding to an unchanged rule
4. modify stable unit truth directly
5. create or update stable rule semantics directly to carry a new version or new meaning
6. leave a touched unbound rule file without an explicit delete-or-keep result

## 2. Consumer Source

The topology graph is:

```text
bound rule -> unit
```

The bound rule graph is derived only from current-layer unit frontmatter `rule_refs`.

Stable global rules are repository-wide defaults and affect every current-layer unit.

Rule files must not store consumer truth. `bound_objects` must not be read or written as a consumer list.

## 3. Required Reads

Before any write, read:

1. `framework/spec_writing_guide.md`
2. `framework/candidate_intent.md`
3. `framework/lifecycle/overview.md`
4. `framework/lifecycle/recovery.md`
5. `framework/governance/rules/rule_sync.md`
6. `docs/specs/_status.md`
7. `docs/specs/repository_mapping.md`
8. every touched rule file that may be split, merged, renamed, replaced, retired, or intentionally kept
9. every current-layer unit main Spec needed to derive the full current bound shared rule graph for touched rules, or every current-layer `_status.md` unit row when a touched rule is a stable global rule
10. `framework/lifecycle/unit_init_new_fork.md` when any affected stable unit would require binding writeback
11. `docs/specs/rules/stable/s_g_rule_repository_baseline.md` when the topology change may become a repository-wide default rule

**Layout-aware path note:** Paths in this section are `<framework-root>`-relative. In `source_repo` layout, `<framework-root>` is `framework/`. In `installed_project` layout, `<framework-root>` uses a `specflow/` prefix before `framework/`. `docs/specs/` paths are project-instance paths and are present only in `installed_project` layout.

## 4. Procedure

1. Confirm that the request is a topology change or terminal-state decision, not simple rule authoring, extraction, binding, or sync.
2. Resolve the complete affected unit set from current-layer unit `rule_refs` for bound shared rules, or from every current-layer `_status.md` unit row for stable global rules.
3. If current repository truth cannot prove the complete affected unit set, stop before writeback and return to `rule_escape`.
4. If any affected unit is stable and the topology plan requires changing that unit's binding or body truth, stop before writeback and return control to `rule_escape` to raise a `prerequisite_action` checkpoint requiring `unit_fork:{unit}` for each such unit.
5. Decide the topology plan explicitly:
   - which rule identities remain
   - which new rule identities are created
   - which touched rule files are updated
   - which touched rule files are deleted
   - which touched rule files remain intentionally unbound
6. Before the first file mutation, capture the recovery baseline required by `framework/lifecycle/recovery.md`.
7. Create, update, or delete rule files according to the topology plan.
8. When a new candidate rule is created for a brand-new rule object, write `rule_version: 0.1.0`.
9. When a candidate rule has a stable sibling, write or validate exactly one valid `promotion_owner_unit`.
10. Rewrite every affected candidate unit `rule_refs` and body explanation required by the topology plan.
11. For every touched bound shared rule file with no formal current consumers after writeback, either delete it or write intentional unbound-retention fields (including `unbound_retention_owner: rule_topology`) in the same round.
12. For every touched bound shared rule file with formal current consumers after writeback, remove or stop carrying unbound-retention fields.
13. Do not write consumer lists or `bound_objects` into rule files.
14. Update `docs/specs/repository_mapping.md` in the same round when the topology plan changes the rule object map.
15. Run `rule_sync` after any rule-file write, unit `rule_refs` write, or rule object-map write.
    Execution-local inputs for `rule_sync` (general topology-change case):
    - `rule_refs`: all changed rule refs (split, merged, renamed, replaced, or newly created refs)
    - `rule_ids`: all touched rule ids
    - `units`: affected candidate unit set (units whose `rule_refs` or body explanation were rewritten)
    - `deleted_rule_refs`: only when the effect is terminal deletion after Step 11 has already proven that no current-layer unit consumes the deleted exact rule ref
    - when the only remaining effect for a touched bound shared Rule is terminal deletion, run the `rule_sync` terminal no-impact path with that exact `deleted_rule_ref`
    - that no-impact path may close only when affected candidate units are `none`, affected stable units are `none`, and no current-layer unit `rule_refs` still contains the deleted ref
    - if the deleted ref still has a current-layer consumer, the topology round must not claim no-impact closure; it must route through the normal affected-unit reconciliation or recover before rerouting

If repository truth becomes insufficient before any mutation, stop and return to `rule_escape`. If mutation already happened and closure is no longer safe, apply `framework/lifecycle/recovery.md` before returning to `framework/operations/entry_routing.md`.

## 5. Stop Conditions

Stop when one of these is true:

1. the topology plan is fully written, every touched rule file has a terminal state, any repository mapping update is complete, and `rule_sync` has closed reconciliation or terminal no-impact
2. the request belongs to another rule flow
3. a stable unit requires a `rule_escape` prerequisite checkpoint before binding writeback can continue
4. repository truth is insufficient to prove the affected unit set or topology plan
5. a candidate rule with a stable sibling would exist without exactly one valid `promotion_owner_unit`
6. a touched unbound rule file cannot be safely deleted or intentionally retained

## 6. Output Contract

The output must report:

1. the recognized topology intent
2. the touched rule objects
3. the affected unit set derived from current-layer `rule_refs` for bound shared rules or from all current-layer units for stable global rules
4. the topology result for each touched rule file
5. every rule file created, updated, deleted, or intentionally retained
6. the written `rule_version` for each created or updated candidate rule
7. the `promotion_owner_unit` result when required
8. every unit candidate binding rewrite
9. any repository mapping writeback
10. confirmation that touched rule files do not carry `bound_objects`
11. the deleted bound shared Rule no-impact result when a touched Rule file was deleted after having no current consumers
12. the `rule_escape` prerequisite checkpoint result when stable unit fork prerequisites block writeback
13. the `rule_sync` result or the recovery and rerouting result
