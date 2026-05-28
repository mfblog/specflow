## Host Instructions

Content outside the managed block below belongs to the host repository.

Keep repository-specific rules outside the managed block. `specFlow` tooling may update only the managed block.



<!-- SPECFLOW:BEGIN -->
## specFlow Addendum

### 1. What specFlow Governs

specFlow is the repository workflow for governed engineering changes. It keeps written project truth, candidate and stable state, implementation permission, verification evidence, and promotion aligned before code moves ahead of the project record.
It does not mean every code edit must change spec documents.

A unit is one governed engineering responsibility. It may describe a package, feature, service, or end-to-end result, but it is not defined by directory shape.
Stable truth is the accepted current description. Candidate truth is the proposed next description. A rule is reusable truth shared across units.
A Context Card is the command-specific action card that says what to read, what may be written, what is forbidden, and how to close.

### 2. When This Applies

Treat a request as specFlow work when it names a specFlow command, or when the requested work may affect unit behavior, boundary, protocol, acceptance, rule, lifecycle state, or path ownership.
A normal implementation request becomes specFlow work as soon as testing, debugging, review, or exploration shows one of those surfaces may change.
A code-only implementation edit may stay implementation_only when current stable or candidate truth already allows the requested result.
If the path may belong to a formal unit, do not decide from directory shape alone.

### 3. Before Repo Edits

Before editing repo-tracked code, tests, configs, prompts, fixtures, integration scripts, or other implementation-side files, first decide whether specFlow owns the path or decision.

1. If the request exactly matches `unit_advance:{unit}`, read `specflow/framework/advance_policy.md`.
2. If the request exactly matches a standard lifecycle command (`unit_init:{unit}`, `unit_new:{unit}`, `unit_fork:{unit}`, `unit_check:{unit}`, `unit_plan:{unit}`, `unit_impl:{unit}`, `unit_verify:{unit}`, `unit_promote:{unit}`, `unit_stable_verify:{unit}`), read `specflow/framework/lifecycle/overview.md`, then the matching Context Card under `specflow/framework/lifecycle/`. Do not route an exact lifecycle command through direct implementation-change classification.
3. For every other specFlow request, read `specflow/framework/operations/entry_routing.md` first.
4. Do not edit truth, process evidence, lifecycle status, rules, repository mapping, or implementation-side files until the active Context Card or operation explicitly allows that write.
5. Do not close an advancing gate from self-assessment. When the active Context Card requires independent evaluation, the process evidence must contain a valid independent reviewer receipt.
6. Do not guess project terms, object ownership, or lifecycle state from directory shape or chat. Read the durable source named by the active Context Card or operation.

Framework-root relative paths in routed files use `framework/...` as the logical framework root. In installed projects, resolve them under `specflow/framework/...`; project refs such as `docs/specs/...` remain repository-root relative.

### 4. Implementation-Change Classification

Use this gate only when no exact lifecycle Context Card is already active.
If the path may belong to a formal unit and the request may edit repo-tracked code, tests, configs, prompts, fixtures, integration scripts, or other implementation-side files, read `docs/specs/_status.md` before editing.
Read `docs/specs/repository_mapping.md` when path ownership, object ownership, implementation path registration, or support-surface ownership is unclear.
Then read `specflow/framework/operations/implementation_change.md`.

That operation owns the detailed classification. Classify the request as implementation_only, truth_writeback_required, or boundary_unclear before editing.
Implementation-only means the current stable or candidate truth already allows the requested edit without changing behavior, protocol, boundary, acceptance, rule, lifecycle state, or ownership.
Truth writeback required means the project record must change before implementation continues.
Boundary unclear means current truth is not enough to prove one safe implementation result; treat it like truth writeback required.

### 5. Drift During Work

If testing, debugging, review, or exploration discovers behavior, protocol, boundary, acceptance, rule, or ownership change, stop ordinary implementation and reroute through `specflow/framework/operations/entry_routing.md`.
For an existing stable unit that needs a behavior or boundary change, the next truth step is `unit_fork:{unit}` only when `docs/specs/_status.md` records that command as legal.
If unsure whether the work is still implementation-only, stop and classify again before any further implementation-side edit.

### 6. Examples

1. A spelling or comment-only repair that does not change meaning may be `implementation_only` after classification.
2. Fixing an implementation deviation where current truth already defines the intended behavior may be `implementation_only`.
3. Changing an external protocol, acceptance rule, or unit boundary is `truth_writeback_required`.

After routing, read only the active Context Card's required context. Enter on-demand expansions only when their trigger appears.
For any specFlow route, report the user-facing answer first and keep traceability details separate.
<!-- SPECFLOW:END -->
