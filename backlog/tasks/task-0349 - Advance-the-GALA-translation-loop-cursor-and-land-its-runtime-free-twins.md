---
id: TASK-0349
title: Advance the GALA translation loop cursor and land its runtime-free twins
status: In Progress
assignee: []
created_date: '2026-10-05 08:44'
updated_date: '2026-10-09 09:13'
labels: []
dependencies: []
references:
  - server/GALA.md
  - server/README.md
  - server/scripts/gala/verify.sh
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
documentation:
  - server/GALA.md
  - server/README.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
modified_files:
  - server/GALA.md
  - server/routes/users.gala
  - server/routes/users.go
  - server/discord_bot/commands/active_users.gala
  - server/discord_bot/commands/active_users.go
  - server/discord_bot/commands/active_users_extra.go
  - server/slackbot/commands/active_users.gala
  - server/slackbot/commands/active_users.go
  - server/slackbot/commands/active_users_extra.go
  - server/slackbot/commands/utils.go
  - server/models/datetime.gala
  - server/models/datetime.go
  - server/models/datetime_extra.go
  - server/models/uuid.gala
  - server/models/uuid.go
  - server/models/uuid_extra.go
  - server/scripts/gala/probes/models-defined-scalar-methods/notes.md
  - server/scripts/gala/probes/models-fixed-array/notes.md
priority: medium
ordinal: 353005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After TASK-0348 pinned GALA 0.85.0, the compiler emits declaration comments, which unblocks the cursor's first candidate that TASK-0336 had to abort: `routes/users.go`. The cursor still lists nine candidate groups, and each translation is one run of the `gala-loop` skill under the ledger's provenance, verification, and finding rules.

This task tracks that loop across the several iterations planned for it: each iteration lands at most one runtime-free twin with its registry row, its cursor removal, and its evidence, and keeps `server/GALA.md` the single source of truth for what exists and what still blocks. The first iteration (`routes/users`) is already landed on this branch and is recorded in the task notes; the remaining iterations continue from the cursor's next entry.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `server/routes/users.gala` and its generated `server/routes/users.go` are committed halves with a registry row and updated count in `server/GALA.md`, and no `routes/users.go` entry remains in the cursor.
- [ ] #2 Every twin this task lands regenerates byte-identically from its `.gala` source under `server/scripts/gala/verify.sh`, is `gofmt`-clean, and names no GALA runtime.
- [ ] #3 `go build ./...` in `server/` and the canonical Compose backend test sequence from `server/README.md` pass after each landed twin.
- [ ] #4 Swag-annotated handlers keep `server/docs` regenerating with their endpoints present, or the candidate stays handwritten with the reason recorded as a finding.
- [ ] #5 Exported Go-facing signatures, struct tags, and wire formats are unchanged, no generated file is hand-edited, and no test is relaxed.
- [ ] #6 Each iteration's finding changes are recorded in `server/GALA.md` per the loop, and a no-workaround blocker gets a committed probe under `server/scripts/gala/probes/` plus a Backlog handoff task.
- [ ] #7 The registry and the cursor stay consistent after each iteration: a landed twin has a registry row and no cursor entry, and a blocked candidate stays in the cursor with its finding.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Iteration procedure

Run the `gala-loop` skill one iteration at a time, taking the next entry from the cursor in `server/GALA.md`.

1. Sync check: compare the flake rev with `server/GALA_COMPILER`; a mismatch is its own bump iteration and never mixed with a translation.
2. Re-check the candidate's constructs on the pinned compiler before writing GALA; treat a finding recorded as fixed on this rev as a workaround candidate rather than a wall.
3. Transpile to a scratch path from the package directory, then write the real generated twin and register it in the ledger.
4. Per twin: `server/scripts/gala/verify.sh`, `go build ./...` in `server/`, and the canonical Compose test sequence in `server/README.md`; regenerate `server/docs` with `swag init` when annotations are involved.
5. Record finding changes and any no-workaround blocker per the loop before moving on.
6. Commit only task-relevant files; leave the pre-existing `backlog/backlog.md` modification alone.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Iteration 1 — routes/users (2026-10-05)

**Provenance:** `nix develop` gala `0.85.0` at `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`; flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`; extraction marker `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c` matches `server/GALA_COMPILER`. No sync needed.

**Landed:** `server/routes/users.gala` -> generated `server/routes/users.go` (runtime-free, both `var` bindings, Go-style `models.User{...}` / `responses.Error{Error: ...}` composite literals, `@Router` block emitted).

**Evidence:**
- Double transpile byte-identical (`diff` clean) and `gofmt -l` clean.
- `server/scripts/gala/verify.sh` OK (16 twins) including `go build ./...`.
- Canonical Compose backend sequence green across every package; `timeful/server/routes` 2.332s against PostgreSQL.
- `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --parseDependency` from `server/` regenerated `server/docs` byte-identically (no `git` diff) and `/users/{userId}` is present, so the annotation survived the generated file.

**Ledger:** registry row added and count 15 -> 16; cursor entry 1 removed and the list renumbered; added the note that declaration comments and swag annotations survive but an in-body comment is still dropped, so the `InitUsers` route-ordering note lives only in `users.gala`.

**State:** changes are in the worktree, not committed. The pre-existing unrelated `backlog/backlog.md` modification was left untouched.

**Next:** cursor entry 1 is now `discord_bot/commands/active_users.go`, `slackbot/commands/active_users.go`, `slackbot/commands/utils.go`. TASK-0336 recorded these blocked on `.Size()` over a Go-returned slice, but the ledger says the inferred-receiver and bare-name type-position defects were closed before the current lock; re-check the constructs on 0.85.0 before translating.

## Iteration 2 — discord_bot/commands/active_users (2026-10-05)

**Provenance:** `gala` `0.85.0` at `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`; flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`; extraction marker `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c` matches, so no sync. Iteration 1's four staged files were committed first (`150e5a03`) so both halves of every twin are committed and the tree held only the pre-existing `backlog/backlog.md` edit.

**Re-derived rows (pinned-compiler probes in `/tmp` scratch):** Go slice literal `GALA-E0007`, map literal `GALA-E0008`, `make` parse error, and `append` `GALA-E0035`; every substitute names a runtime package, so the mixed-package sibling owns them. `var a, b = f()`, `var b, _ = f()`, `&T{...}`, and assignment to existing vars are direct. `.Size()` lowers to `len` on a same-package handwritten sibling receiver but passes through on a receiver whose type is declared in an imported handwritten `.go` file, so the first generated file did not build until the counts moved behind sibling helpers.

**Landed:** `server/discord_bot/commands/active_users.gala` -> generated `active_users.go` (runtime-free), plus handwritten `active_users_extra.go` (weekday slice, empty labels/data, label/count append, chart map, and the two count helpers). Local changes only: `err` split into `repoErr`/`logsErr` because a multi-value binding cannot be redeclared, and body comments are dropped in the generated file as recorded.

**Evidence:**
- Double transpile byte-identical (`diff` clean) and `gofmt -l` clean; no `martianoff/gala` import and no `.Size()` left unlowered.
- `server/scripts/gala/verify.sh` OK (17 twins) including `go build ./...`; `go vet ./discord_bot/...` clean.
- Canonical Compose backend sequence green across every package; `timeful/server/routes` 2.216s and `timeful/server/postgres` 5.882s.

**Ledger:** registry count 16 -> 17 with the new row; cursor entry 1 keeps only the slackbot files; added the `len` on an imported handwritten slice row (workaround, sibling helper) and corrected the #613 report-index note to "fixed for same-package declarations".

**Next:** cursor entry 1 is `slackbot/commands/active_users.go` and `slackbot/commands/utils.go`; the same split is expected, with `.Size()` count helpers and a blocks-aware response constructor in the sibling.

## Iteration 3 — slackbot/commands/active_users (2026-10-05)

**Provenance:** `gala` `0.85.0` at `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`; flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`; extraction marker `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c` matches `server/GALA_COMPILER`, so no sync needed. Iteration 2's files were committed in `df626fae`, so the tree held only the pre-existing `backlog/backlog.md` edit.

**Re-derived rows (pinned compiler):** the candidate mirrors the landed `discord_bot/commands` twin; the slackbot-specific construct is the `Response` literal pair, and `newResponse` was already pinned in this package by `num_users.gala`, so a real-file scratch transpile settled the triage on the first pass.

**Landed:** `server/slackbot/commands/active_users.gala` -> generated `active_users.go` (runtime-free), plus handwritten `active_users_extra.go` (weekday slice, empty labels/data, label/count append, chart map, and chart-response constructor). Local changes only: the inline `Execute` closure moved to the named `executeActiveUsers`, `err` split into `repoErr`/`logsErr` as in iteration 2 because a multi-value binding cannot be redeclared, and body comments are dropped by the generator. The `utils.go` `newResponse` comment was generalized to the `.gala` sources.

**Evidence:**
- Double transpile byte-identical (`diff` clean) and `gofmt -l` clean; no `martianoff/gala` import and no `.Size()` left unlowered.
- `server/scripts/gala/verify.sh` OK (18 twins) including `go build ./...`.
- Canonical Compose backend sequence green across every package; `timeful/server/routes` 2.310s and `timeful/server/postgres` 5.957s.

**Ledger:** registry count 17 -> 18 with the new row; cursor entry 1 removed and the list renumbered to 1-8. No finding changed; the existing imported-handwritten-slice `len` row already covers the count helpers.

**Next:** cursor entry 1 is now the `models/` group; the loop owes it a construct re-check on 0.85.0 before any verdict.

## Iteration 4 — models group blockers documented (2026-10-05)

**Provenance:** `gala` `0.85.0` at `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`; flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`; extraction marker matches `server/GALA_COMPILER`, so no sync. Iteration 3 was committed in `0bda519e`, so the tree held only the pre-existing `backlog/backlog.md` edit.

**Re-derived rows (pinned-compiler probes):** the `models/` group stays handwritten. Struct tags, `[16]byte`, and `map[T]struct{}` are parse errors as recorded; `type DateTime int64` plus a method is refused `GALA-E0048`, and the hinted substitute `opaque type DateTime int64` transpiles but unconditionally emits `Hash`/`Compare` through `martianoff/gala/std`, so it cannot enter a runtime-free twin. No twin landed; the candidate stays in the cursor.

**Committed probes:** `server/scripts/gala/probes/models-defined-scalar-methods`, `models-struct-tag`, `models-fixed-array`, and `models-empty-struct`, each with `main.gala` and `notes.md`; generated Go was written only to `/tmp/opencode` scratch paths.

**Upstream:** `#528` (Triage language limitations) is closed upstream as of 2026-10-04, so the report index's `open` was wrong and is corrected; `#621` is closed and its `opaque type` fix is runtime-bound. No specific follow-up issue exists for tags, arrays, or `struct{}` (searched 2026-10-05).

**Ledger:** the open findings table gained the `Method on a defined scalar type` row (documented answer; `opaque type` runtime-bound) and probe paths on the tags, `struct{}`, and fixed-size-array rows; the report index `#528` and `#621` rows were corrected; cursor entry 1 keeps the models group with a probe pointer.

**Handoff:** TASK-0350 was created for the decision to comment upstream with the runtime-free `opaque type` evidence or record the wall as permanent.

**Evidence:** no runtime file changed; `server/scripts/gala/verify.sh` excludes `scripts/gala/probes/` and stays green, and `go build ./...` is unaffected. The tree was left uncommitted only for the probes and ledger until the iteration commit.

## Iteration 5 — models/datetime (2026-10-05)

**Provenance:** `gala` `0.85.0` at `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`; flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`; extraction marker `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c` matches `server/GALA_COMPILER`, so no sync needed. Iteration 4's probes and ledger were committed in `c7b3bd0a`, so the tree held only the pre-existing `backlog/backlog.md` edit.

**Re-derived rows (pinned compiler, `/tmp/opencode` scratch):** `opaque type DateTime int64` plus `Time`/`IsZero` methods and the `NewDateTimeFromTime` expression function transpile runtime-free only when a same-package `.go` sibling declares `Hash`/`Compare`; without the sibling the emitted Go imports `martianoff/gala/std`. Transpiling with the superseded handwritten `datetime.go` still present silently defeated the suppression, so the old file has to be moved aside before the transpile.

**Landed:** `server/models/datetime.gala` -> generated `datetime.go` (runtime-free: `opaque type DateTime int64`, `Time`, `IsZero`, `NewDateTimeFromTime`), plus handwritten `datetime_extra.go` (`Hash`/`Compare` suppression mirroring the runtime's FNV-1a int64 mixing and compare, and the Go-style `MarshalJSON`/pointer-receiver `UnmarshalJSON`). The two exported methods are the cost TASK-0350 left to this iteration; the repository accepted it.

**Evidence:** scratch and real transpiles are byte-identical on repeat and `gofmt -l` clean; no `martianoff/gala` import; `go doc` shows the original API plus `Hash`/`Compare`. `server/scripts/gala/verify.sh` OK (19 twins) including `go build ./...`; `go test ./models/...` and `go vet ./models/...` clean. Canonical Compose backend sequence green across every package; `timeful/server/routes` 2.466s and `timeful/server/postgres` 5.868s.

**Ledger:** registry count 18 -> 19 with the new `models` row; cursor entry 1 now lists the four remaining files; the defined-scalar finding row records the accepted cost and the move-aside-before-transpile requirement, and the prose below the table was updated; probe notes gained a Status section.

**Next:** cursor entry 1 is now `models/uuid.go`, `models/set.go`, `models/location.go`, `models/event.go`; uuid carries the defined-type path plus fixed-size `[16]byte` helpers, and the other three stay tagged/empty-struct blocked.

**State:** Changes are in the worktree, not committed: the three `server/models/datetime.*` files, the `server/GALA.md` updates, the probe-note Status section, and this task's notes. The pre-existing unrelated `backlog/backlog.md` modification was left untouched.

## Iteration 6 — models/uuid (2026-10-05)

**Provenance:** `gala` `0.85.0` at `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`; flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`; extraction marker `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c` matches `server/GALA_COMPILER`, so no sync needed. Iteration 5's files were committed in `51250af5`, so the tree held only the pre-existing `backlog/backlog.md` edit.

**Re-derived rows (pinned compiler):** `opaque type UUID string` is the documented `opaque type` answer for the defined-scalar method, and the file splits at the fixed-size `[16]byte` helpers (`NewUUID`/`formatUUID`); the UUID spelling of the suppression path was already pinned in scratch during iteration 5, and the real-file transpile settled the rest on the first pass.

**Landed:** `server/models/uuid.gala` -> generated `uuid.go` (runtime-free: `opaque type UUID string`, `ZeroUUID`, `String`, `IsZero`), plus handwritten `uuid_extra.go` (`Hash`/`Compare` suppression, `ParseUUID`, `isLowerHex`, `NewUUID`, `formatUUID`, and the JSON/text marshalers). The old `uuid.go` was moved aside before the transpile, per the datetime requirement; the exported API adds only `Hash`/`Compare`, the cost TASK-0350 accepted.

**Evidence:**
- Double transpile byte-identical (`diff` clean) and `gofmt -l` clean; no `martianoff/gala` import.
- `server/scripts/gala/verify.sh` OK (20 twins) including `go build ./...`; `go test ./models/...` and `go vet ./models/...` clean; `go doc` shows the original API plus `Hash`/`Compare`.
- Canonical Compose backend sequence green across every package; `timeful/server/routes` 2.830s and `timeful/server/postgres` 7.227s.

**Ledger:** registry count 19 -> 20 with the new row; cursor entry 1 now lists `models/set.go`, `models/location.go`, `models/event.go`; the defined-scalar row and the models prose record uuid; the fixed-array row and both probe notes record that the array helpers moved to the sibling.

**Next:** cursor entry 1 is `models/set.go`, `models/location.go`, `models/event.go`, all blocked on struct tags and `struct{}` with probes at `scripts/gala/probes/models-*`.

**State:** changes are in the worktree, not committed: the three `server/models/uuid.*` files, the `server/GALA.md` updates, the two probe-note updates, and this task's notes. The pre-existing unrelated `backlog/backlog.md` modification was left untouched.

## Iteration 7 — compiler bump 0.85.0 -> 0.87.1 (2026-10-09)

**Provenance:** `nix flake update gala` moved the lock from `a888e824ff53adb653bbd48dcc83f67788eeced4` (`0.85.0`) to `669c958cfb1f0ef2fe69ecb8e199b33c88456664`; `gala version` reports `GALA version 0.87.1` at `/nix/store/sy03nzcg76p7429jkrkgkpgj73fhzvrm-gala-0.87.1/bin/gala`. After deleting `~/.gala/stdlib/v0.87.1/.stdlib-extracted` and transpiling `server/eventid/eventid.gala` to a scratch path, the marker reads `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa`, which is now in `server/GALA_COMPILER`. Bump-only iteration; no translation. The tree held only the pre-existing `backlog/backlog.md` edit and untracked `TODO.md`.

**Evidence:**
- `server/scripts/gala/verify.sh --write` then `verify.sh`: OK (20 twins), with no diff to any generated twin, so the bump changed no twin shape.
- Canonical Compose backend sequence green across every package (`routes` 2.380s, `postgres` 5.322s). The README's `docker volume create timeful-test-go-build-cache timeful-test-go-mod-cache` fails on this Docker version (`requires at most 1 argument`), so the volumes were created one at a time. The existing `.env.test` (one local line beyond the example) was kept rather than overwritten by `cp`.
- The four `scripts/gala/probes/models-*` probes reproduced their recorded diagnostics and output unchanged.

**Upstream between revs (22 commits):** PR #722 (Go result lists on GALA functions and lambdas), #731/#732 (`gala stdlib export`, `gala export`, and the Go module `go.gala.fyi/stdlib`, which closed #698 on 2026-10-09), and the #697 fix, plus match/interface fixes.

**Finding change:** a scratch transpile showed that `func ParsePort(s string) (int, error) = strconv.Atoi(s)` emits a plain `return strconv.Atoi(s)` with no runtime import, while a `Success`/`Failure` body emits `std.Try` and imports `martianoff/gala/std`. The "Go-style multi-value return signature" row moves from boundary gap to workaround, and the `gala-from-go` construct row is updated to match (no longer a parse error). The #698 row in the report index is now closed.

**Ledger:** the `server/GALA.md` compiler paragraph, table, and #698 prose; the findings row and prose; the report index; and the probe re-verification line. Formatted with `npm run format:markdown`.

**Next:** cursor entry 1 is unchanged (`models/set.go`, `models/location.go`, `models/event.go`, still blocked on struct tags and `struct{}`).

**State:** changes are in the worktree, not committed: `flake.lock`, `server/GALA_COMPILER`, `server/GALA.md`, `.agents/skills/gala-from-go/references/constructs.md`, and this task's notes.

## Iteration 8 — `errs/errors.go` split (2026-10-09)

**Provenance:** `gala version` `GALA version 0.87.1` at `/nix/store/sy03nzcg76p7429jkrkgkpgj73fhzvrm-gala-0.87.1/bin/gala`; flake rev `669c958cfb1f0ef2fe69ecb8e199b33c88456664` matches `server/GALA_COMPILER`, so no bump. The tree held only the pre-existing `backlog/backlog.md` edit and untracked `TODO.md`, both left untouched.

**Candidate:** cursor entry 1 (`models/set.go`, `models/location.go`, `models/event.go`) stays blocked on struct tags and `struct{}`; its probes reproduced on this rev in iteration 7, so the loop took entry 2, `errs/errors.go`.

**Triage:** split. `GoogleAPIError` carries `json` tags that define its wire shape (boundary gap), so it and its pointer-receiver `Error` method move unchanged to the handwritten sibling `server/errs/google_api_error.go`. The 15 error-code strings become `errs/errors.gala` via the `const` row's `var` substitute, the same path `appenv` took; the two untyped consts stay untyped `var X = ...` (inferred `string`) and the typed block stays `var X string = ...`. The old `errors.go` was moved aside to scratch before the transpile.

**API cost:** the names become assignable package vars rather than compile-time constants. `grep` found no caller using any `errs.*` name in a constant expression, array length, or `const` declaration; `go doc ./errs` lists the same 15 names plus `GoogleAPIError`, and the `Errors enum` declaration comment survives on `NotSignedIn`.

**Evidence:**
- Double transpile byte-identical, `gofmt -l` clean, no `martianoff/gala` import, `go vet ./errs/` clean.
- `server/scripts/gala/verify.sh`: OK (21 twins), including `go build ./...`.
- Canonical Compose backend sequence green across every package (`routes` 2.059s, `postgres` 5.274s; `errs` has no test files). Existing `.env.test` kept; volumes created one per command.
- `server/docs` names nothing from `errs`, so no swag regeneration was needed.

**Ledger:** registry count 20 -> 21 (fifteen packages) with the `errs` row; cursor entry 2 removed and the rest renumbered; new open-findings row for `const` declarations that a Go caller reads (language gap, `var` workaround, deliberately unfiled); prose records the split. Formatted with `npm run format:markdown`.

**Next:** cursor entry 1 is still blocked; entry 2 is now `routes/respondent_identity.go`, `routes/group.go`.

**State:** changes are in the worktree, not committed: `server/errs/errors.gala`, `server/errs/errors.go`, `server/errs/google_api_error.go`, `server/GALA.md`, and this task's notes.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: TASK-0350
created: 2026-10-05 09:37
---
TASK-0350 resolved 2026-10-05 with no upstream report: the `opaque type` runtime binding is suppressible. A same-package `.go` sibling declaring `Hash`/`Compare` keeps the generated Go free of `martianoff/gala` (verified on the pinned rev), at the cost of two added exported methods. `server/GALA.md` cursor entry 1 and the defined-scalar finding row now record that path, and the probe notes carry the repro; upstream PR #665 documents the suppression, so no comment on #528/#621 was needed. Struct tags, fixed-size arrays, and `struct{}` stay handwritten under `gaps.md`'s "What Not To File".
---
<!-- COMMENTS:END -->
