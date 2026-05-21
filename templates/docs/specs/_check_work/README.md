# Unit Check Work State Directory

This directory stores intermediate work-state files for `unit_check`.

Allowed object type:

1. `unit`

Supported path:

```text
docs/specs/_check_work/unit/{unit}.md
```

These files are not Specs, not behavior truth, and not downstream pass gates.
They record only the current `unit_check` round's slice progress, input fingerprints, finding references, blocked reason, and resume position.

The only handoff gate from `unit_check` to `unit_plan` remains:

```text
docs/specs/_check_result/unit/{unit}.md
```

Tooling may maintain only mechanical fields:

1. UTC timestamps
2. baseline slice skeleton
3. input fingerprints
4. stale slice marks
5. structural validation

Tooling must not write slice pass judgments, finding content, severity, or the final `pass`, `blocked`, or `fix_required` conclusion.
