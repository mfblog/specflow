# check

```yaml
object_type: unit
object_ref: unit_c
gate: unit_check
decision: pass
allow_next: true
next_command: unit_check
blocking_summary: none
coverage_summary: current candidate
truth_layer_ref: candidate
truth_file_ref: docs/specs/units/candidate/c_unit_unit_c.md
truth_version_ref: c_unit_unit_c@0.1.0
truth_fingerprint: 0cd8cf41a9a34797ea21a5225e99bbf76add05f849214a996c0e1cab4bf2382b
acceptance_behavior_fingerprint: 5443deec8c3d3046bbecf4f35c5be611a8a39f73393c4a6b2be739222241e9ca
acceptance_item_set:
  - id: unit_c.core
    verification_surface: integration
    not_runnable_yet: no
unit_appendix_snapshot:
  - file_ref: docs/specs/units/candidate/appendix/c_unit_unit_c_design.md
    fingerprint: caa77ef8b021a805c79d84eaf9e022a8c057aa599d223327b65337311b8733aa
unit_snapshot:
  - unit: unit_a
    layer: stable
    file_ref: docs/specs/units/stable/s_unit_unit_a.md
    version_ref: s_unit_unit_a@1.0.0
    fingerprint: 23e90df84ec7a643ba9c21fd3f80f4be3e13670986dd08ea8c9f1fdbb9c17542
rule_snapshot:
  - rule_id: default
    layer: stable
    file_ref: docs/specs/rules/stable/s_g_rule_default.md
    version_ref: s_g_rule_default@1.0.0
    fingerprint: 25c810eb452db55c0b748641feedeec5dc792340a663e37bc4aa17f2bf9b90db
  - rule_id: policy
    layer: stable
    file_ref: docs/specs/rules/stable/s_b_rule_policy.md
    version_ref: s_b_rule_policy@1.0.0
    fingerprint: 22a621d0751bade2389ab54cb184fe070d06c1edd709a07ee1ce3c8969d27031
  - rule_id: security
    layer: stable
    file_ref: docs/specs/rules/stable/s_g_rule_security.md
    version_ref: s_g_rule_security@1.0.0
    fingerprint: 9ab0cf4309dc92952ba5dc45c27b1f01167822bcff828b93ce0919788d2da234
evaluation_mode: independent
reviewer_result: pass
reviewer_context: minimal_context
review_input_refs: unit_check_pass;docs/specs/_independent_evaluation/requests/unit/unit_c/unit_check_pass.md;docs/specs/units/candidate/c_unit_unit_c.md
review_findings: none
human_decision_refs: none
```
