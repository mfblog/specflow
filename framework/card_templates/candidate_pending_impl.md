<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit}

## STATUS
- Stage: pending_impl | Next: unit_verify
- Layer: candidate | Intent: {intent}
- Notes: pending_impl
{check_status_line}

## GUIDANCE
You are in the implementation phase. Implement each acceptance criterion from the candidate spec.

> **Pre-check:** Read `docs/specs/_status.md`. Confirm this unit has Next=unit_verify, Notes=pending_impl.

**Execution:**
1. Read `docs/specs/units/candidate/c_unit_{unit}.md` + appendices to understand each acceptance criterion.
2. Follow `framework/lifecycle/unit_impl.md` for trigger command semantics.
3. Write code under `{impl_paths}/**` and `{test_paths}/**` to satisfy each criterion.
4. Conversation and iteration are normal here — you may discuss with the user.

**If you find spec issues during implementation:**
  1. Fix the candidate spec at `docs/specs/units/candidate/c_unit_{unit}.md`.
  2. Run `unit_check:{unit}` to re-validate the changed spec.
  3. On pass, resume implementation.
  4. On fail, fix spec and re-run unit_check.

**Do NOT** advance lifecycle state. This is a trigger command — no command close.

**Terminal action:** When all acceptance criteria are satisfied → run `unit_verify:{unit}`.

> If the current request does not involve spec changes (pure implementation, refactoring, testing, performance optimization):
> You may modify implementation code directly. If you find the spec needs to change, stop and follow the spec-issue path above.

## WRITES (owned by this unit)
- docs/specs/units/candidate/c_unit_{unit}.md
- docs/specs/units/candidate/appendix/c_unit_{unit}_*.md
- {impl_paths}/**
- {test_paths}/**

## READS (read-only context)
- docs/specs/_status.md
- docs/specs/repository_mapping.md
- framework/lifecycle/unit_impl.md
- framework/spec_writing_guide.md
- {rule_refs_paths}  (applicable rules, do not modify)
- {unit_refs_paths}  (dependent stable units, do not modify)

## BLOCKED
- _status.md (use command close)
- Any rule files (use rule governance: specflowctl --object-type rule)
- Any other unit's spec or status
- Stable-layer truth (must fork first)
- rule_refs/unit_refs targets (read-only)
- unit_promote (must verify first)

## CLOSE
pending_impl state uses the unit_impl trigger command; no command close. On completion, run unit_verify:{unit}.

## Next Steps
When all acceptance criteria are satisfied → run unit_verify:{unit}
Re-run: specflowctl context card
