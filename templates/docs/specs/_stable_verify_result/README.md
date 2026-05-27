# Stable Verify Result Directory

This directory stores current stable implementation-alignment evidence written by `unit_stable_verify`.

Allowed object type:

1. `unit`

Supported path:

```text
docs/specs/_stable_verify_result/unit/{unit}.md
```

Each stable verify result records:

1. stable truth ref, fingerprint, and acceptance behavior fingerprint
2. `blocking_summary` and `coverage_summary`
3. repository mapping snapshot
4. acceptance item set
5. acceptance item evidence matrix
6. `unit_appendix_snapshot`
7. `unit_snapshot`
8. `rule_snapshot`
9. implementation surface refs and evidence refs
10. independent evaluation receipt fields
11. conditional freshness reuse receipt fields when accepted `text_drift` keeps evidence reusable

This file is process evidence, not behavior truth.

It is not:

1. a stable promotion summary
2. a substitute for `docs/specs/units/stable/s_unit_{unit}.md`
3. a slice work-state file
4. a separate checklist file
