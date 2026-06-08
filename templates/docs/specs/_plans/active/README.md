# Candidate Active Plans (Agent-Internal)

This directory stores agent-internal implementation plan snapshots. These files are **not SpecFlow process evidence** and are not consumed by SpecFlow lifecycle gates.

## Status

`unit_plan` is no longer a SpecFlow-governed command. Agents handle planning internally. Plan files are optional agent workspace artifacts.

`unit_verify` and `unit_promote` do NOT require or consume active plan files. When reference is desired, verify evidence may optionally include `active_plan_file_ref` and `active_plan_fingerprint`.

## Guidance for Agents

An agent-internal active plan may record:
- `spec_file_ref`, `spec_version_ref`, `spec_fingerprint`
- `acceptance_behavior_fingerprint`
- `execution_surface_plan` — organized around changed execution surfaces
- `planned_change_scope` — delta scopes as `pcs.<slug>` items
- `retirement_targets` — old paths, wrappers, or dependencies to retire
- `package_constraint_review` — how the delta respects package constraints
- `implementation_tasks` — closeable execution slices

These fields are recommendations, not requirements. Agents may structure plans however their framework prefers.

## Lifecycle

- `unit_fork` may delete the previous round's `active/{unit}.md`.
- `unit_promote` may delete the corresponding `active/{unit}.md`.
- When `Candidate=no`, `active/{unit}.md` should not remain.
