# Verify Result Directory

This directory stores unit verify snapshots and stable promotion coverage summaries.

Allowed object type:

1. `unit`

Supported paths:

```text
docs/specs/_verify_result/unit/{unit}.md
docs/specs/_verify_result/stable/unit/{unit}.md
```

The candidate verify snapshot records:

1. current unit truth ref and fingerprint
2. accepted acceptance item set
3. acceptance item evidence matrix
4. `unit_appendix_snapshot`
5. `unit_snapshot`
6. `rule_snapshot`

These files are process evidence, not behavior truth.
