# Shared Extraction Boundary Review

## 1. Purpose

This flow decides whether content currently written inside a module main file or module appendix has reached the threshold where it should be extracted into a Shared Appendix.

It answers four questions:

1. whether the current content is already a shared candidate
2. which modules or shared topics it involves
3. whether it should be extracted into `c_shared_xxx` now
4. if not, why it should stay inside the module for now

This flow is not a normal module command, is not in `{command}:{module}` form, and does not enter `docs/specs/_status.md`.

---

## 2. Scope

By default this flow handles:

1. boundary decisions between `shared` and module main files or module appendices
2. user-requested shared-extraction review
3. validation of shared-candidate signals found during module-command execution
4. deciding whether multiple modules now require a truth that should exist in only one formal definition

By default it does not:

1. create a new `c_shared_xxx.md` or `s_shared_xxx.md` directly
2. directly modify module truth to complete extraction
3. replace `cand_check`, `cand_promote`, or `shared_flow_reconcile`

---

## 3. Trigger Modes

### 3.1 Active Trigger

Enter this flow directly when the user explicitly asks things such as:

1. "Should this be extracted into shared?"
2. "Review whether this should become a common part."
3. "Is this already shared governance content?"

In that case, expanding to cross-module reading is allowed if the question requires it.

### 3.2 Passive Trigger

If the executor notices a shared-candidate signal during a module command, the executor may suggest entering this flow.

Fixed rules:

1. Passive trigger only allows "suggest and ask for confirmation" by default. It must not automatically expand into a full cross-module shared review.
2. Passive trigger does not block the current module command by default.
3. Only if the current command's required reading range already contains clear dual-source-of-truth evidence may the issue be raised directly as a blocking finding.

---

## 4. Preconditions

Before execution:

1. read `specflow/framework/docs/agent_guidelines/spec_policy.md` and `specflow/framework/docs/agent_guidelines/command_policy.md`
2. identify the current formal location of the content under review:
   - module main file
   - module appendix
   - or near an existing shared file
3. read the current-layer main file of the directly relevant module, plus any appendix explicitly referenced by that layer
4. if existing Shared Appendix files are involved, read them
5. if the flow is actively triggered and naturally involves multiple modules, read other modules only as needed; if it is passively triggered and the user has not approved scope expansion, do not expand beyond the current command's required reading range

Additional note:

1. The goal is to decide whether the boundary has already formed, not to assume shared extraction should happen.

---

## 5. Procedure

1. Locate the formal landing point of the candidate content and confirm whether it currently lives in a module main file, module appendix, or near an existing shared file.
2. Judge whether the content hits any shared-candidate signal, including at least:
   - the current text explicitly says "general", "unified", "used by multiple modules", or similar
   - it defines a highly shareable object such as a shared output protocol, fallback, object expansion, few-shot, or failure semantics
   - it overlaps clearly with existing shared content or another formal truth in name, responsibility, or protocol meaning
3. If only candidate signals exist and a second formal module has not yet been confirmed to depend on the same truth, the default conclusion can only be:
   - do not extract yet
   - or ask whether to expand into cross-module review
4. If the user actively triggered the review or approved scope expansion, check whether a second formal module already needs the same formal behavior truth.
5. Distinguish between:
   - the same formal behavior truth
   - versus only similar theme, implementation, or structure
6. The fixed criterion is:
   - if this content should have only one formal definition in the repository, it is a shared candidate
   - if modules can still evolve it independently, it is not shared
7. If it is a shared candidate, decide whether extraction should happen immediately:
   - if keeping it in multiple modules would create a dual-source-of-truth risk, extract now
   - if a second-module need has only just appeared and the shared boundary or naming is not yet closed, keep it as a shared candidate but do not extract yet
8. Output one of four conclusion types:
   - keep it in the current module; not a shared candidate
   - it is a shared candidate, but only record or suggest it for now
   - it is a shared candidate and should be extracted this round into `c_shared_xxx`
   - a dual source of truth already exists and shared closure must happen first

---

## 6. Decision Criteria

### 6.1 Formal Definition Of `shared`

`shared` does not mean "something that might be reused later." It means:

1. a formal behavior truth depended on by multiple formal modules
2. and a truth that should have only one formal definition in the repository

### 6.2 Boundary Against Module Appendix

1. The first appearance of content should stay in the current module main file or module appendix by default.
2. Do not extract to `shared` early just because it might be reused later.
3. Only when a second formal module needs the same formal behavior truth does the content enter shared-candidate review.
4. `shared` solves the "single formal definition" problem, not "similar topic" or "similar implementation."

### 6.3 When Immediate Extraction Is Recommended

Extraction into `c_shared_xxx` should normally be recommended only when all of the following hold:

1. at least two formal modules already depend on the same formal behavior truth
2. keeping it inside multiple modules would create a dual-source-of-truth risk
3. the shared topic is stable enough to be named and hosted independently

### 6.4 When Blocking Is Required

Blocking is required when the current command's required reading range already confirms any of the following:

1. the current module is redefining a truth already formalized in an existing shared file
2. the current module and another already-read formal module have already formed a dual source of truth for the same formal behavior
3. keeping the in-module definition would directly create a formal semantic conflict

### 6.5 File Model

1. One shared object maps to one shared file.
2. A repository may contain multiple shared files at the same time.
3. Do not keep unrelated shared topics permanently stuffed into a single catch-all shared file.

---

## 7. Output Contract

The output must include at least:

1. the current formal location of the reviewed content
2. whether it hits shared-candidate signals
3. whether a second formal module needing the same truth has already been confirmed
4. the conclusion type:
   - not a shared candidate
   - shared candidate, but suggestion only for now
   - recommend extracting into `c_shared_xxx`
   - shared closure is required first
5. if the conclusion is not immediate extraction, explain clearly why the content should remain in the current module
6. if extraction or mandatory closure is recommended, identify the involved modules or shared topic

---

## 8. Non-Goals

This flow does not:

1. automatically create shared files
2. automatically migrate module content
3. automatically modify `_status.md`
4. automatically trigger a large cross-module scan without user approval or command-scope authorization

---

## 9. Examples

### 9.1 First Appearance, Not Shared Yet

`module_a` defines a fallback protocol for the first time.

Conclusion: keep it in the current-layer body or appendix of `module_a`; do not extract it to shared just because it might be reused later.

### 9.2 A Second Module Needs The Same Truth

`module_b` now also needs the same formal fallback protocol as `module_a`.

Conclusion: enter `shared_extract_review` and decide whether it should become `c_shared_fallback_xxx.md`.

### 9.3 Passive Trigger Only Suggests

During `cand_check:module_b`, the current body explicitly says this protocol is reused by multiple modules, but the user has not authorized cross-module review.

Conclusion: suggest that a shared candidate may exist and ask whether to continue with shared extraction boundary review.

### 9.4 Confirmed Dual Source Of Truth Blocks

During `cand_check:module_b`, the command's required reading range already includes an existing shared formal truth, while `module_b` redefines the same protocol.

Conclusion: report a blocking issue in the current command and require shared closure first.
