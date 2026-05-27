# Source Repository Agent Instructions Example

This file is a shareable example for local source-repository agent instructions.
Personal preferences should stay in local `AGENTS.md`, `CLAUDE.md`, or `GEMINI.md` files, which may be ignored by Git.

## Governance Review Shortcut

This is the specFlow source repository, so use local `framework/...` paths.

For `spec_flow_review`, `spec_flow_design_review`, or any governance/design review request:

1. Read `framework/governance/review.md`.
2. Read `framework/governance/review_scope.md`.
3. Default to `scoped_review`.

Use `framework/spec_flow_review.md`, `framework/spec_flow_design_review.md`, `_governance_review/` run-state files, baseline slice tables, dynamic slice tables, or score-state tables only when the user explicitly asks for `full-scope`, `baseline`, `deep audit`, release-level governance audit, `resumable review`, run-state-backed review, or exact `spec_flow_review:full`.
