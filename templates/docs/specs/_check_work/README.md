# Unit Check Checklist Directory

This directory stores optional resume checklists for `unit_check`.

Allowed object type:

1. `unit`

Supported path:

```text
docs/specs/_check_work/unit/{unit}.md
```

These files are not Specs, not behavior truth, and not downstream pass gates.
They record only the current `unit_check` round's checklist progress, input fingerprints, finding references, blocked reason, and resume position.

The handoff from `unit_check` feeds into `unit_impl`:

```text
docs/specs/_check_result/unit/{unit}.md
```

Tooling may maintain only mechanical fields:

1. UTC timestamps
2. baseline checklist skeleton
3. input fingerprints
4. stale checklist item marks
5. structural validation

Tooling must not write checklist pass judgments, finding content, severity, or the final `pass`, `blocked`, or `fix_required` conclusion.
