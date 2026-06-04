# Check Result Directory

This directory stores unit check pass snapshots.
In-progress `unit_check` checklist work may be stored under `docs/specs/_check_work/unit/{unit}.md`; that checklist file is not a pass gate.

Allowed object type:

1. `unit`

Supported path:

```text
docs/specs/_check_result/unit/{unit}.md
```

The snapshot records:

1. current unit truth ref, fingerprint, and acceptance behavior fingerprint
2. accepted acceptance item set
3. `unit_appendix_snapshot`
4. `unit_snapshot`
5. `rule_snapshot`
6. independent evaluation receipt fields
7. conditional freshness reuse receipt fields when accepted `text_drift` keeps evidence reusable

These files are process evidence, not behavior truth.

A pass snapshot means the full unit package is clear and internally consistent enough for package-bounded delta planning.
