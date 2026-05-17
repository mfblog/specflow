# Check Result Directory

This directory stores unit check pass snapshots.

Allowed object type:

1. `unit`

Supported path:

```text
docs/specs/_check_result/unit/{unit}.md
```

The snapshot records:

1. current unit truth ref and fingerprint
2. accepted acceptance item set
3. `unit_appendix_snapshot`
4. `unit_snapshot`
5. `rule_snapshot`

These files are process evidence, not behavior truth.
