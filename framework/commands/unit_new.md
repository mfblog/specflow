# Unit New Command

## 1. Purpose

This command creates the first `candidate` Spec for a brand-new unit.

Goals:

1. define the first complete candidate design
2. establish the starting point of the candidate chain
3. register the unit in `docs/specs/repository_mapping.md`
4. register the unit state in `docs/specs/_status.md`

## 2. Scope

By default this command handles:

1. first-time project initiation for a new unit
2. units that do not yet have any formally effective version
3. creation of the first `candidate`
5. registration of the new unit in `docs/specs/repository_mapping.md`

It does not:

1. invent a shared/unit boundary when the first candidate still depends on rule truth that is not yet formalized
2. write `rule_refs=none` as a placeholder when the new unit already depends on rule truth

### 2.1 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_new`-local entry, output, and stop rules.

## 3. Preconditions

1. complete the required pre-checks
2. the target unit name is explicit
3. the unit is not yet in `_status.md`
4. the goal is future design first, not capturing current truth first
5. read `specflow/framework/repository_mapping_policy.md`
6. read `docs/specs/repository_mapping.md`
7. confirm the target unit is not already present in `Object Registry` and does not conflict with any current `unit`, `scenario`, `rule`, support-surface, or ignore rule
8. read `specflow/framework/onboarding_decision_policy.md` and decide the first candidate's `source_basis` and `evidence_appendix_ref`
9. read `specflow/framework/candidate_intent_policy.md`; first candidates use `candidate_intent=change`
10. if the first candidate uses `source_basis=existing_implementation` or `source_basis=mixed`, prepare the required evidence appendix in the same round
11. if the first candidate depends on rule truth that is not yet formalized as `rule`, or if the shared/unit boundary is still unstable, do not start `unit_new`; resolve that rule truth through natural-language rule governance first
12. if the first candidate reuses already-existing rule truth, read the relevant `rule` files before writing `rule_refs`
13. if the round will create, update, or delete any unit `rule_refs` value or any file under `docs/specs/rules/**`, read `rule_sync.md`
14. if the round may remove intentional-unbound retention fields from a touched Rule file, read every current-layer unit or scenario main file needed to derive the real repository-wide binding set of each touched Rule from `rule_refs`

## 4. Procedure

1. if `s_g_rule_repository_baseline.md` exists, read it as the current formal global baseline; otherwise continue with the "no formal global baseline yet" state
2. decide whether the first candidate already reuses existing rule truth:
   - if no, the round may initialize `rule_refs=none`
   - if yes, the round must bind that rule truth explicitly in the first candidate instead of using `none`
3. define the new unit's goals, boundaries, protocols, and main flow
4. prepare the `docs/specs/repository_mapping.md` writeback for the new unit before candidate or `_status.md` mutation:
   - add or update one `Object Registry` row for the target unit
   - set `kind=unit`, `id={unit}`, `scope=capability`, and the one-line responsibility
   - set `spec_files=docs/specs/units/candidate/c_unit_{unit}.md` after the candidate file is created in this same round
   - set `registration_state=landed` only when concrete implementation paths are declared
   - if no implementation path is declared yet, set `registration_state=planned` and `implementation_paths=none`
   - record any implementation surface, support surface, governed root, ignore rule, or conflict rule that this first unit round already needs
   - if current repository truth is insufficient to write the exact mapping update without guessing, stop before candidate and `_status.md` writeback
5. create `docs/specs/units/candidate/c_unit_{unit}.md`
6. initialize `frontmatter.version` to `0.1.0`
7. initialize `frontmatter.candidate_intent=change`
8. initialize `frontmatter.source_basis` and `frontmatter.evidence_appendix_ref` according to `onboarding_decision_policy.md`
9. if `source_basis=existing_implementation` or `source_basis=mixed`, create the evidence appendix named by `evidence_appendix_ref`; if `source_basis=new_design` or `source_basis=replacement`, write `evidence_appendix_ref=none`
10. ensure the file covers the core sections of a formal Spec, including `Testability / Acceptance Criteria` with explicit acceptance items that satisfy `spec_writing_guide.md` Section 5
11. initialize `Rule Alignment`:
   - write `rule_refs=none` only when the first candidate does not yet reuse rule truth
   - if the first candidate already reuses existing rule truth, write the explicit `rule_refs` set using the Rule binding contract from `specflow/framework/spec_policy.md` Section 6.1 and explain that reuse in the candidate body in the same round
   - `rule_reuse_summary`
   - `rule_exceptions`
12. write the prepared `docs/specs/repository_mapping.md` update in the same round as the candidate writeback
13. if the round changed Rule bindings or touched Rule files:
   - derive the real repository-wide binding set of each touched Rule from current-layer unit and scenario `rule_refs` plus this round's prepared target-unit candidate writeback
   - if current repository truth is insufficient to derive that touched real binding set safely, stop and reroute through natural-language rule governance from current repository truth instead of guessing
   - do not write consumer metadata into touched Rule files; every touched Rule file must omit `bound_objects` after this writeback
   - if a touched Rule file now has one or more formal bound units after this round, remove or stop carrying any `unbound_retention`, `unbound_retention_reason`, and `unbound_retention_owner` fields from that resulting bound file state in the same round
14. update `_status.md`:
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=unit_check`
   - the deterministic command closure may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_new --object-type unit --object {unit} --outcome candidate_created --notes <status-note> --apply`
15. if the round changed any unit `rule_refs` value or any file under `docs/specs/rules/**`, run `rule_sync` after `_status.md` has been updated, even when no additional affected object is known yet
   - the deterministic reconciliation part may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> rule sync-impact --rule-refs <rule-ref> --units {unit}` or the corresponding `--rule-ids` form, and at least one rule trigger input must already be known before this deterministic execution starts

## 5. Stop Conditions

1. the first `candidate` exists
2. `docs/specs/repository_mapping.md` includes the new unit in `Object Registry` with its implementation registration state, the created candidate Spec file, and any implementation paths required by this first unit round
3. `_status.md` registration is complete
4. any first-round rule binding required by the candidate has been written explicitly instead of being left as placeholder `none`
5. Rule side effects, if any, are closed
6. the command does not automatically continue into implementation
7. if repository truth was insufficient to write the required repository mapping update safely, the command stopped before candidate and `_status.md` writeback instead of guessing
8. if repository truth was insufficient to close rule-truth binding metadata safely, the command stopped and rerouted through natural-language rule governance instead of guessing

## 6. Output Contract

1. initiation judgment
2. created file path
3. initialized candidate version
4. initialized `candidate_intent=change`
5. initialized `source_basis`
6. initialized `evidence_appendix_ref` and evidence appendix write result when required
7. initialized formal global baseline reference or `none`
8. initialized acceptance-item structure result
9. initialized explicit Rule binding set or confirmed `rule_refs=none`
10. whether the command had to stop and reroute through natural-language rule governance because repository truth was insufficient to close rule-truth binding metadata safely
11. `docs/specs/repository_mapping.md` writeback result, including the new `Object Registry` row and any path-ownership entries written in this round
12. `_status.md` update result
13. Rule reconciliation result when the round changed rule truth or bindings
14. remaining closure items
15. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - `current state` must explicitly confirm `Active Layer=candidate` and `Next Command=unit_check`
   - `next-stage entry gap` must explicitly confirm that entry into the later different command `unit_check` is already satisfied after `unit_new` closes

## 7. Non-Goals

1. creating the first formal `stable`
2. capturing historical behavior
3. automatically entering `unit_impl`
4. creating an independent stable `g_` rule candidate file
5. using `rule_refs=none` to postpone required rule-truth closure

## 8. Example

```md
unit_new:executor
```
