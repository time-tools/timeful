---
id: TASK-0365
title: >-
  Bump the GALA compiler pin to upstream HEAD 8ba02b2b (0.93.0) and re-baseline
  the twins
status: Done
assignee: []
created_date: '2026-10-25 11:46'
updated_date: '2026-10-25 11:49'
labels:
  - gala
  - tooling
dependencies: []
references:
  - flake.lock
  - server/GALA.md
  - server/GALA_COMPILER
  - server/scripts/gala/verify.py
  - server/scripts/gala/verify.sh
  - 'https://github.com/martianoff/gala/releases/tag/0.92.2'
  - 'https://github.com/martianoff/gala/releases/tag/0.93.0'
priority: medium
type: chore
ordinal: 369300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The GALA flake input sits at `911cc059` (0.92.1) while upstream `master` is `8ba02b2b`, the merge of the 0.93.0 version bump, 66 commits ahead. The range covers the 0.92.2 and 0.93.0 releases: path-resolution caching and `testdata` copying in 0.92.2, and in 0.93.0 `Option.Exists`/`Option.ForAll`, hash-based `Distinct`, the new `GALA-E0081`, `GALA-E0082`, and `GALA-E0083` diagnostics, cross-package default and package-level `val` inference fixes, a pointer-preserving `std.Copy`, and the named-import qualifier behavior change. `server/GALA.md` requires a bump iteration whenever the flake rev and `server/GALA_COMPILER` disagree: refresh the extraction provenance, move the paired `go.gala.fyi/stdlib` module, regenerate all committed twins, re-run the committed probes, and run the build and the canonical Compose backend suite. The 0.92.1 bump also left the verifier reading the extraction marker from `~/.gala` even though `.envrc` now sets `GALA_HOME` to `.local/gala`, so the provenance check fails under the documented dev shell and needs the same fix here.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 flake.lock locks martianoff/gala at 8ba02b2b, and `nix develop` provides `GALA version 0.93.0` with a working `gala-local`.
- [x] #2 server/GALA_COMPILER records 0.93.0, flake rev 8ba02b2b, extraction `0.93.0 6368c09381c6f9d8d72f1551101ca48207cb2e1c34122627ca928261ad05b13c`, and date 2026-10-25.
- [x] #3 server/go.mod and server/go.sum require go.gala.fyi/stdlib v0.93.0, and the module builds against it.
- [x] #4 scripts/gala/verify.sh --write regenerates all 30 twins with no diff, verify.sh reports OK, and any shape change is recorded rather than papered over.
- [x] #5 All six committed probes are re-run on 0.93.0, their notes carry a Re-verified (2026-10-25) section, and the cursor records rev 8ba02b2b.
- [x] #6 server/GALA.md records the new provenance, the 0.92.2/0.93.0 range changes, and the extraction-marker lookup through GALA_HOME.
- [x] #7 go build ./... in server/ and the canonical Compose backend test sequence pass under the bumped compiler.
- [x] #8 verify.py resolves the extraction marker from GALA_HOME when set, and the provenance check passes under .local/gala as well as the ~/.gala fallback.
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
1. Record baseline provenance and the current worktree state.
2. Update the gala flake input, build the dev shell, and confirm GALA 0.93.0 plus gala-local.
3. Force the stdlib extraction under GALA_HOME and write server/GALA_COMPILER.
4. Move server/go.mod and go.sum to go.gala.fyi/stdlib v0.93.0.
5. Run scripts/gala/verify.sh --write and review the per-twin diff.
6. Re-run the six committed probes, add Re-verified sections, and update the GALA.md cursor and provenance.
7. Run go build ./... and the canonical Compose backend sequence.
8. Teach verify.py to resolve the extraction marker under GALA_HOME and update the ledger's instructions.
9. Format Markdown, refresh the code graph, finalize the task, and commit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Baseline: `gala version` 0.92.1, flake rev 911cc059, extraction `0.92.1 2c6059ca…`; worktree clean. `nix flake update gala` resolved 8ba02b2b and built GALA 0.93.0 plus `gala-local`.

The first `verify.sh` run failed only the provenance extraction check because the verifier still read `~/.gala` while `.envrc`/`GALA_HOME` puts the compiler state at `.local/gala`; the new marker was `0.93.0 6368c09381c6f9d8d72f1551101ca48207cb2e1c34122627ca928261ad05b13c`. The message named the mismatch: the recorded marker path was empty.

`go get go.gala.fyi/stdlib@v0.93.0` initially failed on the read-only `~/go` module cache; it succeeded after pointing `GOPATH`/`GOMODCACHE` at `.local/go` (TASK-0366). `go mod tidy` then dropped the 0.92.1 sums, leaving only 0.93.0.

`verify.sh --write` reported 30 twins unchanged, gofmt ok, orphan scan ok, `go build: ok`, and `verify.sh: OK (30 twins)`. The canonical Compose backend suite then passed, with the container downloading `go.gala.fyi/stdlib v0.93.0`.

Probes re-run on 2026-10-25: models-empty-struct and models-fixed-array reproduced their two parse errors, models-struct-tag its tag parse error, routes-group-chan-recover its chan/defer/go/recover refusals, routes-group-embedded-field its bare-embed parse error, and models-defined-scalar-methods reproduced the runtime-backed default, the GALA-E0048 alias refusal, and the runtime-free suppression output.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Moved the `gala` flake input from `911cc059` (0.92.1) to `8ba02b2b`, the 0.93.0 post-release `master` head, and completed the ledger's bump iteration. The 0.92.2/0.93.0 range changes no generated Go here: all 30 twins regenerate byte-identically, so no `.gala` source or sibling changed.

Changed: `flake.lock` (gala rev/narHash), `server/GALA_COMPILER` (0.93.0, rev, extraction `0.93.0 6368c0…`, 2026-10-25), `server/go.mod`/`go.sum` (`go.gala.fyi/stdlib v0.93.0`, old sums removed by `go mod tidy`), `server/GALA.md` (provenance table, range changes, cursor rev, probe re-verification dates, `$GALA_HOME` extraction instructions), `server/scripts/gala/verify.py` (new `gala_state_dir()` so the extraction marker is read from `GALA_HOME` when set, falling back to `~/.gala`), and all six probe notes (Re-verified 2026-10-25 sections).

Evidence: `verify.sh` first failed only on `PROVENANCE: … records extraction … but /home/…/.gala/stdlib/v0.93.0/.stdlib-extracted is ''` under the relocated `GALA_HOME`; after the marker was taken from `.local/gala` and `GALA_COMPILER` was written, `verify.sh --write` reported `twins: 30 discovered, 30 transpiled, 30 unchanged`, `gofmt ok`, `orphan scan ok`, `go build: ok`, `verify.sh: OK (30 twins)`. The canonical Compose backend sequence passed with the container downloading `go.gala.fyi/stdlib v0.93.0` (routes 2.599s, postgres 7.067s). The six probes reproduced their recorded diagnostics (including the `opaque type` default/alias/suppression variants). `go build ./...` passed again after `go mod tidy`; `npm run format:markdown:check` passed; `codebase-memory-mcp cli index_repository` indexed 9102 nodes / 29467 edges.

Follow-up: the verifier's `~/.gala` assumption came from the older dev setup and is now covered here; TASK-0360 (0.92.0 bump) remains In Progress from the previous session even though the 0.92.1 bump landed in `4a7a5901`.
<!-- SECTION:FINAL_SUMMARY:END -->
