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

The exact review forms are `spec_flow_review`, `spec_flow_review:full`, and `spec_flow_design_review`. These are referenced by the keyword table below.

0. If the entry expression does not match an exact review form, match against this keyword table in order:
   - If the expression contains "mechanism audit" or "framework correctness" → treat as `spec_flow_review`, default `scoped_review`.
   - If the expression contains "governance review" or "governance audit" → treat as `spec_flow_review`, default `scoped_review`.
   - If the expression contains "design review" or "design audit" or "design quality" → treat as `spec_flow_design_review`.
   - If none match → stop per Section "Unrecognized Entry".

1. `spec_flow_review` checks mechanism correctness.
2. `spec_flow_design_review` checks design quality and agent operability.

Plain exact `spec_flow_review` routes through this file first and remains scoped.
The only full-scope mechanism review entry is exact `spec_flow_review:full`.

Plain exact `spec_flow_design_review` routes through this file first, then directly delegates to `framework/spec_flow_design_review.md`.

When the entry is exact `spec_flow_review:full`, `framework/spec_flow_review.md` is the deep-audit owner.

If the entry expression does not match any recognized review entry (`spec_flow_review`, `spec_flow_review:full`, `spec_flow_design_review`), stop and report that the entry is unrecognized. Do not silently fall through to a default or guess the caller's intent.

### Unrecognized Entry

The paragraph above is the Unrecognized Entry stop condition.

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

### Layout-Aware Path Resolution

Files in `framework/` reference paths that resolve differently depending on layout.
`docs/specs/` paths are project-instance files present only in `installed_project` layout;
in `source_repo` layout they do not exist and must be treated as informational references
(agents must check path existence before reading and skip non-existent paths with a documented note).
Lifecycle and rule files at `framework/lifecycle/` and `framework/governance/rules/` may include
layout-aware notes on specific Required Reads entries; this section is the centralized authority
for how those path references should be resolved.

Project entry files (`AGENTS.md`, `CLAUDE.md`, `GEMINI.md`) contain managed blocks
(defined in `framework/operations/entry_routing.md` Entry File Registration section)
that must not contradict framework governance rules on routing, review scope, or lifecycle
progression. Use `specflowctl entry check` to verify managed-block consistency across
all registered entry files. When a managed block contradicts framework governance rules,
the framework rule takes precedence and the executor must report the contradiction as a
governance concern.

If `--layout auto` detects both `installed_project` and `source_repo` markers, the review must stop and require an explicit `--layout installed` or `--layout source` argument. Auto-detection must not silently choose one layout.

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
