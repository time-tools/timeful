---
id: TASK-0347
title: >-
  Bump the GALA compiler pin to upstream HEAD and remove the vendored standard
  library
status: Done
assignee:
  - '@opencode'
created_date: '2026-10-05 04:12'
updated_date: '2026-10-05 06:30'
labels:
  - gala
  - tooling
dependencies: []
references:
  - flake.lock
  - server/GALA.md
  - server/scripts/gala/verify.sh
  - .agents/skills/gala-loop/SKILL.md
  - 'https://github.com/martianoff/gala/blob/master/docs/GO_INTEROP.MD'
  - 'https://github.com/martianoff/gala/issues/698'
documentation:
  - server/GALA.md
  - server/README.md
  - .agents/skills/gala-loop/SKILL.md
modified_files:
  - flake.lock
  - server/GALA.md
  - server/GALA_COMPILER
  - server/README.md
  - .agents/skills/gala-loop/SKILL.md
  - server/scripts/gala/verify.sh
  - server/go.mod
  - server/Dockerfile
  - server/middleware/auth.gala
  - server/middleware/auth.go
  - server/middleware/auth_session.go
  - server/postgres/dailylogs.gala
  - server/postgres/dailylogs.go
  - server/postgres/dailylogs_methods.go
  - server/postgres/dailylogs_types.go
  - server/discord_bot/init.gala
  - server/discord_bot/init.go
  - server/discord_bot/interop.go
  - server/third_party/gala/
priority: medium
type: task
ordinal: 351005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Upstream HEAD moved past the locked `d38fadb8` (GO_INTEROP.MD, #703) to `ddbe839b` with compiler fixes #699, #705, #706, #707, #708, #709, #710 plus the #711 test timeout. The user also wants the vendored GALA runtime gone if GO_INTEROP.MD's mixed-package guidance makes that possible.

Facts that frame the work: `server/third_party/gala/` exists only because three generated twins import `martianoff/gala` — `middleware/auth.go` (`std.As` from a typed `match` on `any`), `postgres/dailylogs.go` (`std.Copy`/`Equal`/`Unapply`/`StructMeta_*` emitted for GALA `struct` declarations), and `discord_bot/init.go` (`go_interop.MapEmpty`, `go_interop.SliceFrom`). The other twelve twins are already runtime-free and import no `martianoff/gala` package. `server/go.mod` requires `martianoff/gala v0.0.0` and replaces it with the local tree; `server/Dockerfile` copies the vendored `go.mod` before `go mod download`. Upstream's module path `martianoff/gala` has no dot in its first element, so it is not fetchable as a normal Go module; the GO_INTEROP.MD route is Part 3 mixed packages — keep the Go-facing shape and move what GALA cannot express into a handwritten Go sibling in the same package.

Constraints: no hand-edited generated twin; no Go-facing signature, struct shape, or behavior change; no relaxed test; transpile from the package directory; bump and de-vendor have separate evidence because the bump re-baselines every twin against a new compiler; `server/GALA.md` and the `gala-loop` skill must state the new contract so the next iteration does not re-vendor.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The `gala` input in `flake.lock` is bumped to current upstream master (`ddbe839b` at filing, or a recorded newer rev), `nix develop` still builds `gala-local`, and the dev shell runs the pinned compiler
- [x] #2 No generated twin imports `martianoff/gala`: after deletion, `grep -rn '"martianoff/gala' server --include='*.go'` returns nothing
- [x] #3 The three runtime-enabled twins are rewritten runtime-free per GO_INTEROP.MD Part 3: `middleware/auth.gala` calls a handwritten session helper (no `std.As`), `postgres/dailylogs.gala` drops its struct declarations to a handwritten sibling, and `discord_bot/init.gala` calls handwritten map/slice helpers (no `go_interop`), with no Go-facing shape or behavior change
- [x] #4 `server/third_party/gala/` is deleted, `server/go.mod` no longer requires or replaces `martianoff/gala`, `server/Dockerfile` no longer copies the vendored `go.mod`, and no other build, deploy, or CI reference remains
- [x] #5 `server/scripts/gala/verify.sh` no longer requires a vendored runtime or its extraction marker but still fails on missing provenance, drift, gofmt, orphans, or a failing `go build ./...`, and records the compiler pin from `flake.lock`
- [x] #6 `server/GALA.md`, `server/README.md`, and the `gala-loop` skill state the new contract: flake-pinned compiler, runtime-free twins only, no vendored runtime; the twin registry rows for the three rewritten twins are updated
- [x] #7 `nix develop --command server/scripts/gala/verify.sh --write` regenerates every twin against the bumped compiler, the per-twin diff is reviewed, and every shape change is recorded with no regression papered over
- [x] #8 The canonical backend test sequence in `server/README.md` passes with the vendor removed, and the task records provenance, per-twin diff, verify.sh, build, and test evidence
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
1. Record provenance on the current lock (`gala version`, `which gala`, `jq -r '.nodes.gala.locked.rev' flake.lock`, `cat server/third_party/gala/VENDORED_FROM`) and capture a green baseline `verify.sh`.
2. De-vendor iteration on the old pinned compiler: add `middleware/auth_session.go` with the session identity type assertion and rewrite `middleware/auth.gala` to call it; move `DailyUserLog`/`DailyUserLogMember` to a handwritten `postgres/dailylogs_types.go` and drop them from `dailylogs.gala`; add `discord_bot/interop.go` with `newCommandMap`/`argsFrom` and rewrite `init.gala` to call them. Transpile each from its package directory, confirm no `martianoff/gala` import, and update the registry rows.
3. Remove the vendor: delete `server/third_party/gala/`, drop the require/replace from `server/go.mod`, drop the vendored `COPY` from `server/Dockerfile`, and update `verify.sh` to read a slim `server/GALA_COMPILER` provenance file instead of `VENDORED_FROM`.
4. Update the contract docs: `server/GALA.md` compiler section and re-vendor procedure become a bump + re-baseline procedure with the runtime-free-only policy; update the `server/README.md` provenance line and the `gala-loop` skill preconditions and Step 0.
5. Bump iteration: `nix flake update gala`, verify `nix develop --command gala version` and `gala-local`, refresh the extraction marker, record the new fingerprint in `GALA_COMPILER`, run `verify.sh --write`, and review the per-twin diff.
6. Run `go build ./...`, the canonical backend test sequence, `npm run format:markdown` checks, root `fmt:check`, and refresh the code graph; record all evidence in the task and close it.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Provenance (before, 2026-10-05)
- `gala version` = `GALA version 0.85.0`; `which gala` = `/nix/store/5jy1pg6q86i36bldfydbj8mwwy9hxzlb-gala-0.85.0/bin/gala`; `jq -r '.nodes.gala.locked.rev' flake.lock` = `d38fadb844f0849c6c6ac8a47e8f91e025eace6b`; VENDORED_FROM extraction = `0.85.0 8206b230...`. Baseline `verify.sh`: OK (15 twins).
- Runtime-import census before: only `middleware/auth.go` (`std.As`), `postgres/dailylogs.go` (`std.Copy`/`Equal`/`Unapply`/`StructMeta_*`), `discord_bot/init.go` (`go_interop.MapEmpty`/`SliceFrom`). Upstream's module path `martianoff/gala` is not fetchable (no dot in the first element), so GO_INTEROP.MD Part 3 mixed packages is the route to remove the vendor.

## De-vendor iteration (old compiler)
- `middleware/auth_session.go` adds `sessionIdentityID`, the exact `sessions.Default(c).Get("userId").(string)` assertion from the pre-translation Go; `auth.gala` calls it and drops the `sessions` import and the type-pattern match.
- `postgres/dailylogs_types.go` declares `DailyUserLog` and `DailyUserLogMember` with the original doc comments; `dailylogs.gala` keeps `dailyLogDate` and the two query constants; `dailylogs_methods.go`'s header comment updated. Nothing referenced the removed `Copy`/`Equal`/`Unapply`/`StructMeta_*`.
- `discord_bot/interop.go` adds `newCommandMap` (`map[string]commands.Command{}`) and `argsFrom` (`args[index:]`), restoring the pre-translation Go expressions; `init.gala` drops the `go_interop` import.
- Each regenerated from its package directory: no `martianoff/gala` import remains; `gofmt -l` clean; `go build ./...` green; `verify.sh` OK (15 twins).

## Vendor removal
- `server/third_party/gala/` moved aside to `/tmp/opencode/gala-vendor-backup/gala` (2.0 MB, 118 tracked files deleted in git status).
- `server/go.mod`: dropped `martianoff/gala v0.0.0` and `replace martianoff/gala => ./third_party/gala`; `go mod tidy` made no further changes; `go build ./...` green. `server/Dockerfile`: dropped the `COPY third_party/gala/go.mod third_party/gala/` line and its comment.
- Repo-wide grep for `third_party/gala|VENDORED_FROM|"martianoff/gala` leaves only the flake input and the docs that describe the new contract.

## Bump iteration
- `nix flake update gala` hit anonymous GitHub API rate limits (rotating egress IPs), so the rev was fetched through the authenticated `gh` session with `NIX_CONFIG="access-tokens = github.com=$(gh auth token)"`; the token was never printed and the lock kept the `github:martianoff/gala` input type.
- Locked rev `ddbe839b5d038ef3445e1dac93f09e624144012f` (2026-10-05), narHash `sha256-Fwc2T+dCE0i5evIWZL6/f30s7UHx0tFsGlhKE3Aqyok=`. Dev shell rebuilt: `which gala` = `/nix/store/bwmk5ivbn7hq4pj3k3xc4zvar1blbyvq-gala-0.85.0/bin/gala`, so `gala-local` still builds.
- Extraction refresh: deleted `~/.gala/stdlib/v0.85.0/.stdlib-extracted`, transpiled `server/eventid/eventid.gala` with the new compiler; marker came back identical, `0.85.0 8206b230308f8892c8c41dbb1fdf034c2010acc5b120f7f766c29edcfe9faebd` (embedded stdlib snapshot unchanged between the two revs).
- `server/GALA_COMPILER` written with version, rev, extraction, and date; it replaces `VENDORED_FROM` as the verify provenance.
- `verify.sh` (check) found zero drift; `verify.sh --write` printed no WRITE lines, so no twin was re-baselined and no codegen shape changed on the new compiler.

## Tooling and docs
- `verify.sh`: `PROVENANCE_FILE="$SERVER_DIR/GALA_COMPILER"`; messages say "bump iteration"; new RUNTIME check per twin; `third_party` find excludes removed. Negative test: a throwaway twin with a GALA struct emitted `RUNTIME: ... keep the twin runtime-free (GO_INTEROP.MD Part 3)` and FAILED, then was deleted.
- `server/GALA.md`: contract is now runtime-free-only with the sibling rule and GO_INTEROP.MD Part 3 link; "Compiler and runtime" became "Compiler" with the bump + re-baseline procedure; registry lists the three rewritten twins as runtime-free with their new siblings; findings and stop conditions updated.
- `server/README.md` and `.agents/skills/gala-loop/SKILL.md` (preconditions, Step 0, stop conditions) updated from re-vendor to bump + re-baseline.

## Verification (2026-10-05)
- `nix develop --command server/scripts/gala/verify.sh`: OK (15 twins), provenance matching.
- Canonical backend sequence: `docker volume create` per volume, `docker compose ... up -d postgres-test postgres-test-bootstrap postgres-test-migrate`, then `run --rm server-route-test` exit 0 with 15 packages `ok` and no `FAIL` (`/tmp/opencode/task-0347/backend-tests.log`). The test image build exercises the simplified `server/Dockerfile`.
- `npm run format:markdown` formatted `server/GALA.md`; `npm run format:markdown:check` clean; root `npm run fmt:check` clean; `bash -n server/scripts/gala/verify.sh` clean.
- `codebase-memory-mcp` index refreshed: 8906 nodes, 28213 edges.
- e2e not run: no frontend/browser surface changed. Swagger not regenerated: no route annotations changed.
- Worktree also carries the pre-existing unrelated `backlog/backlog.md` edit, left untouched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Bumped the GALA flake pin from `d38fadb8` to upstream HEAD `ddbe839b` and removed the vendored GALA standard library by making every twin runtime-free through GO_INTEROP.MD Part 3 mixed-package siblings.

Delivered:
- `flake.lock` gala locked to `ddbe839b5d038ef3445e1dac93f09e624144012f` (github input type preserved); dev-shell compiler is `/nix/store/bwmk5ivbn7hq4pj3k3xc4zvar1blbyvq-gala-0.85.0/bin/gala`. The rev carries #699, #705, #706, #707, #708, #709, #710 plus the #711 test timeout on top of GO_INTEROP.MD.
- Runtime-free rewrites: `middleware/auth_session.go` supplies the session type assertion so `auth.gala` no longer lowers to `std.As`; `postgres/dailylogs_types.go` holds `DailyUserLog`/`DailyUserLogMember` so `dailylogs.gala` no longer emits `Copy`/`Equal`/`Unapply`/struct metadata; `discord_bot/interop.go` supplies `newCommandMap`/`argsFrom` so `init.gala` no longer imports `go_interop`. Go-facing shapes and behavior are unchanged.
- `server/third_party/gala/` (2.0 MB, 118 files) removed; `server/go.mod` dropped the `martianoff/gala` require and replace (`go mod tidy` no-op); `server/Dockerfile` dropped the vendored `go.mod` copy.
- `server/scripts/gala/verify.sh` now reads `server/GALA_COMPILER` for provenance, keeps the flake-rev and extraction-marker checks, and gains a RUNTIME guard that fails any twin whose generated Go imports `martianoff/gala`.
- Contract docs updated: `server/GALA.md` (runtime-free-only contract, compiler section and bump procedure, registry rows), `server/README.md`, and the `gala-loop` skill (bump instead of re-vendor).

Evidence:
- `verify.sh`: OK (15 twins) before the bump, after the rewrites, and after the bump; the extraction fingerprint stayed `0.85.0 8206b230...`, so the compiler rev moved while the embedded stdlib snapshot did not. `verify.sh --write` printed no WRITE/DRIFT lines.
- Negative guard check: a throwaway twin declaring a GALA struct produced `RUNTIME: ... imports the GALA runtime ... keep the twin runtime-free (GO_INTEROP.MD Part 3)` and a FAILED status; the probe was removed.
- Canonical backend Compose suite: exit 0, 15 packages `ok`, 0 `FAIL` (log `/tmp/opencode/task-0347/backend-tests.log`).
- `npm run format:markdown` + `:check` clean; root `npm run fmt:check` clean; code graph refreshed (8906 nodes, 28213 edges).
- e2e not run: no frontend surface changed. Swagger not regenerated: no route annotations changed. Not committed: the user did not ask for a commit.
<!-- SECTION:FINAL_SUMMARY:END -->
