# Candidate Active Plans

This directory stores the consumable implementation plans for candidate progression.

Rules:

1. Each unit normally has one `active/{unit}.md`.
2. These files are not formal Specs and are not behavior sources of truth.
3. Each active plan file carries these formal planning sections:
   - `Execution Surface Plan`
   - `Retirement Targets`
   - `Verification Targets`
4. Each active plan file also carries:
   - `Implementation Tasks`
   - current-round progress, blockers, and verification focus
5. `Execution Surface Plan` must be organized around the changed execution surfaces of the round rather than one forced whole-unit path.
6. Each execution surface must name its current known path, target path, and first cutover slices.
7. `Retirement Targets` must name which retired paths, helpers, patches, wrappers, or equivalent dependencies should stop being required for the round to close.
8. `Verification Targets` must name which retirement goals `unit_verify` must explicitly prove.
9. `Verification Targets` must also reference the candidate Spec acceptance item `id` values they prove.
10. `Implementation Tasks` should be organized as closeable execution slices rather than one undifferentiated task block.
11. Each slice should at minimum name:
   - objective
   - file scope
   - dependencies
   - verification action
   - done condition
   - current status
12. Each active plan file must also record:
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `acceptance_behavior_fingerprint`
   - `stable_candidate_diff_refs`
   - `implementation_gap_refs`
   - `unit_appendix_snapshot`
   - `rule_snapshot`
   - `acceptance_item_plan_coverage`
   - `retirement_targets`
   - independent evaluation receipt fields
   - conditional freshness reuse receipt fields when accepted `text_drift` keeps the plan reusable
13. `acceptance_item_plan_coverage` must map each current-gate acceptance item `id` to implementation slices or `Verification Targets`.
14. `stable_candidate_diff_refs` must be literal `none`, or a semicolon-delimited durable ref scalar. When stable truth exists, it must cite both the stable and candidate main Specs.
15. `implementation_gap_refs` must be literal `none`, or a semicolon-delimited durable ref scalar naming the repository mapping and implementation refs inspected for current entry points, main paths, rendering/API/generation paths, and gaps.
16. `retirement_targets` must be literal `none`, or a YAML list of `rt.<slug>` items with `target_ref`, `target_kind`, `retirement_method`, `verification_action`, and `acceptance_item_ids`.
    `acceptance_item_ids` must be a single comma-separated scalar of current acceptance item ids; it must not be a YAML list or semicolon-delimited value.
17. For `candidate_intent: change` with `source_basis: replacement`, `retirement_targets` must not be `none`; the plan must name old primary paths, new primary paths, cutover slices, and retirement target ids.
18. If a current-gate acceptance item is not covered by an implementation slice or verification target, `active/{unit}.md` is not consumable by `unit_impl` or `unit_verify`.
19. If `retirement_targets` is missing, malformed, or references unknown acceptance item ids, `active/{unit}.md` is not consumable by `unit_impl`, `unit_verify`, or `unit_promote`.
20. `active/{unit}.md` does not carry the gate decision of `unit_check` or the promotion decision of `unit_verify`.
21. `unit_plan` writes or updates `active/{unit}.md` only when the round is `plan-ready`.
22. If planning is still blocked or still inside a decision checkpoint, `unit_plan` must keep updating `draft/{unit}.md` instead of rewriting `active/{unit}.md`.
23. `unit_impl` may keep writing progress, blockers, verification notes, and retirement progression, but must not rewrite binding fields.
24. For each advanced slice, `unit_impl` should write back at minimum:
   - `execution_surface`
   - `cutover_result`
   - `retirement_result`
   - `verification_note`
25. `unit_verify` still requires a current valid `active/{unit}.md`.
26. `unit_verify` must consume `Execution Surface Plan`, `Retirement Targets`, `Verification Targets`, `acceptance_item_plan_coverage`, and `retirement_targets` as part of the round's formal verification input.
27. If candidate truth changes, deterministic validation must classify the freshness impact before choosing fallback.
28. Accepted `text_drift` can remain reusable; semantic, acceptance, dependency, schema, and unknown drift cannot.
29. If only the active plan is missing, malformed, not tool-valid, or missing required coverage while the check gate still covers current truth, the recovery layer is `plan_layer`; the flow returns to `unit_plan` without deleting a still-valid check gate.
30. `unit_fork` must delete the previous round's `active/{unit}.md`.
31. `unit_promote` must delete the corresponding `active/{unit}.md`.
32. When `Candidate=no`, `active/{unit}.md` should not remain.
33. This README is also constrained by `specflow/framework/candidate_handoff_contract.md`.
34. Snapshot fields in this file must use the fixed definitions from `specflow/framework/process_snapshot_contract.md`.
