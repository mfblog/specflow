# Candidate Plans (Agent-Internal)

This directory stores plan-family files that are **agent-internal artifacts**. They are not SpecFlow process evidence and are not consumed by SpecFlow lifecycle gates.

## Structure

`_plans/` is divided into:
- `draft/` — work-in-progress planning notes
- `active/` — completed internal plan snapshots

## Status

Plan files are no longer SpecFlow-governed. The `unit_plan` command has been removed from the SpecFlow lifecycle. `unit_impl` is a lifecycle state set by `unit_check pass` close, not a user command. Agents handle planning and implementation internally.

SpecFlow lifecycle commands (`unit_verify`, `unit_promote`) do NOT require or consume plan files. Plan fields in verify evidence (`active_plan_file_ref`, `retirement_evidence_matrix`, `package_delta_verification`) are optional.

## Guidance for Agents

- Plans may be structured however the agent framework prefers.
- Plan files may be kept, discarded, or updated at the agent's discretion.
- If an agent chooses to record retirement targets or planned change scope in a plan, it may optionally reference them in verify evidence.
- Stale plan files from prior lifecycle rounds may be cleaned up or preserved as historical context.
