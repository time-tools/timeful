---
id: TASK-0356
title: >-
  Repair the gala-loop and gala-from-go instruction gaps found landing
  postgres/pool
status: Done
assignee: []
created_date: '2026-10-10 16:36'
updated_date: '2026-10-10 16:50'
labels: []
dependencies: []
references:
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/GALA.md
  - BACKLOG_WORKFLOW.md
documentation:
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/GALA.md
modified_files:
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/GALA.md
priority: medium
type: task
ordinal: 360300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The GALA loop's instructions accumulated gaps that cost rework in iteration 14 (postgres/pool) and recur in every iteration. A retrospective identified misleading and missing guidance: the clean-tree precondition is unsatisfiable because `backlog/backlog.md` is a permanent unrelated edit; the loop never states that a landed twin is committed; a cursor entry blocked by a recorded finding with a committed probe is re-probed on every iteration; `verify.sh` can fail on a stale `server/.gala/` cache whose remedy lives only in the ledger; `resource.Bracket`'s panic-value conversion and the function-type type-argument refusal were absent from `gala-from-go`; cursor entries with multiple candidates lack order/dependency notes; and probe conventions (one probe per directory, scratch output) are unstated.

Repair the instructions so the next loop iteration does not repeat that work. Keep each rule in its owning file: procedure in `.agents/skills/gala-loop/SKILL.md`, construct rules in `.agents/skills/gala-from-go`, and ledger and cursor conventions in `server/GALA.md`. The function-type type-argument caveat was already added to `gala-from-go` during iteration 14; this task covers the remaining items.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The gala-loop precondition requires twin-related paths to be committed and attributable, names known unrelated edits (at least `backlog/backlog.md`) as exempt, and no longer requires a globally clean `git status`.
- [x] #2 The loop's "One iteration" section has an explicit finalize step: commit the landed twin, the ledger update, and the task notes, leaving unrelated edits alone.
- [x] #3 The loop states that a cursor entry blocked by a recorded finding with a committed probe on the current compiler rev needs no per-iteration re-probe until the flake rev or the upstream state changes.
- [x] #4 The loop's verify step names the stale `server/.gala/` cache failure mode and its remedy, and says consecutive verify runs that disagree are a cache signal to clear rather than drift.
- [x] #5 Cursor entries in `server/GALA.md` with more than one candidate state the candidate order and any dependency between them, and blocked entries record the compiler rev they were last re-verified on.
- [x] #6 `gala-from-go` records that `resource.Bracket` releases on every exit path but re-raises a caught panic as a `*panicError`, so a string panic's value type differs from Go's `defer`.
- [x] #7 The loop's no-workaround-blocker step requires a newly derived construct substitution to be added to the `gala-from-go` reference when no row covers it, in addition to the ledger finding row.
- [x] #8 The loop's probe instructions state that a probe lives in its own directory because the transpiler loads every `.gala` file of the input directory, and that generated Go is written to a scratch path.
- [x] #9 TASK-0349's acceptance criteria are refreshed to the adopted `go.gala.fyi/stdlib` contract (its #2 currently requires twins to name no GALA runtime), or a note in TASK-0349 records why they stay as written.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Plan

1. `gala-loop` preconditions: replace the globally-clean `git status` requirement with a committed-twin-paths requirement and an explicit `backlog/backlog.md` exemption.
2. `gala-loop` One iteration: add the finalize step (commit twin halves, ledger, task notes), the no-re-probe rule for blocked entries with committed probes on the current rev, and the stale `server/.gala/` cache signal in the verify step.
3. `gala-loop` No-workaround blockers: state the probe's own-directory reason and the rule that a newly derived substitution no `gala-from-go` row covers is added there in addition to the ledger row; mirror the substitution-row rule in Workaround blockers.
4. `gala-from-go`: record `resource.Bracket`'s panic re-raise as a `*panicError` in the resource prose of `references/constructs.md` and as a trap in `SKILL.md`.
5. `server/GALA.md` cursor: state the candidate-order/dependency convention and give entries 1, 4, 5, and 6 their order and dependency notes; record entry 1's last re-verification rev.
6. TASK-0349: refresh AC #2 to the adopted `go.gala.fyi/stdlib` contract and note why.
7. Format changed Markdown with `npm run format:markdown`, then verify every AC and finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-10-10)

Repaired the instructions with each rule in its owning file; no runtime code, tests, generated twins, or build files changed.

- `.agents/skills/gala-loop/SKILL.md`: the precondition now requires committed twin-related paths and names `backlog/backlog.md` as a permanent exempt edit instead of a globally clean tree; the One iteration intro adds the no-re-probe rule for blocked entries whose committed probe was re-verified on the current rev and the rev/upstream triggers; the verify step names the stale `server/.gala/` cache signal; step 7 finalizes by committing the twin halves, ledger, and task notes; Workaround blockers and No-workaround step 5 require a newly derived substitution to get a `gala-from-go` row when none covers it; the probe step states the own-directory reason and the scratch output path.
- `.agents/skills/gala-from-go`: `references/constructs.md`'s resource prose and a new `SKILL.md` trap record that `Bracket` releases on every exit path but re-raises a passed-through panic as the runtime's unexported `*panicError`, so an outer `recover` sees an error rather than the original `string` panic value; `Unwrap`/`errors.Is`/`errors.As` still reach the original.
- `server/GALA.md` cursor: conventions now require multi-candidate entries to state order and dependencies and blocked entries to record the last re-verified rev; entries 1, 4, 5, and 6 carry those notes, with entry 1 on rev `cd2fdcb5` (2026-10-10); the loop-intro sentence now says "next workable entry".
- TASK-0349: AC #2 refreshed to the adopted `go.gala.fyi/stdlib` contract, the description's runtime-free phrasing dropped, and a comment records the refresh and the title's historical label.

## Evidence

- `grep` assertions over the four changed files confirmed the rule text for every acceptance criterion (19 checks, all pass) and confirmed the old `git status` clean requirement is gone.
- `npm run format:markdown` changed nothing; `npm run format:markdown:check` passes. The script covers tracked root Markdown (`server/GALA.md`) but excludes `.agents/**` and `backlog/**`, whose edits follow the existing sentence-per-line style by hand.
- Documentation-only change: unit/e2e tests, Swagger, and the code index are exempt because no runtime code, test, build or deployment configuration, generated artifact, or runtime asset changed. No commit was made; the Backlog MCP task edits staged the changed task and Markdown files in the index.
- Follow-up observation (outside these acceptance criteria): the `gala-from-go` `SKILL.md` resource-combinator trap still repeats the retracted claim that omitting the result type argument emits the enclosing function with no result, which `constructs.md` and `gaps.md`'s #617 entry record as wrong; the user was asked before any follow-up is created.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-10-10 16:50
---
Follow-up from the final summary created with the user's approval: TASK-0357 — Fix the stale resource-combinator result-type trap in gala-from-go. It covers the `SKILL.md` Traps bullet that still repeats the retracted signature claim contradicted by `constructs.md` and `gaps.md` #617.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Repaired the gala-loop and gala-from-go instructions so the next translation iteration does not repeat iteration 14's rework, keeping each rule in its owning file.

## Changes

- `.agents/skills/gala-loop/SKILL.md`: preconditions require twin-related paths to be committed and attributable, name `backlog/backlog.md` as a permanent exempt edit, and no longer require a globally clean `git status`; One iteration gains the no-re-probe rule for blocked entries whose committed probe was re-verified on the current rev (re-run only when the flake rev or upstream state changes), the stale `server/.gala/` cache signal in the verify step, and an explicit finalize step (step 7) committing the twin halves, ledger, and task notes; Workaround blockers and No-workaround step 5 require a newly derived substitution to get a `gala-from-go` row when none covers it; the probe step documents the own-directory reason and scratch output.
- `.agents/skills/gala-from-go`: `references/constructs.md`'s resource prose and a new `SKILL.md` trap record that `resource.Bracket` releases on every exit path but re-raises a passed-through panic as the runtime's unexported `*panicError`, so a `string` panic reaches an outer `recover` as an error rather than a `string` (with `Unwrap` still reaching the original).
- `server/GALA.md`: cursor conventions require multi-candidate entries to state order and dependencies and blocked entries to record the last re-verified rev; entries 1, 4, 5, and 6 carry those notes, with entry 1 on rev `cd2fdcb5` (2026-10-10).
- TASK-0349: AC #2 refreshed to the adopted `go.gala.fyi/stdlib` contract, the description's runtime-free phrasing dropped, and a comment records the refresh and the title's historical label.

## Verification

- 19 `grep` assertions over the changed files confirm the rule text for every acceptance criterion and the absence of the retired clean-tree requirement; all pass.
- `npm run format:markdown` changed nothing and `npm run format:markdown:check` passes. The pipeline covers `server/GALA.md` and excludes `.agents/**`/`backlog/**`, whose edits follow the existing sentence-per-line style by hand.
- Documentation-only change: no runtime code, tests, generated twins, build or deployment configuration, generated artifacts, or runtime assets changed, so unit/e2e tests, Swagger, and the code-index refresh are exempt by the Definition of Done's own conditions. No commit was made; the Backlog MCP task edits staged the changed task and Markdown files.

## Follow-ups

- The `gala-from-go` `SKILL.md` resource-combinator trap still repeats the retracted claim that omitting the result type argument emits the enclosing function with no result; `constructs.md` and `gaps.md`'s #617 entry record that claim as wrong. This is outside this task's acceptance criteria and is proposed to the user as a separate task.
<!-- SECTION:FINAL_SUMMARY:END -->
