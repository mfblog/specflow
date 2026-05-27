# Governance Review

Framework review checks whether specFlow governance is coherent and operable.

## Default Mode

Ordinary `spec_flow_review` and governance review use `scoped_review` by default.

Read `framework/governance/review_scope.md` first. Review only the changed files, direct owner files, necessary boundary refs, and minimal convergence refs needed to answer the request.

`scoped_review` does not use `_governance_review/` run-state, baseline slice tables, dynamic slice tables, or score-state tables by default.

Plain exact `spec_flow_design_review` is not scoped.
It always delegates to `framework/spec_flow_design_review.md` and runs the default full-scope design-baseline review.
There is no narrowed or scoped `spec_flow_design_review` mode.

## Entries

1. `spec_flow_review` checks mechanism correctness.
2. `spec_flow_design_review` checks design quality and agent operability.

Plain exact `spec_flow_review` routes through this file first and remains scoped.
The only full-scope mechanism review entry is exact `spec_flow_review:full`.

Plain exact `spec_flow_design_review` routes through this file first, then directly delegates to `framework/spec_flow_design_review.md`.

When the entry is exact `spec_flow_review:full`, `spec_flow_review` delegates to `framework/spec_flow_review.md`.

## Review Layout

Deep-audit review tooling is layout-aware.

Supported layouts:

1. `installed_project`
   - framework inputs live under `specflow/framework/`
   - templates live under `specflow/templates/`
   - tooling lives under `specflow/tooling/`
   - project-instance compatibility reviews real `docs/specs/` files
2. `source_repo`
   - framework inputs live under local `framework/`
   - templates live under local `templates/`
   - tooling lives under local `tooling/`
   - project-instance compatibility reviews template bootstrap compatibility and does not require real `docs/specs/` instance files

`specflowctl review ... --layout auto` detects the layout. Callers may pass `--layout installed` or `--layout source` to force one layout.

Review layout applies to exact `spec_flow_review:full` mechanism deep audit and every `spec_flow_design_review`.
It does not widen ordinary `spec_flow_review` scoped review.

## Active Scope

Default `spec_flow_review` scoped review uses the layered framework structure:

1. `core/`
2. `lifecycle/`
3. `governance/`
4. `operations/`
5. guidance skills
6. templates
7. tooling contracts and source

`framework/spec_flow_review.md` is the mechanism deep-audit owner and is not ordinary default context for scoped review.
`framework/spec_flow_design_review.md` is the ordinary owner for every `spec_flow_design_review`.
