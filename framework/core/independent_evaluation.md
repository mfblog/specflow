# Independent Evaluation

Independent evaluation prevents an executor from approving its own advancing gate.

specFlow defines the receipt contract, gate requirement, and handoff request file.
The agent harness is responsible for creating a separate reviewer conversation with the minimal review pack named by the active Context Card.
specFlow creates independent evaluation request files.
specFlow does not create harness commands, reviewer sessions, tokens, or task scheduling.

## Handoff Requests

Before an executor asks for independent evaluation of an advancing gate, the executor must generate a request file:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> evaluation request --repo-root <repo-root> --object-type unit --object <unit> --pack <reviewer_pack> [--process check|plan|verify|stable_verify]
```

The request file is written under:

```text
docs/specs/_independent_evaluation/requests/unit/{unit}/{reviewer_pack}.md
```

The request file is a handoff instruction. It is not lifecycle evidence and is not consumed by `command close`.
The request file distinguishes review standard refs, review file refs, and review evidence refs.
Review standard refs are the authoritative criteria for the reviewer decision.

After the request file exists:

1. if the current agent runtime explicitly exposes an independent executor capability, the executor may send only the request file to that independent executor.
2. if no such capability is explicitly available, the executor must stop and give the user the request file path plus the trigger instruction from the command output.
3. the reviewer reads the request file and returns `pass`, `blocked`, or `needs_human_decision`.
4. the reviewer must not modify repository files.
5. after `pass`, the executor writes the receipt into the process evidence and records the reviewer pack, request file, and supplied durable refs in `review_input_refs`; freshness reuse records the same ref shape in `freshness_review_input_refs`.

The tooling validates that the candidate process artifact is mechanically ready for review without requiring the not-yet-written independent evaluation receipt.
For freshness reuse, request generation is allowed only when deterministic validation reports `text_drift` with `evidence_reuse: pending_review`.
This tooling check does not prove reviewer isolation and does not judge whether the reviewer made a good semantic decision.

## Roles

Executor may write specs, plans, implementation, and process evidence when the active Context Card allows those writes.

Reviewer must not modify repository files. The reviewer only reads the review pack and returns one result:

```text
pass | blocked | needs_human_decision
```

## Minimal Context

`reviewer_context: minimal_context` means the reviewer receives only the fixed reviewer pack for the current gate.

The reviewer must not inherit the executor's full working context as authority, including:

1. executor chain-of-thought, draft rationale, or private notes.
2. broad repository scans that are not named by the pack.
3. unverified assumptions from the implementation session.
4. unrelated policy, governance, recovery, or migration context.
5. prior chat conclusions unless they are recorded as durable human-confirmation refs.

The reviewer reads only:

1. the user goal or command target needed for the gate.
2. the current artifact under review.
3. the minimal durable truth required by the Context Card.
4. the current evaluation criteria.

## Reviewer Packs

The review criteria are embedded in the Evaluation Questions section of each pack. Review Standard Refs listed below are informational — they show the original source of the criteria but are not required reading. The reviewer should answer the Evaluation Questions directly.

### `unit_check_pass`

Review Standard Refs:

1. `framework/core/independent_evaluation.md` - reviewer isolation, legal reviewer outputs, receipt rules, and anti-patterns.
2. `framework/lifecycle/unit_check.md` - whether candidate truth is clear enough for downstream work.

Allowed Inputs:

1. user goal or exact `unit_check:{unit}` target.
2. candidate unit truth, candidate appendices owned by the unit, stable truth, and rules.
3. `_check_result/unit/{unit}.md`.
4. `framework/lifecycle/unit_check.md` check questions.

Forbidden Inputs:

1. implementation files unless repository mapping is part of the boundary question.
2. executor rationale not present in durable truth or `_check_result`.

Evaluation Questions:

1. Is the unit goal, responsibility, boundary, dependency truth, and rule binding explicit enough for downstream work?
2. Is the full unit package, including main Spec, owned appendices, unit dependencies, and applicable rules, clear and consistent enough for downstream work?
3. Are acceptance items testable without inventing behavior?
4. Does `_check_result` match the candidate truth and evidence refs?

Legal Output:

```text
pass | blocked | needs_human_decision
```

### `unit_verify_ready_to_promote`

Review Standard Refs:

1. `framework/core/independent_evaluation.md` - reviewer isolation, legal reviewer outputs, receipt rules, and anti-patterns.
2. `framework/lifecycle/unit_verify.md` - whether verification evidence is sufficient for promotion readiness.
3. `framework/spec_writing_guide.md` - acceptance item format standard, including `verification_type`, `evidence_requirements`, and `affects` fields.

Allowed Inputs:

1. user goal or exact `unit_verify:{unit}` target.
2. candidate unit truth, valid check result, and active plan.
3. verify result under review.
4. evidence refs needed to inspect acceptance coverage, retirement evidence, and package-aware delta verification.

Forbidden Inputs:

1. unrecorded executor claims that tests passed.
2. implementation changes not represented by plan or evidence refs.
3. promotion judgment not grounded in verify evidence.

Evaluation Questions:

**Functional Correctness:**
1. Does the verify result cover every executable acceptance item?
2. Does each executable acceptance item have inspectable evidence refs that prove the candidate behavior through the declared verification surface?
3. Does the verify result reject weak evidence as sufficient by itself, including generic test success, absent old strings, present new files, or present new fields?

**Scope Verification:**
4. For acceptance items that declare `affects`, does the `scope_verification` record confirm that all affected files, appendices, rules, and dependencies were verified?

**Code Quality:**
5. Does the implementation contain dead code, unnecessary abstractions, or duplicated logic that could be simplified?
6. Is the implementation concise and proportional to the acceptance item scope? (For replacement scenes: is the new code volume proportionate to the replaced code volume?)
7. Does the implementation introduce over-engineering (layers, interfaces, or patterns not justified by current requirements)?

**Retirement Completeness (replacement scenes only):**
8. Are the old code paths declared in `affects.files` fully removed (not left as dead wrappers, compatibility stubs, or commented-out code)?
9. Is there any remaining module, test, or configuration that references the deleted paths?

The reviewer records findings for each dimension. An outcome of `pass` requires all functional and scope questions to pass. Code quality and retirement questions may produce `quality_concern` findings that are recorded in the review findings but do not automatically block promotion; the executor may address them in the current round or defer to a follow-up round.
4. For primary protocol, default page, primary presentation, API, or artifact-generation changes, does the evidence inspect real generated artifacts, API return values, DOM/screenshots, rendered text, CLI output, or tests proving the mainline path uses the candidate protocol?
5. Does the verify result prove every retirement target with `pass` and `mainline_dependency: not_required` evidence?
6. Does `package_delta_verification` prove every `planned_change_scope` entry without violating appendix, rule, unit dependency, or acceptance truth?
7. Is the candidate ready for promotion without hiding unresolved gaps?

Legal Output:

```text
pass | blocked | needs_human_decision
```

### `unit_stable_verify_advancing`

Review Standard Refs:

1. `framework/core/independent_evaluation.md` - reviewer isolation, legal reviewer outputs, receipt rules, and anti-patterns.
2. `framework/lifecycle/unit_stable_verify.md` - whether stable alignment or the controlled next step is supported.

Allowed Inputs:

1. exact `unit_stable_verify:{unit}` target.
2. stable unit truth, stable appendices owned by the unit, rules, and repository mapping snapshot.
3. stable verify result under review.
4. implementation surface refs and evidence refs needed to inspect stable alignment.
5. decision criteria from `framework/lifecycle/unit_stable_verify.md`.

Forbidden Inputs:

1. candidate truth unless the stable verify result explicitly cites it as historical context.
2. proposed repairs or changes not captured in the stable verify result.
3. executor preference for aligned, controlled repair, or controlled change outcomes.

Evaluation Questions:

1. Does current implementation align with stable truth, or does the stored decision correctly identify the controlled next step?
2. Does the evidence matrix cover every current stable acceptance item?
3. Are implementation surface refs and evidence refs sufficient for the stored decision?

Legal Output:

```text
pass | blocked | needs_human_decision
```

### `freshness_text_drift_reuse`

Review Standard Refs:

1. `framework/core/independent_evaluation.md` - reviewer isolation, legal reviewer outputs, freshness receipt rules, and anti-patterns.
2. `framework/core/freshness.md` - whether text drift may safely reuse existing process evidence.

Allowed Inputs:

1. current truth or spec file.
2. prior process evidence being reused.
3. deterministic freshness classification showing `text_drift`.
4. acceptance behavior fingerprint comparison and current fingerprint reported by tooling.

Forbidden Inputs:

1. reuse claims when deterministic validation reports `semantic_drift`, `acceptance_drift`, `dependency_drift`, `schema_drift`, or `unknown_drift`.
2. executor assertions that the text change is harmless without current file refs.
3. unrelated changes outside the file and process evidence under review.

Evaluation Questions:

1. Is the change only wording, formatting, or clarification that preserves the acceptance behavior already reviewed?
2. Does the prior evidence still answer the same gate question?
3. Is recreating evidence unnecessary for semantic safety?

Legal Output:

```text
pass | blocked | needs_human_decision
```

## Receipt Fields

Advancing process evidence that requires independent evaluation must include:

```yaml
evaluation_mode: independent
reviewer_result: pass
reviewer_context: minimal_context
review_input_refs: {reviewer_pack};{request_file};{durable_input_refs}
review_findings: none
human_decision_refs: none | list
```

`review_input_refs` must record the reviewer pack name, generated request file path, and durable refs that were supplied to the reviewer.

## Gate Rules

An advancing gate may close only when:

1. `evaluation_mode` is `independent`.
2. `reviewer_result` is `pass`.
3. `reviewer_context` is `minimal_context`.
4. `review_findings` is `none`.
5. `review_input_refs` contains the reviewer pack, request file path, and at least one durable input ref.
6. `human_decision_refs` is `none` or points to durable human-confirmation refs.

`blocked` and `needs_human_decision` are valid reviewer outcomes, but they are not advancing outcomes.

## Anti-Patterns

These do not satisfy independent evaluation:

1. executor writes its own pass receipt.
2. reviewer inherits the executor's full conversation or private notes as authority.
3. reviewer evaluates executor rationale instead of durable artifacts.
4. reviewer edits repository files.
5. process evidence records `reviewer_result: pass` while `review_findings` is not `none`.
6. `needs_human_decision` is resolved only by chat text with no durable human-confirmation ref.
7. tooling field validation is treated as proof that reviewer isolation actually happened.
8. tooling field validation is treated as proof that the semantic decision was good.

## Required Gates

The following advancing evidence requires the receipt:

1. `docs/specs/_check_result/unit/{unit}.md`
2. `docs/specs/_verify_result/unit/{unit}.md`
3. `docs/specs/_stable_verify_result/unit/{unit}.md`

`unit_init`, `unit_new`, and `unit_fork` create truth entry points and do not require this receipt by default.

`unit_promote` relies on already verified evidence and does not add a second receipt.
