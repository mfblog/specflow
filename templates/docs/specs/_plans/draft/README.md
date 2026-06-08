# Candidate Plan Drafts (Agent-Internal)

This directory stores agent-internal planning working artifacts. These files are **not SpecFlow process evidence** and are not consumed by SpecFlow lifecycle gates.

## Status

`unit_plan` is no longer a SpecFlow-governed command. Draft plans are optional agent workspace artifacts.

Draft plans are not valid inputs to `unit_verify`, `unit_promote`, or any other SpecFlow lifecycle command.

## Guidance for Agents

A draft plan may record:
- `object_ref`, `truth_file_ref`, `truth_version_ref`, `truth_fingerprint`
- `changed_execution_surfaces`
- `current_known_paths` and `target_paths`
- `retirement_candidates` and `retirement_goals`
- `known_findings` and `open_modeling_unknowns`
- `slice_cutover_plan`
- `research_notes`

Draft plans do not carry gate-bearing semantics. If planning discovers missing behavior truth or acceptance truth, the agent should repair candidate truth before proceeding to verification.
