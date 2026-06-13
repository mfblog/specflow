---
rule_id: default
rule_scope: global
layer: stable
rule_version: 1.0.0
---
# Global Default Rule
All units must produce structured audit logs.
## Rule Body
1. Every operation must log: timestamp, operation name, result.
2. Logs must be in JSON format.
3. Error logs must include stack trace.
