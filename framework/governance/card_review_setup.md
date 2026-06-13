# Card Review Setup

## Purpose

This file defines the procedure for setting up a representative test project and verifying
context card and evaluation request output. The project data lives under `_governance_review/project/`
(committed). Generated artifacts (`cards/`, `requests/`, `specflowctl`) are gitignored.

This procedure is invoked by `spec_flow_review:full` Section 2.8.1 and 2.8.2 only.
Scoped review does not perform full card generation and verification.

## Setup

1. **Rebuild the binary from current source (MANDATORY every review).**
   ```text
   cd tooling && go build -o ../_governance_review/specflowctl ./cmd/specflowctl/
   ```

2. **Copy framework files into the test project.** Project data is committed; framework files
   must be copied from the current source at review time:
   ```text
   cp framework/lifecycle/unit_check.md              _governance_review/project/framework/lifecycle/
   cp framework/lifecycle/unit_verify.md             _governance_review/project/framework/lifecycle/
   cp framework/lifecycle/unit_stable_verify.md      _governance_review/project/framework/lifecycle/
   cp framework/lifecycle/unit_impl.md               _governance_review/project/framework/lifecycle/
   cp framework/lifecycle/unit_promote.md            _governance_review/project/framework/lifecycle/
   cp framework/lifecycle/unit_init_new_fork.md      _governance_review/project/framework/lifecycle/
   cp framework/lifecycle/overview.md                _governance_review/project/framework/lifecycle/
   cp framework/core/object_model.md                 _governance_review/project/framework/core/
   cp framework/core/independent_evaluation.md       _governance_review/project/framework/core/
   cp framework/process_snapshot_contract.md         _governance_review/project/framework/
   cp framework/spec_writing_guide.md                _governance_review/project/framework/
   cp framework/candidate_intent.md                  _governance_review/project/framework/
   cp framework/operations/entry_routing.md          _governance_review/project/framework/operations/
   cp framework/governance/rule_system.md            _governance_review/project/framework/governance/
   cp framework/governance/rules/rule_new.md         _governance_review/project/framework/governance/rules/
   cp framework/governance/rules/rule_sync.md        _governance_review/project/framework/governance/rules/
   ```

3. **Clean previous artifacts.** Remove all previously generated cards and requests:
   ```text
   rm -rf _governance_review/cards/* _governance_review/requests/*
   ```

## Generation

4. Generate unit cards:
   ```text
   _governance_review/specflowctl context card --object-type unit --object <name> --repo-root _governance_review/project
   ```

5. Generate rule cards:
   ```text
   _governance_review/specflowctl context card --object-type rule --object <name> --repo-root _governance_review/project
   ```

6. Generate evaluation requests:
   ```text
   _governance_review/specflowctl evaluation request --object-type unit --object <name> --pack <pack> --repo-root _governance_review/project
   ```

## Verification

7. Verify each card against `framework/spec_flow_review.md` Section 2.8.1 (10 properties).

8. Verify each request against `framework/spec_flow_review.md` Section 2.8.2 (6 properties).

9. Report failures. A single failure blocks the mechanism review from passing.
