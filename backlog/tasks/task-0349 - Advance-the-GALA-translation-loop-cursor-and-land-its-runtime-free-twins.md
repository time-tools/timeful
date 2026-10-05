---
id: TASK-0349
title: Advance the GALA translation loop cursor and land its runtime-free twins
status: In Progress
assignee: []
created_date: '2026-10-05 08:44'
updated_date: '2026-10-05 08:45'
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
priority: medium
type: enhancement
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
<!-- SECTION:NOTES:END -->
