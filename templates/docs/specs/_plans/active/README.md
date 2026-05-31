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
   - `unit_appendix_snapshot`
   - `rule_snapshot`
   - `acceptance_item_plan_coverage`
   - `retirement_targets`
   - independent evaluation receipt fields
   - conditional freshness reuse receipt fields when accepted `text_drift` keeps the plan reusable
13. `acceptance_item_plan_coverage` must map each current-gate acceptance item `id` to implementation slices or `Verification Targets`.
14. `retirement_targets` must be literal `none`, or a YAML list of `rt.<slug>` items with `target_ref`, `target_kind`, `retirement_method`, `verification_action`, and `acceptance_item_ids`.
    `acceptance_item_ids` must be a single comma-separated scalar of current acceptance item ids; it must not be a YAML list or semicolon-delimited value.
15. If a current-gate acceptance item is not covered by an implementation slice or verification target, `active/{unit}.md` is not consumable by `unit_impl` or `unit_verify`.
16. If `retirement_targets` is missing, malformed, or references unknown acceptance item ids, `active/{unit}.md` is not consumable by `unit_impl`, `unit_verify`, or `unit_promote`.
17. `active/{unit}.md` does not carry the gate decision of `unit_check` or the promotion decision of `unit_verify`.
18. `unit_plan` writes or updates `active/{unit}.md` only when the round is `plan-ready`.
19. If planning is still blocked or still inside a decision checkpoint, `unit_plan` must keep updating `draft/{unit}.md` instead of rewriting `active/{unit}.md`.
20. `unit_impl` may keep writing progress, blockers, verification notes, and retirement progression, but must not rewrite binding fields.
21. For each advanced slice, `unit_impl` should write back at minimum:
   - `execution_surface`
   - `cutover_result`
   - `retirement_result`
   - `verification_note`
22. `unit_verify` still requires a current valid `active/{unit}.md`.
23. `unit_verify` must consume `Execution Surface Plan`, `Retirement Targets`, `Verification Targets`, `acceptance_item_plan_coverage`, and `retirement_targets` as part of the round's formal verification input.
24. If candidate truth changes, deterministic validation must classify the freshness impact before choosing fallback.
25. Accepted `text_drift` can remain reusable; semantic, acceptance, dependency, schema, and unknown drift cannot.
26. If only the active plan is missing, malformed, not tool-valid, or missing required coverage while the check gate still covers current truth, the recovery layer is `plan_layer`; the flow returns to `unit_plan` without deleting a still-valid check gate.
27. `unit_fork` must delete the previous round's `active/{unit}.md`.
28. `unit_promote` must delete the corresponding `active/{unit}.md`.
29. When `Candidate=no`, `active/{unit}.md` should not remain.
30. This README is also constrained by `specflow/framework/candidate_handoff_contract.md`.
31. Snapshot fields in this file must use the fixed definitions from `specflow/framework/process_snapshot_contract.md`.
