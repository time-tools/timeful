---
id: TASK-0357
title: Fix the stale resource-combinator result-type trap in gala-from-go
status: Done
assignee:
  - '@opencode'
created_date: '2026-10-10 16:50'
updated_date: '2026-10-10 16:54'
labels:
  - gala
dependencies: []
references:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
  - >-
    backlog/tasks/task-0356 -
    Repair-the-gala-loop-and-gala-from-go-instruction-gaps-found-landing-postgres-pool.md
documentation:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
modified_files:
  - .agents/skills/gala-from-go/SKILL.md
priority: medium
type: docs
ordinal: 361300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`gala-from-go`'s Traps list contradicts its own references on the resource combinators. The trap bullet in `.agents/skills/gala-from-go/SKILL.md` says omitting a combinator's result type argument "defaults its *result* to `any`" and the enclosing function "is emitted with no result at all". `references/constructs.md`'s resource section and `references/gaps.md`'s #617 entry record that claim as retracted: the enclosing function's return type is emitted unchanged in every variant checked, and the real failure is a body parameter degraded to `any` when the resource type is declared in a handwritten `.go` sibling (#618), which `go build` rejects. The same trap also carries the still-true guidance that a function-typed result cannot be spelled as an explicit type argument.

Rewrite the single trap bullet so it states the actual failure modes and drops the retracted signature claim, and confirm the three files agree afterwards. This is the follow-up recorded in TASK-0356's final summary; it is a documentation-only change with no runtime file changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Traps entry for resource combinators in `.agents/skills/gala-from-go/SKILL.md` no longer says a missing result type argument drops the enclosing function's return type or emits a resultless function.
- [x] #2 The entry states the actual failure modes: a resource type not resolvable from `.gala` text binds the body parameter to `any` and does not build (#618), the enclosing function's return type is emitted unchanged, and a function-typed result cannot be an explicit type argument (`resource.Bracket[context.CancelFunc, func()]` is a parse error), so the type is named (`type CloseFn func()`) or the binding is annotated (`var closeFn func() = resource.Bracket(cancel, release, body)`).
- [x] #3 The entry agrees with `references/constructs.md`'s resource-combinator prose and `references/gaps.md`'s #617/#618 entries, with no remaining contradiction among the three files.
- [x] #4 No runtime code, test, generated twin, or build or deployment file changes.
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

1. Rewrite the single resource-combinator trap bullet in `.agents/skills/gala-from-go/SKILL.md` so it:
   - titles the actual failure mode (a resource type not resolvable from `.gala` text binds the body parameter to `any`, #618) rather than the retracted signature claim;
   - states the enclosing function's return type is emitted unchanged, and that the void-function failure belongs to `func F() = <expr>` never inferring a result (#617);
   - keeps the still-true guidance that a function-typed result cannot be an explicit type argument (`resource.Bracket[context.CancelFunc, func()]` is a parse error), so the type is named (`type CloseFn func()`) or the binding is annotated (`var closeFn func() = resource.Bracket(cancel, release, body)`).
2. Verify with grep that the retracted claim is gone and the bullet agrees with `references/constructs.md`'s resource prose and `references/gaps.md`'s #617/#618 entries.
3. Run `npm run format:markdown` (and check) for the changed Markdown; confirm whether `.agents/**` is covered.
4. Check acceptance criteria and DoD items with the evidence, record notes and the final summary, and set the task Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-10-10)

Rewrote the single resource-combinator trap bullet at `SKILL.md:236-240` and confirmed the three files agree; no runtime code, tests, generated twins, or build files changed.

- Retracted-claim absence checks over `SKILL.md`: grep for `emitted with no result`, `drops the enclosing function`, `defaults its *result*`, `no result at all`, and `Supply both type arguments` returns no matches.
- Presence checks: the bullet now names the resource type resolvable from `.gala` text; the Go-sibling failure (`body parameter binds to \`any\``, `does not build`, #618 link); the partial type-argument list failure (transpiler's own type-parameter name); the unchanged enclosing return type and `func F() = <expr>` (#617 link); and the function-typed-result route (`resource.Bracket[context.CancelFunc, func()]` parse error, `type CloseFn func()`, `var closeFn func() = resource.Bracket(cancel, release, body)`).
- Cross-file agreement: new bullet matches `references/constructs.md:205-209` and `references/gaps.md:177-184`; no remaining contradiction found among the three files' resource-combinator text.
- `npm run format:markdown:check` passes; `scripts/markdown.mjs` excludes `.agents/**`, so the edited file was kept in the existing sentence-per-line style by hand. `npx vitest run prettier/markdown/` passes (30 tests).
- `git status --porcelain` shows only `SKILL.md` plus the Backlog-managed `backlog/backlog.md`; no runtime code, tests, generated twins, or build or deployment files changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Rewrote the stale resource-combinator trap in `.agents/skills/gala-from-go/SKILL.md` so the Traps list states the failure modes its own references (`references/constructs.md`, `references/gaps.md`) already record, dropping the retracted signature claim.

## Changes

- `.agents/skills/gala-from-go/SKILL.md` (trap bullet, lines 236-240): now states that a resource type `.gala` text cannot resolve — one declared in a handwritten `.go` sibling — needs both type arguments written out, or the body parameter binds to `any` and the generated file does not build ([#618](https://github.com/martianoff/gala/issues/618)); that a partial type-argument list emits the transpiler's own type-parameter name; that the enclosing function's return type is emitted unchanged in every variant checked, with `func F() = <expr>` as the real no-inference defect ([#617](https://github.com/martianoff/gala/issues/617)); and keeps the still-true function-typed-result guidance (explicit type argument is a parse error, so name the type or annotate the binding). The retracted "defaults its *result* to `any` / emitted with no result" claim and the misattributed `too many return values` errors are gone.

## Verification

- Absence greps over `SKILL.md` confirm the retracted phrasing (`emitted with no result`, `drops the enclosing function`, `defaults its *result*`, `no result at all`, `Supply both type arguments`) is gone; presence greps confirm every AC-required clause; cross-file greps show the bullet matches `constructs.md:205-209` and `gaps.md:177-184`, with no remaining contradiction.
- `npm run format:markdown:check` passes; `scripts/markdown.mjs` excludes `.agents/**` from the format pipeline, so the edit follows the existing sentence-per-line style by hand. `npx vitest run prettier/markdown/` passes (30 tests).
- `git status`/`git diff --stat` show only `SKILL.md` plus the Backlog-managed `backlog/backlog.md`; no runtime code, tests, generated twins, or build or deployment files changed. Documentation-only change, so unit/e2e, Swagger, code-index, and `fmt:check` are exempt by the DoD's own conditions. No commit was made.

## Follow-ups

- None identified.
<!-- SECTION:FINAL_SUMMARY:END -->
