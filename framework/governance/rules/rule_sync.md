# Rule Sync

`rule_sync` is the internal rule-governance flow that computes downstream unit impact after rule truth or rule binding changes.

It is the rule-specific impact discovery layer. Once the affected unit set is fixed, generic fallback and cleanup are handed to `impact_sync`.

### Entry Condition

This flow must be run after any rule truth or rule binding mutation. It computes the set of affected units and determines whether downstream process evidence is invalidated. It is not a lifecycle command — it is invoked automatically by rule-governance flows.

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

1. `framework/spec_writing_guide.md`
2. `framework/candidate_intent.md`
3. `framework/lifecycle/overview.md`
4. `framework/governance/impact_sync.md`
5. `framework/process_snapshot_contract.md`
6. `docs/specs/repository_mapping.md`
7. `docs/specs/_status.md`
8. every in-scope rule file
9. every current-layer unit main Spec needed to rebuild the bound-rule consumer graph from `rule_refs`

==ATOM_BEGIN:rule_layout_note==
**Layout-aware path note:** Paths in this file are `<framework-root>`-relative. In `source_repo` layout, `<framework-root>` is `framework/`. In `installed_project` layout, `<framework-root>` uses a `specflow/` prefix before `framework/`. `docs/specs/` paths are project-instance paths and are present only in `installed_project` layout.
==ATOM_END:rule_layout_note==

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
   - the set of units the caller interacted with, including modified units and
     units whose binding was read. `rule_sync` derives the affected unit set from
     rule triggers (Procedure step 6), not from this input as an independent filter.
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
   - include candidate units whose process snapshots (check or verify result) contain a reference to a changed or removed rule ref, detected via stale-evidence reconciliation. For the manual handoff path (tooling unavailable): read each flagged candidate unit's `_check_result` and `_verify_result` files, extract their `rule_snapshot` fields, and compare rule refs against the in-scope changed rule refs and rule ids. Units whose snapshots reference a changed or removed ref are added to the affected-unit set.
   - include units whose main Spec body text was modified by the caller, as reported through the execution-local inputs
   - include units explicitly retargeted by a same-round stable landing
   - do not include a sibling rule layer only because it has the same `rule_id`, except for the stable global rule all-unit path above
7. Apply only the proven execution-local exceptions:
   - stable landing self-exemption for the exact `current_stable_landing_unit` and exact `stable_landing_rule_refs`
   - explicit candidate fallback for validated `retargeted_units`
8. Convert the final result into `impact_sync` input:
   - `invalidating_rule_refs`: the exact rule refs whose change or removal triggers fallback
   - `affected_candidate_units`: the candidate units whose evidence must be invalidated
   - `affected_stable_units`: the stable units whose evidence must be invalidated
   - `stable_landing_exceptions`: any units exempted by stable-landing self-exemption
   These fields map directly to `impact_sync.md` Consumer Discovery entry points. The
   pre-computed set is authoritative; `impact_sync` must not re-derive consumers from
   `rule_refs` when called from `rule_sync`.
9. Hand the fixed result to `impact_sync`:
   - If tooling is available, run `specflowctl rule sync-impact` with the exact `--rule-refs`, `--rule-ids`, or `--deleted-rule-refs` and any already-proven exception flags. The tooling output subsumes the manual handoff and returns the structured result.
   - If tooling is unavailable, handoff is manual via the Rule Sync Handoff contract in `framework/governance/impact_sync.md` Section "Consumer Discovery". The affected-unit set and invalidating rule refs are the authoritative consumer list for this round.

If repository truth is insufficient, return control to `rule_escape` without performing fallback cleanup. A caller that already mutated truth must follow its own post-mutation recovery rule or caller-owned blocked transition before rerouting. If the caller has no such post-mutation rule, it must stop before mutation instead of leaving mutated truth without an owner.

## 6. Stop Conditions

`rule_sync` terminates through one of the following conditions:

| Condition | Description | Next Action |
|-----------|-------------|-------------|
| **Normal completion** | Impact computed from the in-scope rule refs and current-layer unit truth (see Procedure steps 1–9). The affected unit set and resolved exceptions are packaged into `impact_sync` input. | Hand the fixed result to `framework/governance/impact_sync.md` for downstream unit fallback and **wait for completion**. Impact sync applies fallback routing and returns control. The affected-unit set and exception results are authoritative for this round. |
| **No-impact close** | `deleted_rule_refs` is the only input and verification proves the ref is absent from `docs/specs/rules/**` and from all current-layer unit `rule_refs` (see Procedure step 5) | Close with `impact_sync` fallback not required. Report affected candidate units `none`, affected stable units `none`. |
| **Stale-evidence reconciliation complete** | Process snapshots (check or verify result) for flagged candidate units were found to contain a reference to a changed or removed rule ref (see Procedure step 6 fourth bullet). These units are added to the affected-unit set. | Hand the stale-evidence-flagged units to the affected-unit set for `impact_sync` fallback. The owning caller must include the flagged units in any post-`rule_sync` repair or recovery procedure. |
| **Insufficient repository truth** | Repository mapping truth is missing or conflicting, or current-layer unit truth cannot be read (see Procedure step 2 and post-step-9 paragraph) | Return control to `rule_escape`. The caller must follow its own post-mutation recovery rule or caller-owned blocked transition before rerouting. |

## 7. Fallback Result

Affected candidate units fall back according to the reason proven by `framework/governance/impact_sync.md`.

For rule truth or binding drift, the candidate unit falls back to `unit_check`.

Affected stable units route to `unit_stable_verify`.

## 8. Release-Version Auto-Fork

When a rule version is released (via `specflowctl rule release-version`), stable units that bind the old rule ref and have `Next Command=unit_fork` are automatically forked to candidate with `candidate_intent=change`.

### Auto-Fork Procedure

1. Detect stable units that bind the old rule ref and have `Next Command=unit_fork`.
2. Create a candidate copy with bumped patch version. Rewrite Markdown document references within the candidate main Spec body from stable appendix paths (`s_unit_*`) to candidate appendix paths (`c_unit_*`). Update `docs/specs/repository_mapping.md` spec_files to reference the candidate Spec path.
3. Copy all stable appendices to candidate with frontmatter rewrites and Markdown document ref rewrites (stable appendix paths `s_unit_*` → candidate appendix paths `c_unit_*` within the copied appendix files).
4. Remove process artifacts (check_work, check result, verify result, stable_verify result) and agent-internal artifacts (plan) from any previous lifecycle round for the target unit.
5. Invoke command close with `unit_fork:candidate_created`.

### Governance Rules

1. Auto-fork applies only to stable units with `Next Command=unit_fork`. If any stable consumer of the old rule ref is in a different lifecycle state, the release-version operation fails with an error listing the non-compliant units. Resolve these units' lifecycle state before retrying. (See `tooling/internal/rulesync/release.go` for the tooling constraint.)
2. Stable appendices are copied — not moved — so stable truth remains intact for non-auto-forked consumers.
3. The auto-forked candidate inherits the old rule ref's behavior truth; the unit specification content is preserved but Markdown document references are rewritten from stable to candidate paths per the standard fork contract (see `framework/lifecycle/unit_init_new_fork.md`).
3b. **Intent conflict guard** — `release-version` writes `candidate_intent=change` on the auto-forked candidate. If the stable verify result for the target unit records a decision requiring `candidate_intent=repair` (`controlled_repair_required` or `truth_text_change_required`, mapping per `framework/lifecycle/unit_init_new_fork.md` unit_fork candidate_intent determination), the release-version operation must fail and report the conflict. The unit must be resolved before release-version can proceed. (See `tooling/internal/rulesync/release.go` `StableVerifyCandidateIntentRequirement` for the tooling constraint.)
3c. **Auto-forked `source_basis`** — `release-version` writes `source_basis=new_design` on the auto-forked candidate. The spec content is inherited from stable-layer truth (originally authored from design intent) with no new implementation observation. The `unit_check` gate will verify that the declared `source_basis` is consistent with the evidence appendix.
4. After auto-fork, the candidate unit follows the standard `unit_check` → `unit_verify` → `unit_promote` lifecycle.
5. The `rule_sync` impact handoff runs after writeback, so downstream invalidation and fallback rules apply.

### Tooling Entry

```text
specflowctl rule release-version --rule-id <id> --from-ref <old-ref> --to-ref <new-ref>
```

See `tooling/internal/rulesync/release.go` for implementation details.

## 9. Rejection

`rule_sync` must reject:

1. scenario consumers
2. scenario paths
3. scenario commands
4. `object-type=scenario`
5. any attempt to use `bound_objects` as consumer truth

## 10. Output Contract

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
