<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit}

## STATUS
- Stage: unit_check | Next: unit_verify (on pass)
- Layer: candidate | Intent: {intent}

## GUIDANCE
Check whether the candidate truth is clear and complete enough to enter implementation.

> **Pre-check:** Read `docs/specs/_status.md`. Confirm Next=unit_check (or Next=unit_verify with Notes=pending_impl for re-validation path).

**Execution steps:**
1. Create or update `docs/specs/_check_work/unit/{unit}.md` with the 8-item checklist (see `framework/lifecycle/unit_check.md` "What This Step Does" items 1-8).
2. Evaluate each item against the candidate spec `docs/specs/units/candidate/c_unit_{unit}.md`.
3. Write the check result to `docs/specs/_check_result/unit/{unit}.md` per `framework/process_snapshot_contract.md` format.
4. Independent review is required. Follow `framework/operations/entry_routing.md` Independent Review Stop.
5. All 8 items must pass for a `pass` outcome. Any failure is `fix_required`.

**Close:** `specflowctl command close --command unit_check --object-type unit --object {unit} --outcome <outcome>`
**After close:** pass → run `unit_impl:{unit}` then `unit_verify:{unit}`; fix_required → fix spec and re-run.

**Re-validation exception:** If this is a re-validation during implementation (Next=unit_verify, Notes=pending_impl), the fingerprint in `_check_result` is used for differential re-check — only re-check items affected by the spec change. See `framework/lifecycle/unit_check.md` precondition exception.

> If the current request does not involve spec changes (pure implementation, refactoring, testing, performance optimization):
> Not applicable — the current stage is checking spec quality, not implementation.

## WRITES (owned by this unit)
- docs/specs/_check_result/unit/{unit}.md
- docs/specs/_check_work/unit/{unit}.md
- docs/specs/units/candidate/c_unit_{unit}.md (fixing spec issues)

## READS (read-only context)
- docs/specs/_status.md
- docs/specs/units/candidate/c_unit_{unit}.md + appendices
- Referenced stable-layer truth and rules
- docs/specs/_check_result/unit/{unit}.md (fingerprint on re-validation)
- framework/lifecycle/unit_check.md
- framework/process_snapshot_contract.md
- framework/spec_writing_guide.md
- {rule_refs_paths}
- {unit_refs_paths}

## BLOCKED
- _status.md (use command close)
- Implementation files
- Stable-layer truth
- Any rule files
- Other units' specs or status

## CLOSE
specflowctl command close --command unit_check --object-type unit --object {unit} --outcome <outcome>

## Next Steps
pass → run unit_impl:{unit}, then unit_verify:{unit}
fix_required → fix spec, re-run unit_check:{unit}
blocked → ask the user
Re-run: specflowctl context card
