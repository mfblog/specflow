<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit}

## STATUS
- Stage: stable (idle) | Next: unit_fork
- Layer: stable

## GUIDANCE
This unit is stable with no active candidate round. Your action depends on the user's goal.

> **Pre-check:** Read `docs/specs/_status.md` to confirm this unit's row (Stable=yes, Candidate=no, Active=stable, Next=unit_fork).

**=== A. Behavior change (modify stable truth) ===**
1. Read `docs/specs/units/stable/s_unit_{unit}.md` to understand current stable truth.
2. Run `unit_fork:{unit}` to create a candidate branch.
3. Follow `framework/lifecycle/unit_init_new_fork.md` for the fork procedure (write candidate spec, command close).
4. After fork, proceed through candidate lifecycle: check → impl → verify → promote.

**Close:** `specflowctl command close --command unit_fork --object-type unit --object {unit} --outcome forked`

**=== B. Alignment check (verify impl still matches spec) ===**
1. Read `docs/specs/units/stable/s_unit_{unit}.md`.
2. Run `unit_stable_verify:{unit}`.
3. Follow `framework/lifecycle/unit_stable_verify.md` for the full verification procedure.
4. Possible outcomes: aligned, controlled_repair_required, controlled_change_required, small_repair_required, truth_rejudge_required, evidence_incomplete.
5. Write result to `docs/specs/_stable_verify_result/unit/{unit}.md` per `framework/process_snapshot_contract.md`.

**Close:** `specflowctl command close --command unit_stable_verify --object-type unit --object {unit} --outcome <outcome>`
**After close:** aligned → done; repair/change → run `unit_fork:{unit}`

**=== C. Implementation-only (no spec change) ===**
Apply when the request is pure implementation, refactoring, testing, or performance optimization:
- Modify code under `{impl_paths}/**` and `{test_paths}/**`.
- Do NOT touch spec files, rule files, or `_status.md`.
- If you discover behavior or spec needs to change: stop, run `unit_fork:{unit}` first.
- No command close needed.

## WRITES (owned by this unit)
- {impl_paths}/**
- {test_paths}/**

## READS (read-only context)
- docs/specs/_status.md
- docs/specs/units/stable/s_unit_{unit}.md + appendices
- docs/specs/repository_mapping.md
- framework/lifecycle/overview.md
- framework/core/object_model.md
- {rule_refs_paths}
- {unit_refs_paths}

## BLOCKED
- Modifying stable-layer truth (must go through fork)
- Modifying _status.md (use command close)
- Any rule files
- Other units' specs or status

## CLOSE
stable_idle state does not require command close. Enter a candidate round via unit_fork, or check alignment via unit_stable_verify.

## Next Steps
Need to change spec → run unit_fork:{unit}
Need to check alignment → run unit_stable_verify:{unit}
Change code directly → follow implementation_only path
Re-run: specflowctl context card
