# Lifecycle Authority

Lifecycle authority is evidence-based.

Standard unit lifecycle progression is valid only with current valid evidence, a required independent evaluation receipt when applicable, deterministic validation, and successful `command close`.
Independent evaluation requests are handoff instructions that can precede the receipt; they do not replace the receipt.

This contract is the authority model for standard unit lifecycle progression.

## Authority Sources

`command close` is the only authority that may advance `_status.md` for standard unit lifecycle work.

Advancing evidence is authoritative only when all of these are true:

1. the active lifecycle Context Card allows the evidence write.
2. the current process file passes the matching `snapshot validate-process` check.
3. the process file contains a valid independent reviewer receipt when the Context Card requires one.
4. `command close` accepts the requested outcome and evidence.

Input evidence is consumable only when the current file still passes deterministic validation.

## Iteration Rule

`blocked`, `fix_required`, `decision_checkpoint`, and other non-advancing outcomes do not permanently disqualify later work.

After repair, clarification, or additional evidence, the executor may form current evidence again. That evidence can advance when it passes deterministic validation, carries the required independent reviewer receipt, and `command close` succeeds.

Chat-only statements, local rechecks, manual hashes, editor display, and temporary script output are diagnostic only. They do not become lifecycle authority unless written into the active Context Card's allowed evidence shape and validated by tooling.

## Command-Specific Notes

`unit_check`, `unit_plan`, `unit_verify`, and advancing `unit_stable_verify` outcomes require independent evaluation receipts.

`unit_impl ready_for_verify` does not require an independent reviewer receipt because it means implementation work is ready to be verified. It does not approve correctness.

`unit_promote` relies on already verified current evidence and `command close`; it does not create a second reviewer receipt.

Fallback and recovery outcomes may run without valid downstream evidence only when the active lifecycle Context Card or `framework/lifecycle/recovery.md` defines that outcome as the legal recovery path.
