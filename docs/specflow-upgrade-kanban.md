# specFlow Upgrade Kanban

This board tracks the next design upgrades after the Context Card and independent evaluation round.

Use it as a lightweight planning artifact, not as lifecycle process evidence.

## Status Legend

- `[x]` done
- `[>]` active next
- `[ ]` planned
- `[?]` needs decision

## Done

- [x] Introduce progressive disclosure through lifecycle Context Cards.
- [x] Lighten `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` managed entry blocks.
- [x] Add independent evaluation receipt contract.
- [x] Require receipt fields for `check`, `plan`, `verify`, and `stable_verify` process validation.
- [x] Add tests for missing receipt, blocked reviewer, review findings, command close refusal, and Context Card shape.
- [x] Redefine lifecycle authority away from one new full-scope command run.
- [x] Add freshness impact levels.
- [x] Define incremental adoption modes.
- [x] Strengthen independent evaluation without adding harness commands.
- [x] Slim governance review default burden.

## Active Next

None.

## Planned

## Decision Backlog

- [?] Should governance review keep fixed full-scope slices, or should fixed slices move into an on-demand deep-audit mode?

## Evaluation Checklist

Before closing each upgrade item, answer:

- [ ] Does this reduce default agent context?
- [ ] Does this avoid relying on the executor to self-approve?
- [ ] Does this preserve deterministic tooling boundaries?
- [ ] Does this reduce repeated work in realistic iteration?
- [ ] Does this avoid creating a new harness inside specFlow?
- [ ] Does this keep the framework adoptable in a small project?

## Suggested Order

1. Lifecycle authority model.
2. Freshness impact levels.
3. Incremental adoption modes.
4. Independent evaluation pack examples.
5. Governance review slimming.
