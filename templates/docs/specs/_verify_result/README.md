# Verify Result Directory

This directory stores candidate unit verify pass snapshots and stable promotion coverage summaries.

Allowed object type:

1. `unit`

Supported paths:

```text
docs/specs/_verify_result/unit/{unit}.md
docs/specs/_verify_result/stable/unit/{unit}.md
```

The candidate verify pass snapshot records:

1. current unit truth ref, fingerprint, and acceptance behavior fingerprint
2. accepted acceptance item set
3. acceptance item evidence matrix
4. `unit_appendix_snapshot`
5. `unit_snapshot`
6. `rule_snapshot`
7. independent evaluation receipt fields
8. conditional freshness reuse receipt fields when accepted `text_drift` keeps evidence reusable

These files are process evidence, not behavior truth.

Stable promotion summaries are historical promotion records.
They are not current implementation-alignment evidence after later code changes.

Current stable verification evidence belongs in:

```text
docs/specs/_stable_verify_result/unit/{unit}.md
```
