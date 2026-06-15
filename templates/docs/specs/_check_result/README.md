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
3. `blocking_summary` — summary of any issues that prevent immediate progression (or `none` when clear)
4. `coverage_summary` — summary of what was checked and the coverage scope
5. `unit_appendix_snapshot`
6. `unit_snapshot`
7. `rule_snapshot`
8. independent evaluation receipt fields
9. conditional freshness reuse receipt fields when accepted `text_drift` keeps evidence reusable

These files are process evidence, not behavior truth.

A pass snapshot means the full unit package is clear and internally consistent enough for downstream verification work.
