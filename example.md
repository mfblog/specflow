# Source Repository Agent Instructions Example

This file is a shareable example for local source-repository agent instructions.
Personal preferences should stay in local `AGENTS.md`, `CLAUDE.md`, or `GEMINI.md` files, which may be ignored by Git.

## Governance Review Shortcut

This is the specFlow source repository, so use local `framework/...` paths.

For `spec_flow_review` or ordinary governance review requests:

1. Read `framework/governance/review.md`.
2. Read `framework/governance/review_scope.md`.
3. Default to `scoped_review`.

Use `framework/spec_flow_review.md`, `_governance_review/` run-state files, baseline slice tables, or dynamic slice tables only for exact `spec_flow_review:full`.

For `spec_flow_design_review`:

1. Read `framework/governance/review.md`.
2. Read `framework/spec_flow_design_review.md`.
3. Run the default full-scope design-baseline review. Do not narrow it to `scoped_review`.
