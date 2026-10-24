---
id: TASK-0359
title: >-
  Repair the stale GALA ledger rows for the importer-direction method set and
  the select substitute
status: Done
assignee: []
created_date: '2026-10-18 10:35'
updated_date: '2026-10-18 10:40'
labels: []
dependencies: []
documentation:
  - server/GALA.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
  - backlog/drafts/issue-bracket-panic-reraise.md
ordinal: 363300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Two GALA rows became stale and one overstated its answer, all re-derived on the pinned compiler (0.90.0 / rev 014c2f23) and on 0.91.0 (rev a5206156).

1. The importer-direction method-set finding (server/GALA.md open findings, constructs.md package rows, SKILL.md guidance and Trap, gaps.md #615 entry) claims an imported package's GALA index becomes a receiver's whole method set, refusing an importer's call to a sibling-declared method with a false GALA-E0044. Upstream now records sibling methods in a GoKindMethodsOnly record that the unknown-method check consults, so the call resolves and builds.
2. The select row was marked answered, but concurrent.Future/Race cover only the timeout and first-wins idioms; a multiplex over channels the function does not own has no substitute, which is why main_run.go stays handwritten.
3. resource.Bracket re-raises a body panic as an unexported *std.panicError rather than the panicked value, which deserves a report draft.

Scope: documentation and agent instructions only. No runtime, test, twin, or build change.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The importer-direction method-set claim is corrected in server/GALA.md, references/constructs.md, references/gaps.md, and SKILL.md, and no file still says an importer's call to a sibling-declared method is refused with GALA-E0044.
- [x] #2 The select row distinguishes the answered timeout and first-wins shapes from the unsubstituted multiplex over foreign channels, and gaps.md no longer forbids filing the latter.
- [x] #3 A panic re-raise report draft exists at backlog/drafts/issue-bracket-panic-reraise.md in the gaps.md template shape, backed by an executed repro.
- [x] #4 npm run format:markdown:check passes and no runtime, test, twin, or build file changes.
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Evidence

**Importer-direction method set (corrected, ledger).** Two-package repro: `lib/repo.gala` declares `struct Repo` + `Name()`; `lib/repo.go` declares `Load()` and `NewRepo()`; `app/main.gala` imports `mstest/lib` and calls `r.Load()`. On 0.90.0 (`0rd7iwl6...-gala-0.90.0`) and 0.91.0 (`r5ivr247...-gala-0.91.0`) the importer transpiles clean and emits `return r.Load()`; `go mod tidy && go build ./...` passes. Upstream cause: `internal/transpiler/analyzer/go_types.go` files sibling methods under `GoKindMethodsOnly` in `GoTypeInfo.GalaTypeMethods`, and `transformer/unknown_method.go` consults it before raising E0044. `analyzer/cache.go` v11 records the cross-package intent.

**select.** Re-ran a two-channel `select` under both revs: identical parse error (`missing ')' at 'int'`). `main_run.go` stays handwritten for this reason.

**Bracket panic.** Mixed package: `main.gala` calls `resource.Bracket(1, (r int) => {}, body)` with a panicking body; handwritten `harness.go` recovers. Executed output: `recovered type: *std.panicError` / `value: boom`. `std/types.go:219` declares `type panicError struct` unexported. `resource/resource_test.gala` has no panic test.

## Checks
- `npm run format:markdown:check` passes.
- No runtime, test, twin, generated, or build file changed by this task. `server/observability/readiness.go` and its .gala siblings were already modified in the worktree before this task started (a concurrent iteration) and were left untouched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Corrected the stale importer-direction method-set claim across the four GALA documents and split the select row by shape.

The *Repository finding in server/GALA.md said an imported package's GALA index becomes a receiver's whole method set, so an importer's call to a sibling-declared method is refused with a false GALA-E0044. It no longer is: a two-package repro transpiles clean, emits r.Load(), and builds on both 0.90.0 and 0.91.0. Updated the finding row, the #615 report-index row, the repository prose, the constructs.md row and prose, the SKILL.md guidance bullet and Trap, and the gaps.md defect entry, each with the date and rev that re-derived it. The split itself is kept, now as a readability choice rather than a compiler constraint.

The select row said answered with concurrent.Future or go_interop, but those cover only the timeout and first-wins idioms. A multiplex over channels the function does not own has no substitute, and gaps.md listed select among constructs not to file. Split the row by shape and added a control-flow note, plus a gaps.md exception so that shape can be filed.

Wrote backlog/drafts/issue-bracket-panic-reraise.md in the gaps.md template: resource.Bracket re-raises a body panic as the unexported *std.panicError, so an outer recover sees a wrapper rather than the panicked string, against Bracket's own comment. Backed by an executed repro.

Documentation-only; format:markdown:check passes. No runtime, test, twin, or build file changed.
<!-- SECTION:FINAL_SUMMARY:END -->
