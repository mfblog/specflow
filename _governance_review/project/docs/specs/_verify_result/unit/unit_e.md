# verify

```yaml
object_type: unit
object_ref: unit_e
gate: unit_verify
decision: pass
allow_next: true
next_command: unit_promote
blocking_summary: none
coverage_summary: current candidate
truth_layer_ref: candidate
truth_file_ref: docs/specs/units/candidate/c_unit_unit_e.md
truth_version_ref: c_unit_unit_e@0.2.0
truth_fingerprint: 1360e2f454b41e17b015b55e994c8c31a331c160c70d8480dbc93dcf6dad2e7f
acceptance_behavior_fingerprint: dffd242a5ff12787cb6cdbc457c89c07d8484eea732fa7f702b7836422f3f85a
acceptance_item_set:
  - id: unit_e.core
    verification_surface: integration
    not_runnable_yet: no
unit_appendix_snapshot:
  - file_ref: docs/specs/units/candidate/appendix/c_unit_unit_e_evidence.md
    fingerprint: 2b6edcc2a1bb96a1778531821f0761c419d2505221241d7d74ebbcbd0f2c7027
unit_snapshot: none
rule_snapshot:
  - rule_id: default
    layer: stable
    file_ref: docs/specs/rules/stable/s_g_rule_default.md
    version_ref: s_g_rule_default@1.0.0
    fingerprint: 25c810eb452db55c0b748641feedeec5dc792340a663e37bc4aa17f2bf9b90db
acceptance_item_evidence_matrix:
  - id: unit_e.core
    status: pass
    evidence_refs: go test ./...
retirement_evidence_matrix: none
package_delta_verification:
  - planned_change_scope_id: pcs.core
    result: pass
    evidence_refs: go test ./...
```
