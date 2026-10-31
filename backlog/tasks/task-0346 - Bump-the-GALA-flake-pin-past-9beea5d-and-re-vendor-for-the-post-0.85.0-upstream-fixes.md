---
id: TASK-0346
title: >-
  Bump the GALA flake pin past 9beea5d and re-vendor for the post-0.85.0
  upstream fixes
status: Done
assignee:
  - '@opencode'
created_date: '2026-10-04 23:38'
updated_date: '2026-10-04 23:51'
labels:
  - gala
  - tooling
dependencies: []
references:
  - flake.lock
  - server/GALA.md
  - server/scripts/gala/verify.sh
  - .agents/skills/gala-loop/SKILL.md
  - 'https://github.com/martianoff/gala/issues/691'
  - 'https://github.com/martianoff/gala/issues/692'
  - 'https://github.com/martianoff/gala/issues/695'
documentation:
  - server/GALA.md
  - server/README.md
priority: medium
type: task
ordinal: 350005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Upstream HEAD moved from the locked `9beea5dc350ab566e2e1e31add863afd7d95d29f` to `d38fadb844f0849c6c6ac8a47e8f91e025eace6b` (still version 0.85.0), eight commits with runtime and codegen changes: #690 pins the stdlib transpiler to release 0.85.0, #691 fixes lowercase sealed variant exhaustiveness/match, #692 makes the JSON decoder strict and fixes YAML escaping (json/codec.gala, yaml/codec.go), #695 lifts a Go method returning several results to Try/Tuple, #702 makes `gala build` handle a Go-only main package under cmd/, #694/#701 are CI lanes, #703 is docs. Because the flake rev and server/third_party/gala/VENDORED_FROM disagree, this is a sync iteration and never mixed with a translation. Constraints: follow the corrected re-vendor procedure in server/GALA.md (marker refresh first, because several compilers share the 0.85.0 version string); no hand-edited twin; a regression in a re-baselined twin stops the sync and becomes a finding rather than being papered over; the canonical backend test sequence is required because runtime and codegen both changed. The loop skill is .agents/skills/gala-loop/SKILL.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The `gala` input in `flake.lock` is bumped to upstream HEAD `d38fadb844f0849c6c6ac8a47e8f91e025eace6b` (or a recorded newer rev) and the locked rev is recorded in the task notes
- [x] #2 The sync follows `server/GALA.md`'s re-vendor procedure: extraction marker refreshed by the pinned compiler, extraction copied over `third_party/gala/`, per-package `go.mod`/`go.sum`/marker removed, root `go.mod` rewritten, `go build ./...` green inside the vendored tree
- [x] #3 `server/third_party/gala/VENDORED_FROM` records the new compiler version, flake rev, source path, extraction fingerprint, and date
- [x] #4 `scripts/gala/verify.sh --write` runs under the dev shell; the per-twin diff is reviewed and every shape change is recorded, with no regression papered over
- [x] #5 The open findings touched by #691, #692, and #695 are re-checked on the new compiler and retired, updated, or retained with recorded evidence; `server/GALA.md`'s compiler table and upstream report index reflect the new lock
- [x] #6 Evidence is recorded in the task: provenance commands, extraction marker, per-twin diff review, `verify.sh` result, `go build ./...`, and the canonical backend test sequence result
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
1. Record provenance: `gala version`, `which gala`, `jq -r '.nodes.gala.locked.rev' flake.lock`, `cat server/third_party/gala/VENDORED_FROM`.
2. `nix flake update gala`; confirm the new rev and that `nix develop --command gala version` reports 0.85.0 from the dev-shell store path.
3. Refresh the extraction marker (delete `~/.gala/stdlib/v0.85.0/.stdlib-extracted`, transpile `server/eventid/eventid.gala` with the dev-shell compiler) and record the fingerprint.
4. Re-vendor: copy the extraction over `third_party/gala/`, drop per-package `go.mod`/`go.sum` and the marker, keep `LICENSE`, rewrite the root `go.mod`, `go build ./...` inside the vendored tree.
5. `nix develop --command server/scripts/gala/verify.sh --write`; review the per-twin diff and record every shape change; a regression stops the sync.
6. Re-check the findings touched by #691 (lowercase sealed variants), #692 (strict JSON / YAML escaping), and #695 (Go multi-result methods lifted to Try/Tuple) on the new compiler with scratch repros.
7. Update `server/GALA.md` (compiler table, report index, open findings) and `third_party/gala/VENDORED_FROM`.
8. Run `go build ./...` in `server/`, the canonical backend test sequence from `server/README.md`, the Markdown/root format checks, and refresh the code graph.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Sync evidence (2026-10-04)
- Provenance before: `GALA version 0.85.0`, rev `9beea5dc350ab566e2e1e31add863afd7d95d29f`, vendored extraction `0.85.0 09598ebb...`, dev-shell compiler `/nix/store/gy8vz739ilig9h1k1blbiavkkr2g3339-gala-0.85.0/bin/gala`.
- Bump: `nix flake update gala` locked `d38fadb844f0849c6c6ac8a47e8f91e025eace6b` (2026-10-04). `nix develop --command gala version` prints `GALA version 0.85.0` from `/nix/store/5jy1pg6q86i36bldfydbj8mwwy9hxzlb-gala-0.85.0/bin/gala`, so the #678/#680 `gala-local` path still builds.
- Upstream content: 8 commits past the lock — #690, #691, #692, #695, #702, #694, #701, #703; the only runtime files upstream touched are `json/codec.gala`, `json/codec.gen.go`, and `yaml/codec.go`.
- Extraction refresh: deleted `~/.gala/stdlib/v0.85.0/.stdlib-extracted` (`0.85.0 09598ebb...`), transpiled `server/eventid/eventid.gala` to `/tmp/eventid.gala-sync.go` with the dev-shell compiler; new marker `0.85.0 8206b230308f8892c8c41dbb1fdf034c2010acc5b120f7f766c29edcfe9faebd`.
- Re-vendor: copied the extraction over `server/third_party/gala/`; deleted 22 per-package `go.mod` files and the marker; retained `LICENSE` and the root `go.mod` (`module martianoff/gala`, `go 1.24`, matching every snapshot package); `go build ./...` inside the vendored tree exit 0. `git status` for the tree shows exactly the three upstream-changed files.
- `verify.sh --write` under the dev shell: provenance matched, no WRITE/DRIFT/GOFMT/ORPHAN lines, `verify.sh: OK (15 twins)`. All 15 twins regenerate byte-identically on the new compiler; no re-baseline, no regression.

## Finding re-checks (2026-10-04)
- #691 lowercase sealed variants: upstream `examples/lowercase_sealed_variants.gala` transpiled and ran via `gala run`, exit 0; exhaustive matches without a default and bare zero-field patterns behave as fixed.
- #692 strict JSON/YAML: scratch module `replace martianoff/gala => server/third_party/gala`; decoder rejects `01` with `json at pos 0: invalid number "01"` and depth 10001 with `json at pos 10000: exceeded max depth 10000`. No open ledger finding maps to this runtime change; only the vendored tree changed.
- #695 Go method multi-results: upstream `examples/go_method_multi_results.gala` with the `warehouse_bridge` Go package copied locally transpiled, built, and ran, exit 0 (Try from `(T, error)` methods, Tuple destructuring from `(T, bool)`, `FromError` on error-only methods, chained calls). The ledger's Go-style multi-value return signature finding is declaration-side and stays: `func multi() (int, error)` is still `mismatched input '(' expecting {'=', '{'}` at transpile. `server/GALA.md` names the #695 caveat next to that finding.
- `server/GALA.md`: compiler paragraph and table now record rev `d38fadb`; report index gains #691/#692/#695 as closed and fixed (verified here).

## Verification (2026-10-04)
- `server` `go build ./...` green (inside `verify.sh`).
- Canonical backend sequence: `docker volume create` per volume (the local Docker 28 CLI rejects two names in one call), `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test postgres-test-bootstrap postgres-test-migrate`, then `run --rm server-route-test` exit 0 with 15 packages `ok` and no `FAIL` (log `/tmp/sync-checks/backend-tests.log`).
- `npm run format:markdown` and `npm run format:markdown:check` clean.
- `codebase-memory-mcp cli index_repository --repo-path .`: 11334 nodes, 46461 edges.
- e2e not run: no frontend/browser surface changed. Swagger not regenerated: no route annotations changed. Root `fmt:check` not run: no `scripts/` or `prettier/` changes.
- Not committed; the working tree also carries a pre-existing unrelated `backlog/backlog.md` edit that was left untouched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Bumped the GALA flake pin from `9beea5dc` (0.85.0) to `d38fadb844f0849c6c6ac8a47e8f91e025eace6b` (upstream HEAD) and re-vendored the transpiled runtime, following the corrected procedure in server/GALA.md. The rev still reports `GALA version 0.85.0` but carries the post-release fixes #690 (stdlib transpiler pin), #691 (lowercase sealed variants), #692 (strict JSON decoder; YAML escapes), #695 (Go method multi-results lifted to Try/Tuple), and #702 (`gala build` under cmd/), plus CI #694/#701 and docs #703. `nix develop` and `gala-local` build on the new rev.\n\nDelivered:\n- `flake.lock` gala node bumped; dev-shell compiler is `/nix/store/5jy1pg6q86i36bldfydbj8mwwy9hxzlb-gala-0.85.0/bin/gala`.\n- Fresh extraction written by that compiler (marker `0.85.0 8206b230...`) and copied over `server/third_party/gala/`; per-package go.mod/marker removed; root `go.mod` kept at `go 1.24`, matching every snapshot package; vendored `go build ./...` green. Runtime changes are exactly upstream's `json/codec.gala`, `json/codec.gen.go`, and `yaml/codec.go`.\n- `verify.sh --write` printed no WRITE/DRIFT/GOFMT/ORPHAN lines; `verify.sh: OK (15 twins)` with matching provenance. No twin was re-baselined and no regression appeared.\n- Findings re-checked on the pinned compiler: #691 via upstream `lowercase_sealed_variants.gala` (exhaustive matches without a default, run exit 0), #692 via a scratch consumer of the vendored json package (rejects `01` and depth > 10000), #695 via the upstream Go-method multi-results example adapted to the vendored runtime (transpiles, builds, runs exit 0). A GALA function still cannot declare a Go-style multi-value signature (parse error at `func multi() (int, error)`), so the ledger finding stays with the #695 caveat noted.\n- `server/GALA.md` compiler table/paragraph and the upstream report index record #691/#692/#695 as fixed (verified here); `server/third_party/gala/VENDORED_FROM` records compiler 0.85.0, rev `d38fadb`, source `~/.gala/stdlib/v0.85.0`, marker `8206b230...`, date.\n- Canonical backend Compose suite: exit 0, 15 packages `ok`, no FAIL; `go build ./...` green; `npm run format:markdown` + `:check` clean; code graph refreshed (11334 nodes, 46461 edges).\n\nNot committed: the user did not ask for a commit. The pre-existing unrelated `backlog/backlog.md` edits were left untouched.
<!-- SECTION:FINAL_SUMMARY:END -->
