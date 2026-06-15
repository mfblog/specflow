---
rule_id: security
rule_scope: global
layer: stable
rule_version: 1.0.0
---
# Global Security Rule (Stable)
All units must enforce basic security measures.
## Rule Body
1. Every endpoint must use TLS.
2. All requests must include unique correlation ID.
3. Sensitive data must be encrypted at rest.
