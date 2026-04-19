# Process Snapshot Contract

## 1. Purpose

This file defines the fixed snapshot contract used by Spec Flow process files.

It answers six questions:

1. what `spec_fingerprint` means
2. what `module_appendix_snapshot` means
3. what `system_constraints_stable_fingerprint` means
4. what `shared_contract_snapshot` means
5. how these values must be normalized before comparison
6. which files must use this contract

This is a centralized governance contract. Executors must not invent alternative snapshot shapes or hashing inputs per command.

---

## 2. Scope

This contract governs snapshot fields written into:

1. `docs/specs/_check_result/{module}.md`
2. `docs/specs/_plans/{module}.md`
3. `docs/specs/_verify_result/{module}.md`
4. any governance flow that re-validates those snapshot fields, including `shared_sync`

It does not define business-module truth.
It only defines how process files record the truth version they were written against.

---

## 3. File Fingerprint Contract

### 3.1 `spec_fingerprint`

`spec_fingerprint` is the fingerprint of the module main Spec file bound by the current process file.

Rules:

1. it always fingerprints the exact current-layer main Spec file recorded by `spec_file_ref`
2. it includes the whole file content:
   - frontmatter
   - headings
   - body text
   - fenced code blocks
3. it must not hash only the body while skipping frontmatter
4. it must not hash only selected sections

### 3.2 `module_appendix_snapshot`

`module_appendix_snapshot` records the exact module-local appendix set explicitly referenced by the current-layer main Spec file.

It has only two legal forms:

1. literal `none`
2. a normalized ordered list where each item contains:
   - `file_ref`
   - `appendix_ref`
   - `fingerprint`

Rules:

1. include only current-layer module-local supporting files explicitly referenced by the current-layer main Spec file
2. do not include the main Spec file itself
3. do not include Shared Contract files from `shared_contract_refs`
4. if no module-local appendix file is explicitly referenced by the current-layer main Spec file, use literal `none`
5. do not use an empty list, `null`, omitted field, or natural-language placeholder text

Item meanings:

1. `file_ref`
   - the exact repository path of the bound module-local appendix file
2. `appendix_ref`
   - `<appendix_file_prefix>@<frontmatter.spec_version_ref>` when that frontmatter exists
   - otherwise `<appendix_file_prefix>@unversioned`
3. `fingerprint`
   - the Section 3 hash of that exact appendix file

Ordering rules:

1. sort by `file_ref`
2. then by `appendix_ref`

Executors must compare the normalized ordered form.

### 3.3 `system_constraints_stable_fingerprint`

`system_constraints_stable_fingerprint` is the fingerprint of `docs/specs/system/stable/s_system_constraints.md` when that file exists and is formally bound by the current round.

Rules:

1. if `system_constraints_stable_file_ref` points to a real file, fingerprint that exact file with the same normalization rules as `spec_fingerprint`
2. if no formal global baseline exists and the module correctly binds `system_constraints_stable_ref=none`, then:
   - `system_constraints_stable_file_ref=none`
   - `system_constraints_stable_version_ref=none`
   - `system_constraints_stable_fingerprint=none`

### 3.4 Hash Algorithm

The hash algorithm is fixed:

1. normalize file text according to Section 4
2. encode the normalized text as UTF-8
3. compute `sha256`
4. render the result as lowercase hexadecimal

Executors must not substitute another algorithm.

---

## 4. Text Normalization Rules

Before hashing a markdown truth file, normalize it in this exact order:

1. read the full text of the file
2. convert all line endings to `LF`
3. if the file does not end with `LF`, append exactly one trailing `LF`
4. do not trim leading spaces
5. do not trim trailing spaces inside lines
6. do not remove blank lines
7. do not reorder frontmatter keys
8. do not apply markdown-aware formatting or semantic rewriting

Plain meaning:

1. this contract fingerprints the actual file text after only minimal line-ending normalization
2. it is intentionally stricter than "same meaning"
3. if the text changed, the fingerprint changed

---

## 5. Shared Contract Snapshot Contract

### 5.1 Shape

`shared_contract_snapshot` records the exact Shared Contract set bound by the current module round.

It has only two legal forms:

1. literal `none`
2. a normalized ordered list where each item contains:
   - `shared_contract_id`
   - `layer`
   - `file_ref`
   - `version_ref`
   - `fingerprint`

### 5.2 When To Use `none`

Use literal `none` only when the module current-layer truth binds no Shared Contract files.

Do not use:

1. an empty list
2. `null`
3. omitted field
4. natural-language placeholder text

### 5.3 Item Meanings

For each snapshot item:

1. `shared_contract_id`
   - the `shared_contract_id` from that file's frontmatter
2. `layer`
   - `candidate` or `stable`
3. `file_ref`
   - the exact repository path of the bound Shared Contract file
4. `version_ref`
   - `<shared_file_prefix>@<shared_version>`
   - for example `c_shared_xxx@0.1.0` or `s_shared_xxx@1.0.0`
5. `fingerprint`
   - the Section 3 hash of the exact bound Shared Contract file

### 5.4 Ordering Rules

When `shared_contract_snapshot` is a list, normalize ordering before write-back or comparison:

1. sort by `shared_contract_id`
2. then by `layer`
3. then by `file_ref`

Executors must compare the normalized ordered form.

### 5.5 Inclusion Boundary

`shared_contract_snapshot` records only the Shared Contract files formally bound by the module current-layer truth.

It must not duplicate:

1. the module's own `shared_contract_refs` prose
2. `bound_modules`
3. unrelated shared files not formally bound by the module

If a Shared Contract file's `bound_modules` field changes, the file fingerprint may also change.
That change does not by itself invalidate downstream process files, because `bound_modules` is declarative metadata rather than the module's formal binding source.
Treat a `bound_modules`-only delta as governance drift to be reported and repaired separately.
A re-validating command or governance flow must not infer a `bound_modules`-only delta from fingerprint change alone.

---

## 6. Re-Validation Rules

When a command or governance flow re-validates a process file, it must:

1. rebuild `spec_fingerprint` from the current bound main Spec file
2. rebuild `module_appendix_snapshot` from the current-layer main Spec file's explicitly referenced module-local appendix set, or `none`
3. rebuild `system_constraints_stable_fingerprint` from the current bound stable system-constraints file, or `none`
4. rebuild `shared_contract_snapshot` from the module's current-layer bound Shared Contract set using Section 5
5. compare the rebuilt values against the process file snapshot fields exactly

Shared Contract exception:

1. if a rebuilt Shared Contract snapshot differs only because a bound Shared Contract file is explicitly declared by the active command or governance flow as `bound_modules`-only for the current round, do not invalidate the process file on that basis alone
2. `shared_sync` consumes that declaration through its execution-local `bound_modules_only_shared_file_refs` input field
3. otherwise, do not infer a `bound_modules`-only delta from fingerprint difference alone
4. in that case, report governance drift instead and keep using the module's formal binding source from `shared_contract_refs` only when the active flow has the explicit declaration from Rule 1

Resolver rule:

1. before rebuilding `shared_contract_snapshot`, executors must first resolve the currently bound Shared Contract files from `shared_contract_refs` using the Shared Contract binding contract from `specflow/framework/docs/agent_guidelines/spec_policy.md` Section 6.1

If any one of those values differs, the process file is invalid for downstream consumption.

---

## 7. Relationship To Other Files

This contract works together with:

1. `specflow/framework/docs/agent_guidelines/spec_policy.md`
2. `specflow/framework/docs/agent_guidelines/candidate_handoff_contract.md`
3. process-file READMEs under `docs/specs/` and `specflow/templates/root/docs/specs/`

Priority rules:

1. `spec_policy.md` defines that process files must match current truth bindings
2. this file defines the exact snapshot fields and normalization rules used to judge that match
3. command files and process-file READMEs must stay consistent with this file

---

## 8. Non-Goals

This file does not:

1. define module behavior truth
2. replace command-specific fallback rules
3. define the markdown layout of process files beyond the snapshot fields above
