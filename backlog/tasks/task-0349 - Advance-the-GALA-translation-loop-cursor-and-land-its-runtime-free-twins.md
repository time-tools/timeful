---
id: TASK-0349
title: Advance the GALA translation loop cursor and land its runtime-free twins
status: In Progress
assignee: []
created_date: '2026-10-05 08:44'
updated_date: '2026-10-26 11:56'
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
  - server/postgres/repository.gala
  - server/postgres/repository.go
  - server/postgres/repository_methods.go
  - server/postgres/repository_types.go
  - server/postgres/codec.gala
  - server/postgres/codec.go
  - server/postgres/pool.gala
  - server/postgres/pool.go
  - server/postgres/tracing.gala
  - server/postgres/tracing.go
  - server/main.gala
  - server/main.go
  - server/main_run.go
  - server/observability/provider.gala
  - server/observability/provider.go
  - server/observability/provider_recorder.go
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
priority: medium
ordinal: 353005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After TASK-0348 pinned GALA 0.85.0, the compiler emits declaration comments, which unblocks the cursor's first candidate that TASK-0336 had to abort: `routes/users.go`. The cursor still lists nine candidate groups, and each translation is one run of the `gala-loop` skill under the ledger's provenance, verification, and finding rules.

This task tracks that loop across the several iterations planned for it: each iteration lands at most one twin with its registry row, its cursor removal, and its evidence, and keeps `server/GALA.md` the single source of truth for what exists and what still blocks. The first iteration (`routes/users`) is already landed on this branch and is recorded in the task notes; the remaining iterations continue from the cursor's next entry.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `server/routes/users.gala` and its generated `server/routes/users.go` are committed halves with a registry row and updated count in `server/GALA.md`, and no `routes/users.go` entry remains in the cursor.
- [ ] #2 Every twin this task lands regenerates byte-identically from its `.gala` source under `server/scripts/gala/verify.sh`, is `gofmt`-clean, and follows the adopted runtime contract: it imports only `go.gala.fyi/stdlib/...` from the module pinned in `server/go.mod`, or stays runtime-free where a runtime construct would change its Go-facing surface, and it never names `martianoff/gala`.
- [ ] #3 `go build ./...` in `server/` and the canonical Compose backend test sequence from `server/README.md` pass after each landed twin.
- [ ] #4 Swag-annotated handlers keep `server/docs` regenerating with their endpoints present, or the candidate stays handwritten with the reason recorded as a finding.
- [ ] #5 Exported Go-facing signatures, struct tags, and wire formats are unchanged, no generated file is hand-edited, and no test is relaxed.
- [ ] #6 Each iteration's finding changes are recorded in `server/GALA.md` per the loop, and a no-workaround blocker gets a committed probe under `server/scripts/gala/probes/` plus a Backlog handoff task.
- [ ] #7 The registry and the cursor stay consistent after each iteration: a landed twin has a registry row and no cursor entry, and a blocked candidate stays in the cursor with its finding.
<!-- AC:END -->

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

## Iteration 20 — fourth batch: five twins (2026-10-25)

**Provenance:** `gala` `GALA version 0.93.0` at `/nix/store/v0jxpys9mxb58dvdp3np8bywpchbyppa-gala-0.93.0/bin/gala`; flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c` matches `server/GALA_COMPILER` (extraction `0.93.0 6368c093…`), so no bump. The user asked for ten more files, then narrowed the scope to what is realistically landable; five candidates were taken and one `verify.sh` run covers them.

**Landed (5 twins):**
- `routes/event_timed_slots.gala` -> `event_timed_slots.go`; `event_timed_slots_extra.go` keeps the tagged `timedEventPayloadFields`.
- `observability/metrics.gala` -> `metrics.go`; `metrics_extra.go` keeps `PoolStats` and the two `*Recorder` metric-building methods.
- `routes/analytics.gala` -> `analytics.go`; `analytics_extra.go` keeps `scannedPosterPayload` and its constructor.
- `routes/identity.gala` -> `identity.go`; `identity_extra.go` keeps `visitor`, `newVisitor`, `hashBytes`, and the association request bodies.
- `routes/owner.gala` -> `owner.go`; `owner_extra.go` keeps `archiveEventInput`, its constructor, and `replaceEvent`.

**Construct facts re-derived on the pinned compiler:** a `struct{}` membership set that never leaves one function is `map[K]bool`; `sha256.Sum256`'s `[32]byte` digest stays behind a sibling `hashBytes` and the 32-byte secret is `go_interop.SliceWithSize[byte](32)`; a local binding named after a package type makes the generated `std.Failure[*T]{}` name the local (`visitor` -> `visitorRecord`); `var _ = f()` is `GALA-E0060`; a `(bool, error)` list cannot use `return a, b`; populated `gin.H{...}` is `gin.H{}` plus index assignment; a `gin.HandlerFunc` closure infers `*gin.Context`; cross-`.gala`-file calls resolve once the callee file exists.

**Evidence:** `server/scripts/gala/verify.sh` -> OK (63 twins) in 14m18s (63/63 byte-identical double transpiles, `gofmt` 63 checked, orphan scan over 211 Go files, `go build ./...` ok); `go vet ./...` clean; the canonical isolated Compose sequence passes (`routes` 4.695s, `observability` 0.501s, `postgres` 12.970s); `swag init --parseDependency` regenerated `server/docs` byte-identically.

**Ledger:** registry 58 -> 63 sources (packages stay 26); the cursor drops the four landed `routes`/`observability` entries and now lists `routes/transfers.go`, `routes/auth.go`, `routes/user.go`/`routes/event_routes.go`, and the `postgres` free-function splits; the `struct{}` and fixed-array findings record the `map[K]bool` and `hashBytes` substitutions; a new finding row records the type-shadowing defect; `constructs.md` gained the shadowing and blank-result rows plus the local-set and cross-file prose; a batch prose section records every construct fact.

**Next:** cursor entry 2 is `routes/transfers.go`.

**State:** committed as the fourth batch commit; no generated file hand-edited and no test relaxed.
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

## Iteration 9 — `routes/respondent_identity.go` split (2026-10-09)

**Provenance:** `gala version` `GALA version 0.87.1` at `/nix/store/sy03nzcg76p7429jkrkgkpgj73fhzvrm-gala-0.87.1/bin/gala`; flake rev `669c958cfb1f0ef2fe69ecb8e199b33c88456664` matches `server/GALA_COMPILER`, so no bump. Iteration 8 was committed (`21335f08`, then `9af76d44`); the tree held only the pre-existing `backlog/backlog.md` edit and untracked `TODO.md`, both left untouched.

**Candidate:** cursor entry 1 stays blocked: `scripts/gala/probes/models-struct-tag` and `models-empty-struct` re-run on this rev reproduce their recorded parse errors. The loop took entry 2's first file, `routes/respondent_identity.go`.

**Triage:** split. A whole-file scratch transpile (old file moved aside to scratch first) showed the four single-result helpers emit plain Go: `var clone = *user; return &clone`, the three-value `var firstName, lastName, _ = ...`, field assignment through the pointer, an `if` expression in `canonicalGuestName`, and a `match` over qualified `respondents.GuestName*` constants that lowers to `obj == respondents.X`. `populateSignUpResponsePayloadIdentity` returns `(string, bool)` from a built body, so under the multi-value return row it emitted `std.Tuple` and imported `martianoff/gala/std`; it moves verbatim, with its doc comment, to the handwritten sibling `server/routes/respondent_identity_extra.go`. Its comma-ok map read `var v, ok = m[k]` emitted plain Go and is not a blocker. All five names are unexported and every caller is in `routes`, so no Go-facing signature changed.

**Evidence:**
- Double transpile byte-identical, `gofmt -l` clean on both files, no `martianoff` import and zero `std.` references, relative `//line` directives, `go build ./...` and `go vet ./routes` clean.
- `server/scripts/gala/verify.sh`: OK (22 twins), including `go build ./...`.
- Canonical Compose backend sequence green across every package (`routes` 2.061s, `postgres` 5.969s); `routes/guest_response_ownership_test.go` calls `canonicalGuestName` and `guestNameValidationErrorMessage` directly. Existing `.env.test` kept.
- No swag annotations in the file, so no `server/docs` regeneration. `codebase-memory-mcp` is not on `PATH`, so the index refresh was not run.

**Ledger:** registry count 21 -> 22 (still fifteen packages) with the `routes/respondent_identity` row; cursor entry 2 is now `routes/group.go` alone; prose records the split. No finding changed, because the multi-value return workaround row already covers the sibling. Formatted with `npm run format:markdown`.

**Next:** cursor entry 1 is still blocked; entry 2 is `routes/group.go` (785 lines).

**State:** changes are in the worktree, not committed: `server/routes/respondent_identity.gala`, `server/routes/respondent_identity.go`, `server/routes/respondent_identity_extra.go`, `server/GALA.md`, and this task's notes.

## Iteration 10 — compiler bump 0.87.1 post-release HEAD (2026-10-10)

**Provenance:** `nix flake update gala` moved the lock from `669c958cfb1f0ef2fe69ecb8e199b33c88456664` to `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c`; `gala version` reports `GALA version 0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`. After deleting `~/.gala/stdlib/v0.87.1/.stdlib-extracted` and transpiling `server/eventid/eventid.gala` to a scratch path, the marker reads `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa`, unchanged from the previous lock, so the bump is codegen-only. `server/GALA_COMPILER` records the new rev and date. Bump-only iteration; no translation. The tree held only the pre-existing `backlog/backlog.md` edit and untracked `TODO.md`, both left untouched.

**Upstream between revs (5 commits):** #737 (nix pin of the stdlib transpiler to the 0.87.1 release), #742 (fixes #739: `case 1 | 2` alternative patterns with new `GALA-E0071`), #743 (fixes #740: `gala transpile --stdlib-module`), #746 (sealed match exhaustiveness requires full coverage of each variant).

**Evidence:**
- `server/scripts/gala/verify.sh --write` then `verify.sh`: OK (22 twins), with no diff to any generated twin, so the bump changed no twin shape.
- Canonical Compose backend sequence green across every package (`routes` 3.844s, `postgres` 11.626s); cache volumes created one per command, existing `.env.test` kept.
- Scratch transpiles on the pinned rev: `case 1 | 2` emits `if obj == 1 || obj == 2` (finding #739 closed); `gala transpile --stdlib-module go.gala.fyi/stdlib` rewrote an emitted `martianoff/gala/std` import to `go.gala.fyi/stdlib/std` (finding #740 closed as a transpile capability).

**Finding/ledger changes:** the `Several values in one case` and `GALA runtime from go.gala.fyi/stdlib in a transpiled twin` rows are deleted and recorded in the report index as fixed on the pinned rev; the compiler paragraph, provenance table, version-string caveat, and #698 prose updated; the contract now names both runtime import paths and `verify.sh` rejects both; the `gala-from-go` `fallthrough`/`several values in one case` rows and the `|` trap updated to the alternative-pattern behavior with `GALA-E0071`.

**Next:** cursor entry 1 is unchanged (`models/set.go`, `models/location.go`, `models/event.go`, blocked on struct tags and `struct{}`). The runtime-free contract is unchanged, so `--stdlib-module` stays unused; revisiting that contract is a separate decision.

**State:** changes are in the worktree, not committed: `flake.lock`, `server/GALA_COMPILER`, `server/GALA.md`, `server/scripts/gala/verify.sh`, `.agents/skills/gala-from-go/SKILL.md`, `.agents/skills/gala-from-go/references/constructs.md`.

**Handoff (2026-10-10):** TASK-0353 (`Decide whether server GALA twins adopt the go.gala.fyi/stdlib module`) was created for the runtime-free-contract decision raised by iteration 10. The loop keeps the current contract until TASK-0353 decides otherwise; do not start consuming `--stdlib-module` from the cursor.

## Iteration 11 — routes/group split (2026-10-10)

**Provenance:** `gala version` `GALA version 0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c` matches `server/GALA_COMPILER`, so no bump. The tree held only the pre-existing `backlog/backlog.md` edit, left untouched.

**Candidate:** cursor entry 1 (`models/set.go`, `models/location.go`, `models/event.go`) stays blocked: `scripts/gala/probes/models-struct-tag` and `models-empty-struct` re-ran on this rev and reproduced their recorded parse errors. The loop took entry 2, `routes/group.go` (785 lines), on its 2026-10-10 pre-triage.

**Triage:** split. `eventInput` (embedded `models.Event` promotion plus tag), `groupAttendeePayload` (`json` tags) with `groupAttendeePayloads`, and `getCalendarAvailabilities` (`chan`, `go`, in-place `recover`) moved unchanged to the handwritten sibling `server/routes/group_extra.go`. `declineInvite`'s anonymous tagged body became the sibling's named `declineBody`, and `guestForbidden`/`guestNameError` construction moved behind the sibling helpers `newGuestForbidden`/`newGuestNameError`, so all of `declineInvite` and `mutateGroupResponse` stayed in the `.gala` file. The old `group.go` was moved aside before the transpile.

**Re-derived rows (pinned compiler, `/tmp/opencode` scratch):** a Go result list with a built body must return `Success(...)`/`Failure(...)` values; a tuple `return (v, err)` in an error-result body emits `std.Try[T]` with `std.Tuple` values and does not build, while `(T, bool, error)` lowers through `Try[Tuple[T, bool]]` and builds. A closure passed to a Go function carries an explicitly annotated parameter list and may assign to a captured outer `var`. A struct's fields are all required at construction unless the declaration gives a default (`GALA-E0045`), and only `var` fields stay plain Go fields. A sibling-declared struct works as a `var` type with `&input`, field reads, and `*input.Field`. The byte truncation is `string(go_interop.SliceTake(go_interop.ToBytes(s), n))`; the recorded `out += string(s[i])` loop re-encodes bytes at or above 0x80. `go_interop.MapPut` needs both type arguments for a typed value into a `map[string]any`. A standalone `import "martianoff/gala/go_interop"` declaration is required for the emitted block to be `gofmt`-clean under `--stdlib-module`, because the rewrite makes its path sort before `io` while the transpiler sorts by source path.

**Landed:** `server/routes/group.gala` -> generated `server/routes/group.go` (stdlib: `applyGroupAttendeeEdits`'s `(groupEmailPlan, error)` Try lowering, `mutateGroupResponse` with its transaction lambda, the `groupManualAvailability` alias, both parse/merge helpers, and `declineInvite`), plus handwritten `server/routes/group_extra.go`.

**Evidence:**
- Double transpile byte-identical (`diff` clean) and `gofmt -l` clean on both files; no `martianoff` import.
- `server/scripts/gala/verify.sh`: OK (23 twins), including `go build ./...`; `go vet ./...` clean.
- Canonical Compose backend sequence green across every package (`routes` 2.636s, `postgres` 6.080s); `swag init --parseDependency` (v1.16.6) regenerated `server/docs` byte-identically.
- The only caller of `applyGroupAttendeeEdits` (`routes/event_routes.go:746`) discards the plan on error, so the Try lowering's zero-value-on-failure path preserves behavior; `group_test.go` reads `groupInviteEmailTemplate` as a plain int, so both template ids stayed `var`.

**Ledger:** registry count 22 -> 23 (fifteen packages) with the `routes` group row; cursor entry 2 removed and the list renumbered; the group prose and the embedded/channel/recover/anonymous-struct rows point at `group_extra.go`; the multi-value return row gained the `Success`/`Failure` caveat; two workaround rows added (`MapPut` explicit type arguments, standalone runtime import for `gofmt`); `gaps.md`'s `MapPut` paragraph and `constructs.md`'s truncation row and struct-default note corrected. Formatted with `npm run format:markdown`.

**Next:** cursor entry 1 is still `models/set.go`, `models/location.go`, `models/event.go`; entry 2 is now `postgres/`.

**State:** changes are in the worktree, not committed: `server/routes/group.gala`, `server/routes/group.go`, `server/routes/group_extra.go`, `server/GALA.md`, the two `.agents` references, and this task's notes. The pre-existing unrelated `backlog/backlog.md` modification was left untouched.

## Iteration 12 — postgres/repository declarations-only split (2026-10-10)

**Provenance:** `gala` `GALA version 0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c` and extraction marker match `server/GALA_COMPILER`, so no bump. The tree held the pre-existing `backlog/backlog.md` edit (left untouched) and a concurrent TASK-0355 session that committed `c3682df7` (docs-only: #749/#750 filed, the fixed `len` finding retired) mid-iteration; this iteration builds on that HEAD and leaves TASK-0355's files alone.

**Candidate:** cursor entry 1 (`models/set.go`, `models/location.go`, `models/event.go`) stays blocked: `scripts/gala/probes/models-empty-struct` and `models-struct-tag` re-ran on the pinned rev and reproduced their recorded parse errors. The loop took entry 2, `postgres/`, and the sibling-method note re-checked: a `.gala` file can call methods declared in a handwritten sibling (the F8 wall is closed on this rev, probe `caller.gala` calling `r.WithTransaction`).

**Re-derived rows (pinned compiler):** `type dbtx interface` with unnamed parameters panics the transformer (`GALA-E0017`); naming the parameters emits the same interface. An anonymous `interface{ Scan(...any) error }` parameter is a parse error; a named unexported interface with the same method set is the substitute (`rowScanner`). Integer literal separators (`4_000_000_000`) are a parse error; plain digits work. `std.As[T](v)` is callable directly with `import "martianoff/gala/std"` and emits the same `std.As` a typed-pattern match lowers to, which is the comma-ok substitute the match row does not cover. A typed-pattern match over a sibling struct's field whose type is declared only in the `.gala` file is refused with `cannot infer type of matched expression`; `std.As` needs no inference. `resource.Bracket[pgx.Tx, error]` with a `FromError(tx.Rollback(ctx))` release is the `defer tx.Rollback(ctx)` substitute.

**The wall this iteration found:** a first draft carried all eighteen `*Repository` methods in `repository.gala`; that makes the package's GALA index the receiver's whole method set for every importer, so `discord_bot/commands`, `slackbot/commands`, and `routes/group` were refused with a false `GALA-E0044` on methods their handwritten siblings declare (`ListActiveUserDays` among them). `verify.sh` caught it as five non-transpiling twins. The fix is the declarations-only split: no `.gala` file declares a `*Repository` method.

**Landed:** `server/postgres/repository.gala` -> generated `server/postgres/repository.go` (stdlib; declarations and free functions: `crockfordBase32`, `ErrPoolUninitialized`, `rowScanner`, `NewRepository`, `NewRepositoryFromTx`, `DefaultRepository`, `WithTransaction`, `GenerateShortID`, `GenerateEventShortID`, `GenerateTransferCode`, `formatTransferCode`, `encodeCrockford`, `eventColumns`, `scanEvent`, `listResponsesQuery`, `isUniqueViolation`, `IsUniqueViolation`), plus handwritten `server/postgres/repository_types.go` (`Repository`, `dbtx` verbatim) and `server/postgres/repository_methods.go` (all eighteen `*Repository` methods verbatim, extracted from the original). The two local random buffers became `go_interop.SliceWithSize[byte]`; `scanEvent` takes the named `rowScanner`; every Go-facing name, signature, tag, and behavior is unchanged, and the methods kept their `defer` and comments because they stayed handwritten.

**Evidence:** double transpile byte-identical, `gofmt -l` clean, generated imports only `go.gala.fyi/stdlib/...`; `server/scripts/gala/verify.sh` OK (24 twins) including `go build ./...`; the canonical Compose backend sequence green across every package (`postgres` 10.384s, `routes` 6.703s); the five previously broken importers transpile again.

**Ledger:** registry 23 -> 24 with the `postgres/repository` row; cursor entry 2 rewritten (remaining `postgres` candidates are the free-function files `codec.go`, `pool.go`, `tracing.go`; `*Repository` methods stay handwritten); three open-findings rows added (`*Repository` methods in a `.gala` file hide sibling methods from importers, anonymous interface type, unnamed interface parameters) and the fixed-array row extended with the local-buffer substitute; the `#615` report-index note records the unmerged imported-package direction; prose below the table records the split. `gala-from-go`'s `SKILL.md` Mixed Package section gained the one-receiver-one-language rule and `constructs.md` gained the interface/separator rows, the `std.As` spelling, and the packages-table row plus prose for the importer direction and the match inference.

**Next:** cursor entry 1 stays blocked; entry 2 offers `postgres/codec.go`, `postgres/pool.go`, and `postgres/tracing.go` as free-function candidates, with every `*Repository` method staying handwritten.

**State:** changes are in the worktree, not committed, at the time of this note; the pre-existing unrelated `backlog/backlog.md` modification was left untouched.

**Iteration 12 addendum — stale `.gala` cache (2026-10-10):** a later full `verify.sh` run reported `middleware/auth.go` drift (its transitive `models`/`postgres` standalone imports swapped order) after the v1 experiment's cache entries survived the `repository.gala` replacement; deleting `server/.gala/` and re-running regenerated the committed byte-identical file, and two further full runs report OK (24 twins). The ledger's contract section now records the remedy, and the committed `middleware/auth.go` is unchanged.

## Iteration 13 — postgres/codec whole-file twin (2026-10-10)

**Provenance:** `gala` `GALA version 0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c` and extraction marker match `server/GALA_COMPILER`, so no bump. The tree held only the pre-existing `backlog/backlog.md` edit, left untouched.

**Candidate:** cursor entry 1 (`models/set.go`, `models/location.go`, `models/event.go`) stays blocked: `scripts/gala/probes/models-struct-tag` and `models-empty-struct` re-ran on the pinned rev and reproduced their recorded parse errors. The loop took entry 2's first candidate, `postgres/codec.go`.

**Triage:** rewrite whole (no handwritten remainder). `emptyPayload`, `encodePayload`, and `decodePayload` are all expressible; the old `codec.go` was moved aside before the transpile.

**Re-derived rows (pinned compiler):** the declared `([]byte, error)` result list with a `Success`/`Failure` body emits the `std.Try` IIFE whose failure path returns `*new([]byte), err` (the original `nil, err`); `len(payload)` on `json.RawMessage` lowers from `.Size()`; `json.RawMessage(go_interop.ToBytes("{}"))` is a valid conversion; `go_interop.SliceCopy` accepts a named `[]byte` type and returns `[]byte`. One runtime subtlety was preserved deliberately: `SliceCopy` returns a non-nil empty slice for a non-nil empty input while `append(json.RawMessage(nil), payload...)` returns nil, and an empty non-nil `RawMessage` fails `json.Marshal` (verified with a host scratch program), so `decodePayload` guards the zero-length case and returns a zero-value `json.RawMessage`. `constructs.md`'s builtins prose now records the clone nuance.

**Landed:** `server/postgres/codec.gala` -> generated `server/postgres/codec.go`, a whole-file twin with no handwritten sibling. Every Go-facing name, signature, error string, and clone behavior is unchanged; `encodePayload`/`decodePayload` are still called by `repository_methods.go`, `calendar.go`, `dashboard.go`, and `repository.gala`.

**Evidence:**
- Double transpile byte-identical (`diff` clean) and `gofmt -l` clean; generated imports only `go.gala.fyi/stdlib/std` and `go.gala.fyi/stdlib/go_interop`; no `martianoff/gala`; `go vet ./postgres/` clean.
- `server/scripts/gala/verify.sh` first reported `middleware/auth.go` drift; deleting the stale `server/.gala/` cache (the ledger's recorded remedy) and re-running reports OK (25 twins) with `go build ./...` green and `middleware/auth.go` unchanged.
- Canonical Compose backend sequence green across every package (`postgres` 10.479s, `routes` 2.906s); existing `.env.test` kept.

**Ledger:** registry 24 -> 25 with the `postgres/codec` row; cursor entry 2 now offers only `pool.go` and `tracing.go`; postgres prose records the split. No finding changed.

**Next:** cursor entry 1 stays blocked; entry 2 is `postgres/pool.go` and `postgres/tracing.go`.

**State:** changes are in the worktree, not committed: `server/postgres/codec.gala`, `server/postgres/codec.go`, `server/GALA.md`, `.agents/skills/gala-from-go/references/constructs.md`, and this task's notes. The pre-existing unrelated `backlog/backlog.md` modification was left untouched.

## Iteration 14 — postgres/pool whole-file twin (2026-10-10)

**Provenance:** `gala` `GALA version 0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c` and extraction marker match `server/GALA_COMPILER`, so no bump. The tree held only the pre-existing `backlog/backlog.md` edit, left untouched.

**Candidate:** cursor entry 1 (`models/set.go`, `models/location.go`, `models/event.go`) stays blocked: `scripts/gala/probes/models-struct-tag` and `models-empty-struct` re-ran on the pinned rev and reproduced their recorded parse errors. The loop took entry 2's first candidate, `postgres/pool.go`.

**Triage:** rewrite whole (no handwritten remainder). The old `pool.go` was moved aside before the transpile.

**Re-derived rows (pinned compiler):** `defer cancel()` over a `(Context, CancelFunc)` pair is `resource.Bracket` with the bindings annotated; `resource.Bracket[context.CancelFunc, func()]` is a parse error because a function type cannot be written as an explicit type argument, so the release/body lambdas were inferred from `var closeFn func() = ...` (a named `type CloseFn func()` also passes as a type argument). The dot-imported `go_builtins.Panic` lowers to Go's builtin `panic` and emits no `go_builtins` import. `Panic` before and inside the Bracket body, a nested `return () => {...}` lambda, assignment to the package-level `Pool` from the body lambda, and `var createdPool, connErr = pgxpool.NewWithConfig(...)` all transpile; the emitted import block is `go.gala.fyi/stdlib/resource` alone.

**Behavior note:** `resource.Bracket` releases on every exit path but re-raises a caught string panic as a `*panicError`; nothing in the server recovers `Init`'s panics, so the process-visible behavior is unchanged, and the alternative (explicit `cancel()` before each panic) was rejected as a weaker substitute for the deferred release.

**Landed:** `server/postgres/pool.gala` -> generated `server/postgres/pool.go`, a whole-file twin with no handwritten sibling. The three environment-name `const`s became package `var`s (`pool_test.go` reads `maxConnectionsEnvironment` as a plain value), `Pool` stays an exported `var`, and every Go-facing signature and panic message is unchanged.

**Evidence:**
- Double transpile byte-identical (`diff` clean) and `gofmt -l` clean; generated imports only `go.gala.fyi/stdlib/resource`; no `martianoff/gala`; `go vet ./postgres/` clean.
- `server/scripts/gala/verify.sh` reports OK (26 twins) including `go build ./...`.
- Canonical Compose backend sequence green across every package (`postgres` 8.728s, `routes` 2.309s); existing `.env.test` kept; cache volumes created one per command.

**Ledger:** registry 25 -> 26 with the `postgres/pool` row; cursor entry 2 now offers only `tracing.go`; new open-findings row `Function type as an explicit type argument` (workaround, no probe needed); postgres prose records the landing and the `*panicError` note. `gala-from-go`'s resource trap, the function-type row, and the `panic` row record the type-argument refusal and the builtin-`panic` lowering. Formatted with `npm run format:markdown`.

**Next:** cursor entry 1 stays blocked; entry 2 is `postgres/tracing.go`.

**State:** changes are in the worktree, not committed: `server/postgres/pool.gala`, `server/postgres/pool.go`, `server/GALA.md`, the two `.agents` references, and this task's notes. The pre-existing unrelated `backlog/backlog.md` modification was left untouched.

## Iteration 15 — postgres/tracing whole-file twin (2026-10-10)

**Provenance:** `gala` `GALA version 0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c` and extraction marker match `server/GALA_COMPILER`, so no bump. The tree held only the pre-existing `backlog/backlog.md` edit, left untouched.

**Candidate:** cursor entry 1 (`models/set.go`, `models/location.go`, `models/event.go`) stays blocked (blockers last re-verified on this rev with the `models-*` probes); the loop took the postgres entry's remaining candidate, `tracing.go`.

**Triage:** rewrite whole. `postgresInstrumentationName` is read by `tracing_test.go`, so it is a package `var`; `queryTracer` is a `struct` whose `var tracer trace.Tracer` field keeps the test's keyed literal; its two value-receiver methods stay methods. Re-derived rows: a reassigned parameter is refused (`cannot assign to immutable variable`), so `newQueryTracer` copies it to a `var`; a declared Go result list cannot name its results (`func ConnectionStats() (used int, idle int, max int)` is a parse error) while the unnamed `(int, int, int)` lowers through `std.Tuple3`; `.Size()` on a Go slice and tuple returns `return (a, b, c)` work as recorded.

**Landed:** `server/postgres/tracing.gala` -> generated `server/postgres/tracing.go`, whole-file, no handwritten sibling. Signatures, span attributes, error strings, and status codes unchanged.

**Evidence:** double transpile byte-identical, `gofmt -l` clean, generated imports `go.gala.fyi/stdlib/std` and `go.gala.fyi/stdlib/go_interop`, no `martianoff/gala`; `server/scripts/gala/verify.sh` OK (27 twins) including a clean `go build ./...`; `go test ./postgres/ -run 'TestQueryOperation|TestQueryTracer|TestNewQueryTracer|TestConnectionStats'` passes; canonical Compose backend sequence green across every package (`postgres` 7.528s, `routes` 3.337s).

**Ledger:** registry 26 -> 27 with the `postgres/tracing` row; cursor entry 2 removed and the list renumbered (`main.go` became entry 2); prose records the landing. `gala-from-go` gained a parameter-reassignment row and the named-results note.

**Next:** cursor entry 1 stays blocked; entry 2 is `main.go`.

**State:** changes are in the worktree, not committed: `server/postgres/tracing.gala`, `server/postgres/tracing.go`, `server/GALA.md`, `.agents/skills/gala-from-go/references/constructs.md`, and this task's notes.

## Iteration 16 — main.go split (2026-10-10)

**Provenance:** pinned `gala`/flake rev as iteration 15; the tree held only the pre-existing `backlog/backlog.md` edit.

**Candidate:** cursor entry 2 `main.go`.

**Triage:** split. `main()` stays handwritten in the new sibling `server/main_run.go` because its local `chan error`, `go`, and `select` run loop has no substitute; `main.gala` carries `defaultLogPath` (the `const` row's `var`), `postgresPing` (a typed plain `var` because `main_test.go` assigns it), the swag general-info block, `init`, `initHealthRoute`, `shouldRunInReleaseMode`, `openLogFile`, `configureRuntime`, `validateSessionSecret`, `eventPageHandler`, and `noRouteHandler`.

**Re-derived rows:** a floating package-level comment is dropped, so the `@title`/`@version`/`@description`/`@host` block had to sit contiguously above `init`; all four annotations survive in generated `main.go` and `swag init --parseDependency` regenerated `server/docs` byte-identically afterwards; `len(secret)` became `secret.ByteSize()` to keep the 32 a byte count; the health handler's `defer cancel()` moved through `resource.Bracket` with an explicit `return true` body (a void body cannot infer `A`, and a trailing bare `c.JSON(...)` is auto-returned as `return c.JSON(...)`, which Go rejects); `func init()` and return-position typed closures (`= (c) => {...}`) work as recorded.

**Landed:** `server/main.gala` -> generated `server/main.go`, plus handwritten `server/main_run.go` (`main()` verbatim); `main_test.go` passes unchanged.

**Evidence:** double transpile byte-identical, `gofmt -l` clean on both files, `go build ./...`, `go vet .`, `go test .` green; `swag init --parseDependency` (v1.16.6) regenerated `server/docs` byte-identically; `verify.sh` OK (28 twins); canonical Compose sequence green across every package.

**Ledger:** registry 27 -> 28 (sixteen packages) with the `main` row; cursor entry 2 removed and renumbered; prose records the split, the swag comment requirement, and the Bracket-body rule. `constructs.md` gained the `Bracket` void-body note.

**Next:** cursor entry 2 is the observability group (`provider.go`, `readiness.go`, `transport.go`).

**State:** changes are in the worktree, not committed: `server/main.gala`, `server/main.go`, `server/main_run.go`, `server/GALA.md`, `.agents/skills/gala-from-go/references/constructs.md`, and this task's notes.

## Iteration 17 — observability/provider split (2026-10-10)

**Provenance:** pinned `gala`/flake rev as iteration 15; the tree held only the pre-existing `backlog/backlog.md` edit.

**Candidate:** cursor entry 2's first file, `observability/provider.go`.

**Triage:** split. `Recorder` and its ten provider methods stay handwritten in the new sibling `server/observability/provider_recorder.go` because the receiver's methods are also declared in `metrics.go` and `transport.go` and the receiver has to stay in one language; `provider.gala` carries the const block as package vars, the package comment, `Start`, `newRecorder`, `newResource`, and `ReportExportError`. Re-derived rows: exporter header maps use `go_interop.MapEmpty`/`MapPut` with both type arguments; the shared resource uses `SliceOf`/`SliceAppend` and a spread; a Go brace-literal key that resolves to a local binding is rewritten with the binding's read, so the SDK provider locals must be `var`s (the `val` emission was `&Recorder{logProvider.Get(): ...}`, which fails `go build`).

**Landed:** `server/observability/provider.gala` -> generated `server/observability/provider.go` (package comment emitted), plus `provider_recorder.go`. `Start`'s and `newRecorder`'s signatures, error strings, exporter options, and the `DefaultShutdownTimeout` value are unchanged; `DefaultShutdownTimeout` is a `var` because `const` has no GALA spelling and nothing reads it in a constant expression.

**Evidence:** double transpile byte-identical, `gofmt -l` clean on both files, `go build ./...`, `go vet ./observability/`, `go test ./observability/` green; `verify.sh` OK (29 twins); canonical Compose sequence green (`observability` 0.330s, `postgres` 7.097s, `routes` 3.415s).

**Ledger:** registry 28 -> 29 with the `observability/provider` row; cursor entry 2 rewritten to `readiness.go` then `transport.go` with the `*Recorder` method boundary; the receiver-methods finding row records the observability instance; prose records the split and the brace-key trap.

**Next:** cursor entry 2 is `observability/readiness.go` then `observability/transport.go`.

**State:** changes are in the worktree, not committed: `server/observability/provider.gala`, `provider.go`, `provider_recorder.go`, `server/GALA.md`, `.agents/skills/gala-from-go/references/constructs.md`, and this task's notes.

## Iteration 16 — eight-twin batch (2026-10-25)

**Provenance:** `gala` `GALA version 0.93.0` at `/nix/store/v0jxpys9mxb58dvdp3np8bywpchbyppa-gala-0.93.0/bin/gala`; flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c` matches `server/GALA_COMPILER` (extraction `0.93.0 6368c093…`), so no bump. The user asked for a faster batch because a full `verify.sh` run is the expensive step, so one verifier run covers eight independent candidates instead of one per candidate; each twin still keeps its Go-facing surface and the batch is committed as one change.

**Landed (8 twins):**
- `observability/transport.gala` -> `transport.go`; sibling `transport_extra.go` keeps `(*Recorder).Transport`; `defer span.End()` moves through `resource.Bracket`.
- `services/auth/auth.gala` -> `auth.go`; sibling `auth_extra.go` keeps the tagged `GoogleIdTokenInfo`, `RefreshAccessTokenData`, and the `chan`/`recover` refresh members.
- `services/calendar/types.gala` -> `types.go` (runtime-free; `match` with a block per arm and a final `case _`).
- `services/contacts/contacts.gala` -> `contacts.gala` output `contacts.go`; `contacts_extra.go` holds the tagged People API structs.
- `services/listmonk/listmonk.gala` -> `listmonk.go`; `listmonk_extra.go` holds the tagged subscriber-search structs.
- `services/microsoftgraph/microsoftgraph.gala` -> `microsoftgraph.go` (runtime-free); `microsoftgraph_extra.go` holds `UserInfo` and the Graph response.
- `services/gcloud/tasks.gala` -> `tasks.go` (nested protobuf literals, `MapEmpty[int, *timestamppb.Timestamp]()`, map-index assignment).
- `mockprovider/main.gala` -> `main.go` (constants to vars, `MapEmpty`/`MapPut`/`SliceOf`, `() => {}` closuress).

**Construct facts re-derived on the pinned compiler:** a `Bracket` release over `Close() error` needs `FromError(r.Body.Close())`; an error result list takes `Success`/`Failure` while a concrete pointer error lowers from a tuple; empty named map literals (`http.Header{}`, `url.Values{}`) are direct; a match arm with `return` must be a block and a non-sealed match needs `case _`; `.Size()` on a string is a rune count so `ByteSize()` was used for the byte check in `contacts`; keyed literals of sibling-declared types build.

**Evidence:** `server/scripts/gala/verify.sh` -> OK (38 twins), 38/38 byte-identical double transpiles, gofmt clean, no `martianoff/gala`; `go build ./...` and `go vet ./...` green; canonical Compose sequence `server-route-test` green across every package (`routes` 9.155s, `postgres` 10.004s, `services/auth` 0.173s, `observability` 0.164s, `services/gcloud` 0.046s, `microsoftgraph` 0.014s).

**Ledger:** registry 30 -> 38 sources and 16 -> 23 packages; cursor rewritten to the remaining models group and the `services/calendar` provider files; batch prose records the construct facts. `constructs.md` gained the match-arm-block, empty-named-map-literal, and Bracket-release notes.

**Caveat:** `.gitignore`'s `listmonk/` pattern ignores new files under `server/services/listmonk/`, so `listmonk.gala` and `listmonk_extra.go` were force-added; the tracked `listmonk.go` is unaffected.

**Next:** cursor entry 1 stays the models trio; entry 2 is `services/calendar/ics_calendar.go`, then the other provider files; `services/calendar/calendar.go` stays handwritten on the channel/recover findings.

**State:** committed as the batch commit on `rewrite-in-gala`; the pre-existing unrelated edits were left untouched.

## Iteration 18 — ten-twin batch (2026-10-25)

**Provenance:** `gala` `GALA version 0.93.0` at `/nix/store/v0jxpys9mxb58dvdp3np8bywpchbyppa-gala-0.93.0/bin/gala`; flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c` matches `server/GALA_COMPILER` (extraction `0.93.0 6368c093…`), so no bump. The user asked for ten more Go files and to run `verify.sh` once after translating, because a per-twin verifier run is the expensive step, so one run covers the ten independent candidates; each twin still keeps its Go-facing surface and the batch is committed as one change.

**Landed (10 twins):**
- `accounts/users.gala` -> `users.go` (whole file: `models.User` keyed literal and `models.UUID(...)` conversion, `(*pgstore.Account, error)` helpers).
- `respondents/identity.gala` -> `identity.go`; `identity_extra.go` keeps the validation constants, `GuestNameValidationCode`, and `GuestNameValidationResult` so their Go surface is unchanged.
- `observability/config.gala` -> `config.go`; `config_extra.go` keeps `Config`.
- `observability/record.gala` -> `record.go`; `record_extra.go` keeps `RequestCompletion`.
- `postgres/credentials.gala` -> `credentials.go` (whole file: AES-256-GCM codec, `SliceWithSize`/`SliceAppendAll`/`SliceTake`/`SliceDrop`).
- `services/calendar/ics_calendar.gala`, `google_calendar.gala`, `outlook_calendar.gala`, `apple_calendar.gala` -> their generated files; each sibling keeps the provider struct (embedded auth field) and the named response structs that the anonymous-struct row forces handwritten.
- `slackbot/slackbot.gala` -> `slackbot.go` (runtime-free); `slackbot_extra.go` keeps `SlackbotMessageType`, `GENERAL`, and the swag-annotated `execCommand`.

**Construct facts re-derived on the pinned compiler:** a `.gala` file that declares a method on a sibling-declared type gives that file the receiver but not the field list, so constructing the same type there is refused with `GALA-E0045` (`accounts/delete.go` stayed handwritten on that fact, and the workaround row is in the ledger and `constructs.md`); an exported GALA struct emits `Copy`/`Equal`/`Unapply`/`New<Type>`/`StructMeta_<Type>`, so an exported struct whose synthesized members would widen the API stays in a sibling (`Config`, `RequestCompletion`); a three-result `(T, T, error)` GALA function is written `Failure(err)`/`Success((a, b))` and bound with `var a, b, err = f()` (`apple_calendar.gala`'s `getClients`); a method body reads an embedded struct's promoted field through a sibling receiver (`cal.AccessToken`, `cal.FeedURL`); `parseTimeWithTZ` keeps the original's shadowed `err` in the `TZID` branch deliberately, so a parse failure there still returns the zero time; the standalone runtime import spelling keeps `gofmt` order when `golang.org/x/...` and `go.gala.fyi/...` share a block.

**Evidence:** `server/scripts/gala/verify.sh` -> OK (48 twins) in 6m08s, 48/48 byte-identical double transpiles, `gofmt` 48 checked, orphan scan over 196 Go files, `go build ./...` ok; `go vet ./...` clean; the canonical isolated Compose sequence from `server/README.md` passes across every package (`postgres` 6.979s, `routes` 2.722s, `accounts` 0.215s, `observability` 0.240s, `respondents` 0.008s, `services/auth` 0.059s, `services/gcloud` 0.050s, `services/microsoftgraph` 0.015s, `utils` 0.065s). No generated file hand-edited and no test relaxed.

**Ledger:** registry 38 -> 48 sources and 23 -> 26 packages (new `accounts`, `respondents`, `slackbot` package rows); the cursor's `services/calendar` provider entry is gone and now lists `accounts/accounts.go` then `accounts/delete.go`, `utils/utils.go`, `slackbot/commands/utils.go` then `discord_bot/commands/index.go`, and the blocked `services/calendar/calendar.go`; the open-findings table gained the `.gala`-method-on-a-sibling row; a ten-twin batch prose section records the construct facts; `constructs.md` gained the sibling-method construction row and its prose. `npm run format:markdown` was run on `server/GALA.md`.

**Next:** cursor entry 2 is `accounts/accounts.go`, then `accounts/delete.go` with its construction boundary.

**State:** committed as the ten-twin batch commit on `rewrite-in-gala`.

## Iteration 19 — third ten-twin batch (2026-10-25)

**Provenance:** `gala` `GALA version 0.93.0` at `/nix/store/v0jxpys9mxb58dvdp3np8bywpchbyppa-gala-0.93.0/bin/gala`; flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c` matches `server/GALA_COMPILER` (extraction `0.93.0 6368c093…`), so no bump. The user asked for ten more Go files, so one `verify.sh` run covers the ten independent candidates and the batch is committed as one change.

**Landed (10 twins):**
- `accounts/accounts.gala` -> `accounts.go`; `accounts_extra.go` keeps `Profile`.
- `accounts/calendar.gala` -> `calendar.go`; `calendar_extra.go` keeps `CalendarIntegrations`, `CalendarPreferences`, and `newCalendarIntegrations`.
- `accounts/delete.gala` -> `delete.go` (runtime-free); `delete_extra.go` keeps `Deleter`, `DefaultDeleter`, its `Delete` method, and `defaultDeleter` in one language.
- `utils/utils.gala` -> `utils.go`; `utils_extra.go` keeps `GetAuthUser`, `GetAuthAccount`, and `CORSOrigins`.
- `slackbot/commands/utils.gala` -> `utils.go`; `utils_extra.go` keeps `Response`, `Command`, `CommandMap`, and `newResponse`.
- `discord_bot/commands/index.gala` -> `index.go`; `index_extra.go` keeps `Command`, `Commands`, and `Init`.
- `routes/events.gala` -> `events.go`; `events_extra.go` keeps `normalizeTimedResponseAvailabilitySlots`.
- `routes/grants.gala` -> `grants.go`; `grants_extra.go` keeps the named body struct, its constructor, and the session helper.
- `routes/folders.gala` -> `folders.go`; `folders_extra.go` keeps the tagged response structs, the named bodies, the identity helper, and the response builders.
- `observability/middleware.gala` -> `middleware.go`; `middleware_extra.go` keeps `errorCapture` and its methods, `newErrorCapture`, `newRequestCompletion`, and `maxCapturedErrorBytes`.

**Construct facts re-derived on the pinned compiler:** GALA keywords are not identifiers (`val` in a `range` binding and `embed` as a parameter are parse errors); a non-error result list takes a plain tuple return while `Success` there emits a `std.Try` the emitted `std.Tuple` body cannot assign; `Success((nil, nil))` infers a three-result list's tuple types; `c.Errors.Size()` does not resolve on gin's defined `Errors` slice, so `go_interop.SliceFrom(c.Errors, 0).Size()` is the count; `gin.H{...}` payloads became `MapPut`/`MapEmpty` maps; a void function's `defer resp.Body.Close()` is `resource.Bracket(..., (_) => { return true })` in statement position; the slicing loop moved to `SliceTake`/`Slice`/`SliceFrom` over `ToBytes` with a `var remaining` copy; `[16]byte` became `SliceWithSize[byte](16)`; the `go_interop` import had to stand alone in `routes/events.gala` for `gofmt` order.

**Regression found and worked around:** `accounts.Resolve`'s `*pgstore.Account` leaked the alias qualifier into `middleware/auth.gala`'s regeneration without its import, and the drift survived a cache clear. The three `accounts` twins moved to the `"timeful/server/postgres"` package-name spelling, which made `middleware/auth.gala` regenerate byte-identically again; the ledger's open findings table and `constructs.md` gained the row, and `accounts/users.go` was re-baselined onto the same spelling.

**Evidence:** `server/scripts/gala/verify.sh` -> OK (58 twins) in 8m02s (58/58 byte-identical, `gofmt` 58 checked, orphan scan over 206 Go files, `go build ./...` ok); `go vet ./...` clean; the canonical isolated Compose sequence passes (`routes` 2.703s, `postgres` 8.607s, `accounts` 0.205s, `observability` 0.235s, `utils` 0.012s); `swag init --parseDependency` regenerated `server/docs` byte-identically; `npm run format:markdown` and `:check` pass.

**Ledger:** registry 48 -> 58 sources (packages stay 26), with the ten rows and corrected `slackbot/commands` sibling names; a third batch prose section records the construct facts; the `.gala`-method-on-a-sibling row now records the `delete.gala` split; the aliased-qualifier finding row is new; the cursor drops the three landed entries and lists `routes/identity.go`/`routes/owner.go`, `routes/analytics.go`/`routes/event_timed_slots.go`, and `observability/metrics.go`.

**Next:** cursor entry 2 is `routes/identity.go`, then `routes/owner.go`.

**State:** committed as the third ten-twin batch commit on `rewrite-in-gala`; no generated file hand-edited and no test relaxed.

## Iteration 21 — sixth batch: routes/user and the postgres repository reads (2026-10-26)

**Provenance:** `gala` `GALA version 0.93.0` at `/nix/store/v0jxpys9mxb58dvdp3np8bywpchbyppa-gala-0.93.0/bin/gala`; flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c` matches `server/GALA_COMPILER` (extraction `0.93.0 6368c093…`), so no bump. The user asked for more server Go files translated with parallel sub-agents, one `verify.sh` run, then a fixer agent; the harness terminated several sub-agents mid-run, so the integrator reconciled their uncommitted drafts, completed the split boundaries, and ran the verification itself.

**Landed (6 twins):**
- `routes/user.gala` -> `user.go`; `user_extra.go` keeps the tagged request and query bodies and their constructors plus `addCalendarAccountArgs` and its constructor; the swag blocks survive in the generated file.
- `postgres/accounts.gala` -> `accounts.go`; `accounts_extra.go` keeps `Account`, `newAccount`, `createPlatformIdentity`, `ListAccountsByPlatformIdentityIDs`, and one-line exported method wrappers over the twin's free-function bodies.
- `postgres/signup.gala` -> `signup.go`; `signup_extra.go` keeps `SignupBlock`, `SignupResponse`, and every `*Repository` method.
- `postgres/calendar.gala` -> `calendar.go`; `calendar_extra.go` keeps the six `Calendar*` structs, `encodedCalendarSecrets`, `(*CalendarAccount).requiresEncryption`, and every `*Repository` method.
- `postgres/folders.gala` -> `folders.go`; `folders_extra.go` keeps `Folder`, `FolderMember`, and the six methods.
- `postgres/attendees.gala` -> `attendees.go`; `attendees_extra.go` keeps `Attendee` and the six methods.

**Construct facts on the pinned compiler:**
- A `.gala` method on a Go-declared receiver becomes that receiver's complete GALA method set and hides the package's sibling methods: `folders.gala`'s six declarations were refused with `GALA-E0044: Repository has no method withTransaction ... Repository declares: <the six>`. Every `*Repository` method therefore stays in the sibling, and `accounts` moves the bodies into free functions behind thin exported wrappers.
- `go_interop.SliceEmpty[T]()` returns nil on stdlib v0.93.0, so `go_interop.SliceWithSize[T](0)` is the faithful non-nil `[]T{}` wherever nil-ness is observable (folders' `Folder.Members` serializes as JSON).
- A declared Go result list cannot return a non-nil value together with a non-nil error, so `ListAccountsByPlatformIdentityIDs` stays handwritten.
- Zero-value construction of a sibling-declared struct works from `.gala`; a keyed construction works too when no `.gala` method is declared on that type.
- A raw `*T` local cannot call a method declared on `Ptr[T]` (`GALA-E0080`); `RequirePtr` narrows it after a nil guard, and `Success(nil)` fills a declared `(*string, error)` empty branch.
- A named interface (`repository.gala`'s `rowScanner`) crosses `.gala` files; `resource.Bracket` covers `defer rows.Close()`; `go_interop.OptionFromMap` covers a comma-ok lookup.

**Evidence:** `server/scripts/gala/verify.sh` -> `OK (74 twins)` in 8m36s (74/74 byte-identical double transpiles, `gofmt` 74 checked, orphan scan 222 Go files, `go build ./...` ok); `go vet ./...` clean; the canonical isolated backend test sequence's `postgres-test-migrate` and `server-route-test` images passed every package against `timeful-test-postgres` (15 npm-style `ok` lines, `postgres` 13.350s, `routes` 4.595s; the Docker bridge could not route between this session's test containers, so the same images ran with host networking against the isolated database); `swag init --parseDependency` regenerated `server/docs` byte-identically; `npm run format:markdown:check` passes, while `npm run lint:markdown` still reports the pre-existing `docs/environments.md:95` sentence-per-line issue (file untouched by this batch).

**Ledger:** registry 68 -> 74 sources (26 packages); sixth-batch section; two new findings rows (`SliceEmpty` nil slice, non-nil value with error); the `*Repository` row records the 2026-10-26 refusal; the cursor drops `routes/user` and lists the remaining postgres repository reads; `constructs.md` gained the receiver-method-set row, the `SliceEmpty` refinement, the sibling-construction note, and the `RequirePtr` note.

**State:** committed as the sixth-batch commit; no generated file hand-edited, no test relaxed, no probe warranted.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: TASK-0350
created: 2026-10-05 09:37
---
TASK-0350 resolved 2026-10-05 with no upstream report: the `opaque type` runtime binding is suppressible. A same-package `.go` sibling declaring `Hash`/`Compare` keeps the generated Go free of `martianoff/gala` (verified on the pinned rev), at the cost of two added exported methods. `server/GALA.md` cursor entry 1 and the defined-scalar finding row now record that path, and the probe notes carry the repro; upstream PR #665 documents the suppression, so no comment on #528/#621 was needed. Struct tags, fixed-size arrays, and `struct{}` stay handwritten under `gaps.md`'s "What Not To File".
---

author: opencode
created: 2026-10-10 16:46
---
TASK-0356 refreshed acceptance criterion #2 to the adopted `go.gala.fyi/stdlib` contract. TASK-0353 superseded the runtime-free requirement, and iterations 12–14 already landed stdlib twins under this task, so the criterion now allows the pinned runtime where a runtime construct keeps the Go-facing surface, silently-evaluated side of the old rule, and keeps `martianoff/gala` banned. The description no longer says the iterations land runtime-free twins; the title stays as its historical label because the refreshed criterion is the operative contract.
---

author: opencode
created: 2026-10-10 16:47
---
Correction to the previous note: its middle clause is garbled. It should read: the refreshed criterion allows the pinned `go.gala.fyi/stdlib` runtime where a runtime construct keeps the Go-facing surface, allows a runtime-free form where a runtime construct would change that surface, and keeps `martianoff/gala` banned.
---
<!-- COMMENTS:END -->
