---
id: TASK-0348
title: Bump the GALA compiler pin to upstream HEAD a888e824 (#712)
status: Done
assignee: []
created_date: '2026-10-05 08:22'
updated_date: '2026-10-05 08:32'
labels:
  - gala
  - tooling
dependencies: []
references:
  - flake.lock
  - server/GALA.md
  - server/GALA_COMPILER
  - server/scripts/gala/verify.sh
  - 'https://github.com/martianoff/gala/pull/712'
  - 'https://github.com/martianoff/gala/issues/687'
modified_files:
  - flake.lock
  - server/GALA_COMPILER
  - server/GALA.md
priority: medium
type: chore
ordinal: 352005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Upstream master moved past the locked `ddbe839b` to `a888e824` with one commit: #712 "Go's encoding/json, YAML and fmt see an immutable field's value, not {}" (closes #687), which adds MarshalJSON/UnmarshalJSON, MarshalYAML/UnmarshalYAML, fmt.Formatter, and IsZero to `std.Immutable[T]` plus tests, examples, and docs. The repository's twins are runtime-free, so this is expected to be a re-baseline with no codegen change to the committed twins, but only verify.sh can establish that.

Constraints from server/GALA.md: bump and translation are separate iterations; never translate across a rev mismatch; never hand-edit a generated twin; never relax a test; transpile from the package directory. The version string stays 0.85.0, so the compiler is proved by the `.stdlib-extracted` fingerprint, not by `gala version`.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 flake.lock's gala input is updated to a888e824 (or a recorded newer rev), `nix develop` builds gala-local and puts the newly pinned compiler on PATH
- [x] #2 The extraction marker is refreshed by a transpile with the pinned compiler and its fingerprint, version, flake rev, and date are recorded in server/GALA_COMPILER
- [x] #3 `nix develop --command server/scripts/gala/verify.sh --write` regenerates every committed twin, the per-twin diff is reviewed, and every shape change is recorded with no regression papered over
- [x] #4 `go build ./...` in server/ and the canonical backend test sequence in server/README.md pass under the bumped compiler
- [x] #5 server/GALA.md's compiler section, provenance table, and upstream report index record the new rev and the #712 content; the cursor and registry are unchanged by a no-drift bump
- [x] #6 Changed Markdown is formatted, the code knowledge graph is refreshed, and root `npm run fmt:check` is run if scripts changed
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
## Provenance before (baseline)
- `nix develop -c gala version` = `GALA version 0.85.0`; `which gala` = `/nix/store/bwmk5ivbn7hq4pj3k3xc4zvar1blbyvq-gala-0.85.0/bin/gala`; `flake.lock` rev `ddbe839b5d038ef3445e1dac93f09e624144012f`; extraction marker `0.85.0 8206b230...`; `verify.sh` OK (15 twins). Ambient PATH `gala` is 0.84.1, so every command ran under `nix develop -c`.
- Precondition note: the worktree already carried an unrelated unstaged edit to `backlog/backlog.md`; it was left untouched. Both halves of all 15 twins were committed.

## Upstream delta
- `gh api repos/martianoff/gala/compare/ddbe839b...a888e824` = ahead_by 1: commit `a888e824ff53adb653bbd48dcc83f67788eeced4`, PR #712 "Go's encoding/json, YAML and fmt see an immutable field's value, not {}" (closes #687). It modifies `std/types.go` (MarshalJSON/UnmarshalJSON, MarshalYAML/UnmarshalYAML, fmt.Formatter, IsZero) plus tests, examples, and docs; the Go-facing transpiler paths and the version string are unchanged.

## Bump
- `NIX_CONFIG="access-tokens = github.com=$(gh auth token)" nix flake update gala` (same authenticated route as TASK-0347, because anonymous GitHub fetches hit timeouts/rate limits); token never printed. Locked rev `a888e824ff53adb653bbd48dcc83f67788eeced4`, lastModified 1791179377, narHash `sha256-x6gc0MyjSoLkdiEivV8bxIvKaZ0SHg7sDnL6nT7W5wA=`, input type `github:martianoff/gala` preserved.
- Dev shell rebuilt: `which gala` = `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`, so `gala-local` still builds.
- Extraction refresh: deleted `~/.gala/stdlib/v0.85.0/.stdlib-extracted`, transpiled `server/eventid/eventid.gala` to a scratch path with the pinned compiler; marker rewritten to `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`. The fingerprint changed because #712 edited the embedded stdlib snapshot; this is the proof the new compiler, not the old one, wrote the extraction.
- `server/GALA_COMPILER` updated with compiler, rev, extraction, and date.

## Re-baseline and review
- `nix develop -c server/scripts/gala/verify.sh --write`: no `WRITE`, `DRIFT`, `GOFMT`, `ORPHAN`, or `RUNTIME` lines; `verify.sh: OK (15 twins)`. The per-twin diff is empty: every twin regenerates byte-identically under `a888e824`, so the bump produced no codegen shape change and there is no regression to paper over. Runtime-free census unchanged.
- `nix develop -c server/scripts/gala/verify.sh` (check): provenance matches, `go build ./...` green, OK (15 twins).

## Tests
- Canonical Compose backend sequence from `server/README.md` (`.env.test` and cache volumes already present, so `cp`/`docker volume create` were no-ops): `up -d postgres-test postgres-test-bootstrap postgres-test-migrate`, then `run --rm server-route-test` exit 0 with every package `ok` and no `FAIL` (log `/tmp/opencode/task-0348-backend-tests.log`).

## Docs and checks
- `server/GALA.md`: compiler paragraph now names `a888e824` and #712 (closes #687) on top of the #699/#705-#710 rev; provenance table records the new rev and extraction. The upstream report index is unchanged because #712 is an upstream fix, not a report this repository filed; the twin registry and cursor are unchanged because no candidate was translated.
- `npm run format:markdown` + `:check` clean; root `npm run fmt:check` clean (scripts/prettier unchanged); code graph refreshed to 8906 nodes / 28215 edges.
- e2e not run: no frontend or browser surface changed. Swagger not regenerated: no route annotation changed. No commit made: the user did not ask for one.

Bump evidence: new compiler `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`; extraction `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`; `verify.sh --write` printed zero WRITE/DRIFT lines (15 twins byte-identical); backend Compose suite exit 0; logs `/tmp/opencode/task-0348-verify-write.log`, `/tmp/opencode/task-0348-verify-check.log`, `/tmp/opencode/task-0348-backend-tests.log`.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Bumped the GALA flake pin from `ddbe839b` to upstream master `a888e824` (PR #712, closes #687) and re-baselined all committed twins against it.

Delivered:
- `flake.lock` gala node locked to `a888e824ff53adb653bbd48dcc83f67788eeced4` (input type preserved); the rebuilt dev shell provides `/nix/store/ydla797gibrgplg2kwn0mwy8vhvj7mip-gala-0.85.0/bin/gala`, so `gala-local` builds and the version string stays 0.85.0.
- Extraction refreshed by the pinned compiler: `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`, changed from `8206b230...` because #712 edited the embedded stdlib snapshot. `server/GALA_COMPILER` records version, rev, fingerprint, and date.
- `verify.sh --write` regenerated all 15 twins: no WRITE/DRIFT/GOFMT/ORPHAN/RUNTIME lines, per-twin diff empty, so #712 changes no committed twin; no shape change or regression to record.
- `server/GALA.md` compiler paragraph and provenance table now name `a888e824` and #712 on top of the #699/#705-#710 rev; the registry and cursor are untouched.
- `verify.sh` check OK (15 twins) with `go build ./...` green; the canonical Compose backend suite passed, every package `ok`, no `FAIL`.

Evidence:
- Upstream delta: `gh api .../compare/ddbe839b...a888e824` ahead_by 1, one commit modifying `std/types.go`, tests, examples, docs.
- Logs: `/tmp/opencode/task-0348-verify-write.log`, `/tmp/opencode/task-0348-verify-check.log`, `/tmp/opencode/task-0348-backend-tests.log`.
- `npm run format:markdown` + `:check` clean; root `npm run fmt:check` clean; code graph refreshed (8906 nodes, 28215 edges).
- e2e and Swagger not required: no frontend surface and no route annotation changed. The pre-existing unrelated `backlog/backlog.md` edit was left untouched, and nothing was committed (not requested).
<!-- SECTION:FINAL_SUMMARY:END -->
