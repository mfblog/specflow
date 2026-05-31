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
4. `active_plan_file_ref` and `active_plan_fingerprint`
5. `retirement_evidence_matrix`
6. `unit_appendix_snapshot`
7. `unit_snapshot`
8. `rule_snapshot`
9. independent evaluation receipt fields
10. conditional freshness reuse receipt fields when accepted `text_drift` keeps evidence reusable

`retirement_evidence_matrix` must be literal `none` when the active plan has `retirement_targets: none`.
When the active plan lists retirement targets, every target id must appear exactly once with `result: pass`, `mainline_dependency: not_required`, and durable `evidence_refs`.
The verify result proves planned retirement targets; it does not authorize automatic code deletion.

These files are process evidence, not behavior truth.

Stable promotion summaries are historical promotion records.
They are not current implementation-alignment evidence after later code changes.

Current stable verification evidence belongs in:

```text
docs/specs/_stable_verify_result/unit/{unit}.md
```
