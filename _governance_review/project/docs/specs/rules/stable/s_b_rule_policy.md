---
rule_id: policy
rule_scope: bound
layer: stable
rule_version: 1.0.0
---
# Policy Rule
Units bound to this rule must validate all external inputs.
1. All user-facing inputs must be sanitized.
2. All API inputs must be validated against schema.
3. Invalid input must return structured error responses.
