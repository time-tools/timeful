---
id: TASK-0353
title: Decide whether server GALA twins adopt the go.gala.fyi/stdlib module
status: Done
assignee:
  - '@opencode'
created_date: '2026-10-10 10:16'
updated_date: '2026-10-10 10:49'
labels: []
dependencies: []
references:
  - server/GALA.md
  - server/GALA_COMPILER
  - server/scripts/gala/verify.sh
  - .agents/skills/gala-from-go/SKILL.md
  - 'https://github.com/martianoff/gala/issues/740'
  - 'https://github.com/martianoff/gala/pull/743'
  - >-
    backlog/tasks/task-0349 -
    Advance-the-GALA-translation-loop-cursor-and-land-its-runtime-free-twins.md
documentation:
  - server/GALA.md
  - server/README.md
priority: medium
ordinal: 356005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Context

Every committed twin under `server/` is runtime-free by the contract in `server/GALA.md`: its generated Go names no `martianoff/gala/...` or `go.gala.fyi/stdlib/...` package, `server/go.mod` carries no GALA runtime require, and `server/scripts/gala/verify.sh` rejects both import paths. That contract is why twins bind raw Go values with `var`, keep map and slice literals, multi-value results, and `Hash`/`Compare` in handwritten `.go` siblings, and avoid `val`, `Try`, GALA structs, and sealed types.

The pinned compiler (GALA 0.87.1 post-release HEAD, flake rev `cd2fdcb5`) adds `gala transpile --stdlib-module go.gala.fyi/stdlib` (upstream PR #743, closing this repository's issue #740). A scratch transpile on 2026-10-10 rewrote an emitted `martianoff/gala/std` import to `go.gala.fyi/stdlib/std`, and the published module is a plain Go module at v0.87.1, so the earlier objection that only `gala export` could rewrite imports and that a `replace` was refused by Go no longer applies to `gala transpile` output.

The decision was surfaced by TASK-0349's iteration 10 compiler bump and is deliberately deferred there: the runtime-free contract is unchanged until this task decides otherwise.

## Decision required

Keep the runtime-free contract, or adopt the published module for server twins. If adopting, the decision must also fix the scope (all existing twins, new twins only, or named candidates such as the still-blocked `models/` group or `postgres/`) and the migration and documentation work it triggers.

Weigh at least: the server build and test paths (`server/Dockerfile`, Compose stack, `verify.sh` regeneration, offline Go caches), the GALA-runtime dependency policy TASK-0347 established when it removed the vendored stdlib, and the runtime-free guidance in the `gala-from-go` and `gala-loop` skills.

This task decides and, if it changes course, creates the follow-up tasks; it does not migrate twins itself.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A reproducible scratch check on the pinned compiler (flake rev cd2fdcb5, `gala transpile --stdlib-module go.gala.fyi/stdlib`) shows generated Go building and running in a plain Go module that requires the published stdlib; the task notes record the module version, its transitive requirements, and the exact commands.
- [x] #2 The keep-or-adopt decision and, if adopting, its scope are recorded in `server/GALA.md`, with the runtime-free contract wording either confirmed with the new rationale or replaced by the adopted dependency policy.
- [ ] #3 If the decision is keep: `server/GALA.md` states why the capability stays unused, and no runtime, verification, or skill behavior changes.
- [x] #4 If the decision is adopt: follow-up Backlog tasks exist for the go.mod/verify.sh contract change, the twin migration scope, and the gala-from-go and gala-loop updates; no twin is migrated in this task.
- [x] #5 Changed Markdown files are formatted with `npm run format:markdown`.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Phase 0: record provenance (nix develop, gala 0.87.1 @ cd2fdcb5, extraction 7b1dd208...) and a green 22-twin `server/scripts/gala/verify.sh` baseline.
2. Phase 1 (AC#1): scratch GALA source using runtime constructs, transpiled with `gala transpile --stdlib-module go.gala.fyi/stdlib`; build and run the generated Go in a plain Go module requiring `go.gala.fyi/stdlib v0.87.1`; record module version, transitive requirements, exact commands. Add a minimal scratch Docker check mirroring `go mod download` + build + run, with a warmed-cache fallback if egress is flaky. Verify a runtime-free twin regenerates byte-identically under the flag.
3. Phase 2 (AC#2): record the adopt decision in `server/GALA.md`; scope is all existing twins re-baselined onto the module in an API/wire-shape-preserving migration, plus new twins and cursor candidates; `martianoff/gala` stays banned; no go.mod change and no twin migration in this task.
4. Phase 3 (AC#4): search existing tasks, then create three follow-ups: contract adoption (go.mod/go.sum, verify.sh, README/registry), gala-from-go and gala-loop updates, and the runtime migration task; leave TASK-0349 as-is.
5. Phase 4: `npm run format:markdown`, record evidence in task notes, write the final summary, and mark the task Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Provenance (2026-10-10)

- `nix develop`: `GALA version 0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c`; extraction `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa` matches `server/GALA_COMPILER`.
- Baseline `server/scripts/gala/verify.sh`: `OK (22 twins)`.

## AC#1 scratch check (host, pinned compiler)

Source `/tmp/opencode/task-0353/check/main.gala` uses `val`, a GALA `struct Person(Name string, Age int)`, `opaque type Count int`, a `sealed type Shape` with a `match`, and `std.Try`.

- `gala transpile --stdlib-module go.gala.fyi/stdlib -i main.gala -o main.go` -> exit 0; emitted Go imports only `go.gala.fyi/stdlib/std` (no `martianoff/gala`; 30 `std.` references).
- `go mod init stdlibcheck && go mod edit -require=go.gala.fyi/stdlib@v0.87.1 && GOFLAGS=-mod=mod go mod tidy` -> `go.mod` carries exactly that one require (zero transitive requirements); `go.sum`:
  - `go.gala.fyi/stdlib v0.87.1 h1:gNJcSfJcN3+00fV3D0zZoTdQDKdMeg8d61NVHAofA1I=`
  - `go.gala.fyi/stdlib v0.87.1/go.mod h1:TQbUMH1dZYQvO8CqtIO8fwayMqE2fY1XmA/ePR/5B+M=`
- `go build .` exit 0; `go run .` printed `person Ada 36 false`, `area 12`, `count 1453600679`, `half 5`, `fallback 0`; `gofmt -l main.go` clean.
- Module metadata: publish origin `github.com/martianoff/gala-stdlib` tag `v0.87.1` (commit `b7a0d920c4c2780a29940a15df58ffa3f4e41232`), module `go 1.24`, no requires, no `replace`; the module `VERSION` says `gala 0.87.1` at `669c958c` (the previous lock, which shares the stdlib extraction fingerprint with the pinned `cd2fdcb5`).

## AC#1 identity check

- `cd server/eventid && gala transpile --stdlib-module go.gala.fyi/stdlib -i eventid.gala -o /tmp/opencode/task-0353/eventid-flagged.go` -> byte-identical to committed `server/eventid/eventid.go` (`diff` clean), so the flag can be applied to runtime-free twins with no output change.

## AC#1 Docker check

- `docker build` of `/tmp/opencode/task-0353/docker-check` (Dockerfile mirroring `server/Dockerfile` `testdeps` -> `builder` -> runtime on `golang:1.26.4-alpine3.24`) with an empty module cache: `go mod download` ran 12.1s fetching from the network, `go build` succeeded, and `docker run --rm stdlibcheck:scratch` printed the same program output.
- Isolated-stack shape (source mount + fresh `GOMODCACHE` volume + `go run .`): the first fetch hit a transient `proxy.golang.org` zip EOF (this environment's egress to the proxy is flaky; the host saw the same), then a retried `go mod download` succeeded and `go run .` on the warmed volume printed the same output. CI's go.sum-keyed cache volume seeded from `actions/cache` is that warmed-cache path; a fresh runner downloads once with normal egress.
- No GALA compiler is needed in any image; only committed generated Go is built.

## AC#2 decision record

- `server/GALA.md` gained `### Adopted runtime (decided 2026-10-10)` (adopt decision; scope = every existing twin plus new twins and cursor candidates, surface-preserving only; `martianoff/gala` banned and `go.gala.fyi/stdlib/...` the only runtime path; module facts; scratch and Docker evidence; follow-up task IDs) and `### Runtime-free rules (in force until TASK-0353.01)` so the current rules stay truthful until the contract change lands; the compiler section and the #740 prose were updated to the decision.
- No `server/go.mod` change, no `verify.sh` change, and no twin migrated in this task.

## AC#4 follow-up tasks

- TASK-0353.01 (High, no deps): contract adoption - go.mod/go.sum require, verify.sh `--stdlib-module` plus guard change, registry/README commands, Docker/Compose evidence.
- TASK-0353.02 (Medium, depends on .01): gala-from-go and gala-loop update for the adopted runtime.
- TASK-0353.03 (High, depends on .01/.02): migrate every existing twin and the cursor, surface-preserving, with per-iteration loop evidence.
- TASK-0349 is left as-is, per the decision task's scope.

## AC#5 formatting

- `npm run format:markdown` (no file changed) and `npm run format:markdown:check` clean.

## Worktree

- Changed: `server/GALA.md` plus the TASK-0353 file and the three new subtask files (Backlog MCP). The pre-existing `backlog/backlog.md` modification was left untouched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Decided to adopt the published `go.gala.fyi/stdlib` v0.87.1 module for server GALA twins, with the scope covering every existing twin plus new twins and cursor candidates, migrating a twin only where the emitted Go preserves its Go-facing API and wire format; `martianoff/gala` stays banned and the only runtime import path is the module pinned in `server/go.mod`.

Verified on the pinned compiler (flake rev `cd2fdcb5`, gala 0.87.1, extraction `7b1dd208...`): a scratch source using `val`, a GALA struct, an `opaque type`, a sealed type with a `match`, and `std.Try` transpiled with `gala transpile --stdlib-module go.gala.fyi/stdlib` to Go importing only `go.gala.fyi/stdlib/std`; that Go built and ran in a plain Go module requiring `go.gala.fyi/stdlib v0.87.1`, which has zero transitive requirements, and `server/eventid/eventid.go` regenerated byte-identically under the flag. A minimal Docker build mirroring `server/Dockerfile` fetched the module with an empty cache and ran the binary, and a test-stack-shaped container (source mount + `GOMODCACHE` volume + `go run`) fetched it into the volume and ran, so the Docker and CI cache paths need no image or workflow change.

Recorded the decision, scope, module facts, evidence, and interim runtime-free rules in `server/GALA.md`. Created follow-ups TASK-0353.01 (go.mod/verify.sh/ledger contract), TASK-0353.02 (gala-from-go and gala-loop updates), and TASK-0353.03 (twin and cursor migration); TASK-0349 is left as-is. No `server/go.mod` change and no twin migration in this task.

Evidence and exact commands are in the task notes; `npm run format:markdown` and `format:markdown:check` are clean. Not committed: the user did not ask for a commit.
<!-- SECTION:FINAL_SUMMARY:END -->

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
