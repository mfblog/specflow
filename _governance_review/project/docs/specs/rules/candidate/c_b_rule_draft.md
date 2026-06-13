---
rule_id: draft
rule_scope: bound
layer: candidate
rule_version: 0.1.0
---
# Draft Rule
Proposed retry policy for transient failures.
1. All upstream calls must retry up to 3 times.
2. Exponential backoff: 1s, 2s, 4s.
3. Circuit breaker after 5 consecutive failures.
