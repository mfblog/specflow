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
3. acceptance item evidence matrix with per-item `evidence_refs` and optional `scope_verification`
4. `unit_appendix_snapshot`
5. `unit_snapshot`
6. `rule_snapshot`
7. independent evaluation receipt fields
8. conditional freshness reuse receipt fields when accepted `text_drift` keeps evidence reusable

When an acceptance item declares `affects` in its definition, the evidence matrix item should include `scope_verification` recording the verification result for each affected file, appendix, rule, and dependency. All scope items must pass for the acceptance item to be promotion-ready.

When the agent created an internal plan and chooses to reference it, these optional fields may also appear:

- `active_plan_file_ref` and `active_plan_fingerprint` — reference to the agent-internal plan
- `retirement_evidence_matrix` — retirement target evidence (optional, `none` when no plan or no retirement targets)
- `package_delta_verification` — planned change scope evidence (optional, `none` when no plan)

When present, retirement evidence follows: each target id must appear exactly once with `result: pass`, `mainline_dependency: not_required`, and durable `evidence_refs`.
When present, package delta evidence requires every item to use `result: pass` with durable `evidence_refs`.

Each executable acceptance item in `acceptance_item_evidence_matrix` must record `status: pass` and durable `evidence_refs` before promotion readiness can close.
Items marked `not_runnable_yet: yes` in current truth must record `status: not_runnable_yet`.
Generic test success, missing old strings, present new files, or present new fields are not sufficient by themselves for semantic replacement evidence.

These files are process evidence, not behavior truth.

Stable promotion summaries are historical promotion records.
They are not current implementation-alignment evidence after later code changes.

Current stable verification evidence belongs in:

```text
docs/specs/_stable_verify_result/unit/{unit}.md
```
