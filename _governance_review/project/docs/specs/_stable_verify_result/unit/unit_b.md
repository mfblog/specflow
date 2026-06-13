# stable verify

```yaml
object_type: unit
object_ref: unit_b
gate: unit_stable_verify
decision: aligned
allow_next: true
next_command: unit_fork
blocking_summary: none
coverage_summary: current stable implementation
truth_layer_ref: stable
truth_file_ref: docs/specs/units/stable/s_unit_unit_b.md
truth_version_ref: s_unit_unit_b@1.0.0
truth_fingerprint: bd18e8333e52cc2325f1570ae33e8ab53874e28a6b9b1e065dae9a5a13cd30f7
acceptance_behavior_fingerprint: 7ce4f46a63af839e474c2c42599b8470805ded2ae78f73f78cf535216d92b391
acceptance_item_set:
  - id: unit_b.core
    verification_surface: integration
    not_runnable_yet: no
unit_appendix_snapshot: none
unit_snapshot: none
rule_snapshot:
  - rule_id: default
    layer: stable
    file_ref: docs/specs/rules/stable/s_g_rule_default.md
    version_ref: s_g_rule_default@1.0.0
    fingerprint: 25c810eb452db55c0b748641feedeec5dc792340a663e37bc4aa17f2bf9b90db
acceptance_item_evidence_matrix:
  - id: unit_b.core
    status: pass
    evidence_refs: go test ./...
implementation_surface_refs: src/unit_b
evidence_refs: go test ./...
repository_mapping_snapshot:
  file_ref: docs/specs/repository_mapping.md
  version_ref: repository_mapping@0.1.0
  fingerprint: 9b615f039876ca4afccd6cc965bf51fe4f51eaae1c864ec1d1e7c0bacfd5efb7
```
