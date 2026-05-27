# Governance Review

Framework review checks whether specFlow governance is coherent and operable.

## Default Mode

Ordinary governance or design review uses `scoped_review` by default.

Read `framework/governance/review_scope.md` first. Review only the changed files, direct owner files, necessary boundary refs, and minimal convergence refs needed to answer the request.

`scoped_review` does not use `_governance_review/` run-state, baseline slice tables, dynamic slice tables, or score-state tables by default.

## Entries

1. `spec_flow_review` checks mechanism correctness.
2. `spec_flow_design_review` checks design quality and agent operability.

Plain exact entries route through this file first. They remain scoped unless the user explicitly asks for `full-scope`, `baseline`, `deep audit`, release-level governance audit, `resumable review`, or run-state-backed review.
For mechanism review, exact `spec_flow_review:full` is also explicit deep-audit intent.

When deep audit is explicit:

1. `spec_flow_review` delegates to `framework/spec_flow_review.md`.
2. `spec_flow_design_review` delegates to `framework/spec_flow_design_review.md`.

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

Review layout applies to explicit deep audit and run-state-backed review only. It does not widen ordinary `scoped_review`.

## Active Scope

Default review scope uses the layered framework structure:

1. `core/`
2. `lifecycle/`
3. `governance/`
4. `operations/`
5. guidance skills
6. templates
7. tooling contracts and source

`spec_flow_review.md` and `spec_flow_design_review.md` are deep-audit owners. They are not ordinary default context for scoped review.
