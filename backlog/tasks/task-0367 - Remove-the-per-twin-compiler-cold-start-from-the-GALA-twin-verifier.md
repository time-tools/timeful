---
id: TASK-0367
title: Remove the per-twin compiler cold start from the GALA twin verifier
status: Done
assignee:
  - dsh
created_date: '2026-10-27 10:40'
updated_date: '2026-10-27 11:14'
labels:
  - gala
  - tooling
  - performance
dependencies: []
references:
  - server/scripts/gala/verify.py
  - server/scripts/gala/verify.sh
  - >-
    backlog/tasks/task-0349 -
    Advance-the-GALA-translation-loop-cursor-and-land-its-runtime-free-twins.md
documentation:
  - server/GALA.md
  - .agents/skills/gala-loop/SKILL.md
priority: high
type: task
ordinal: 371300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`server/scripts/gala/verify.sh` gates every GALA translation batch, and its runtime is now the loop's dominant cost: recent full passes took 8-14 minutes for 48-77 twins, so batches are scheduled around it. Profiling shows the GALA work is not the problem: parsing, transforming, and generating a twin together take under 1% of its run. Almost all of the time is the analyzer rebuilding the twin package's Go type information, and because the verifier starts a fresh compiler process for every discovered `.gala` file, a folder of twins that shares one Go import graph pays that cold start once per twin. The verification contract itself must not change; this task removes the repeated cold start so an iteration can verify its batch in minutes rather than a quarter of an hour. Compiler-side improvements (fewer Go package resolutions, shared export-data loading) are upstream and out of scope.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The verifier pays the compiler cold start at most once per package instead of once per twin (one shared invocation covering several packages also satisfies this), while every twin is still checked independently for missing output, drift, a banned `martianoff/gala` import, and `gofmt` eligibility.
- [x] #2 Check mode on the current tree reports every committed twin `ok` with zero drift, and `--write` leaves every committed twin byte-identical and reports them unchanged.
- [x] #3 A deliberately modified `.gala` twin is reported as `DRIFT` in check mode and regenerated only in `--write` mode; a twin whose `.go` file is missing is reported as `MISSING`; a twin that does not transpile is reported individually with its path and compiler diagnostic while the rest of its package and the remaining packages are still verified.
- [x] #4 The command-line and log contract is unchanged: per-twin progress lines, `--check`, `--write`, `--verbose`, `--log FILE`, `--no-log`, the summary counts, and the exit codes behave as they do today.
- [x] #5 A regression check that does not need a full real-compiler run covers the shared-invocation behavior and the per-twin failure attribution (a stub compiler executable in a temporary fixture is acceptable).
- [x] #6 A full `server/scripts/gala/verify.sh` pass on the current tree is green, and the task records the twins-phase wall time before and after the change with the command and machine used.
- [x] #7 The verifier's help/docstring and the verification contract in `server/GALA.md` describe the shared invocation model and the measured speedup; changed Markdown is formatted.
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
## Final implementation plan (updated 2026-10-27 after execution)

### Approach as landed

1. `server/scripts/gala/verify.py`
   - `group_sources()` splits the sorted twins into consecutive source directories.
   - `check_package()` runs one shared `gala transpile-package --scan --stdlib-module go.gala.fyi/stdlib --inputs <basenames> --outputs <scratch paths>` per directory with cwd set to the package directory.
   - `check_twin()` keeps every per-twin check (MISSING committed twin, scratch output presence, banned `martianoff/gala`, byte identity in `--check` / regeneration in `--write`, `gofmt` target, RUNTIME) and the exact `[i/N pct%] path ... status` line format. A twin whose shared output is missing is retried alone with `gala transpile`, so its own path and compiler diagnostic are reported while the rest of the package and the remaining packages continue; the shared failure output is recorded in the run log.
   - `--verbose` prints the shared command and the shared or retried compiler output; every CLI flag, summary count, and exit code is unchanged.
2. `server/scripts/gala/test_verify.py` — stdlib `unittest` with a temporary git fixture and stub `gala`/`go`/`gofmt` that log every invocation; six cases cover one shared call per package with no per-twin calls, drift plus `--write` repair, MISSING, per-twin RUNTIME, per-twin TRANSPILE attribution with the diagnostic and the rest still verified, and the CLI/log contract.
3. Documentation — `verify.py` module docstring and argparse help, a `server/GALA.md` "Verifier invocation model" section with the measured speedup, and the verifier paragraph in `server/README.md`.

### Evidence

- Before (HEAD verifier copy in `tmp/task-0367/`, same tree and machine): `twins: 77 discovered, 77 transpiled, 77 ok (12m00s)`, total 12m05s.
- After: `server/scripts/gala/verify.sh --check --log tmp/task-0367/verify-after.log` -> `77 ok (4m14s)`, total 4m17s; `server/scripts/gala/verify.sh --write --log tmp/task-0367/verify-write.log` -> `77 discovered, 77 transpiled, 77 unchanged (4m49s)`, total 4m52s, both `verify.sh: OK (77 twins)` and no generated file modified.
- Machine: `Eyjafjallajokull`, Linux 6.8.0-147-generic x86_64, Intel Core i7-1165G7, 8 CPUs, GALA 0.93.0, Go 1.26.7.
- `python3 server/scripts/gala/test_verify.py`: 6 tests, OK.
- `npm run format:markdown`, `npm run format:markdown:check`, `npm run fmt:check`: pass.

### Deviations from the pre-implementation plan

- Grouping uses the source directory (plus the compiler's own `--scan` sibling resolution) instead of reading package clauses; this is simpler and keeps the compiler in charge of package membership.
- No dim fallback announcement or heartbeat lines were added; a failed shared invocation's output is recorded in the run log and printed under `--verbose`, keeping terminal output close to the old contract.
- `server/README.md` was updated too, because it documents the verifier's behavior.

### Findings outside scope

- `docs/environments.md` has a pre-existing `sentences-per-line/one` lint error on the `VITE_ENABLE_SIGN_IN` bullet (present at HEAD and unrelated to this task); `format:markdown` does not rewrite it, and the Markdown changed by this task passes both formatting commands.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Progress 2026-10-27

Implemented the shared invocation in `server/scripts/gala/verify.py`:

- Twins are grouped into consecutive source directories (`group_sources`) and each directory gets one `gala transpile-package --scan --stdlib-module go.gala.fyi/stdlib --inputs ... --outputs ...` run with cwd set to the package directory; the inputs stay the discovered twins even when one has no committed `.go` file, so sibling resolution and per-twin control flow match the old per-file call.
- Each twin still gets its own `[i/N pct%] path ... status` line and its own checks: MISSING, scratch output presence, banned `martianoff/gala`, byte-identity (or `--write` regeneration), `gofmt` target list, RUNTIME.
- When the shared run leaves a twin without an output, that twin alone is retried with `gala transpile` so its path and compiler diagnostic stay attributable; the rest of the package and the remaining packages continue. The failed shared output goes to the run log.
- `--verbose` prints the shared command and the compiler output of a package or a retried twin; all CLI flags, summary counts, and exit codes are unchanged.

Added `server/scripts/gala/test_verify.py` (stdlib unittest, temporary git fixture, stub `gala`/`go`/`gofmt` that log every invocation). `python3 server/scripts/gala/test_verify.py` passes 6 tests: one shared invocation per package with no per-twin calls, per-twin drift + `--write` repair, MISSING, per-twin RUNTIME, per-twin TRANSPILE attribution with the diagnostic and the rest still verified, and the CLI/log contract.

Measured twins-phase wall time on `Eyjafjallajokull` (Linux 6.8.0-147-generic x86_64, Intel Core i7-1165G7, 8 CPUs, GALA 0.93.0, Go 1.26.7), same checkout and warm compiler cache:

- Before: `python3 tmp/task-0367/verify-baseline.py --check --log tmp/task-0367/verify-baseline.log` (HEAD verifier copy) -> `twins: 77 discovered, 77 transpiled, 77 ok (12m00s)`, total 12m05s (`go build` 3.9s).
- After: `server/scripts/gala/verify.sh --check --log tmp/task-0367/verify-after.log` -> `twins: 77 discovered, 77 transpiled, 77 ok (4m14s)`, total 4m17s (`go build` 2.2s).
- Both runs report all 77 twins `ok` with zero drift; ratio 12m00s / 4m14s = 2.8x in the twins phase.

Next evidence: the `--write` pass on the same tree must report all 77 unchanged, then the Markdown formatting and docs checks.

## Completion evidence 2026-10-27

- `server/scripts/gala/verify.sh --check --log tmp/task-0367/verify-after.log`: `twins: 77 discovered, 77 transpiled, 77 ok (4m14s)`, `gofmt: 77 checked, ok`, `orphan scan: 222 Go files, ok`, `go build: ok`, total 4m17s, exit 0.
- `server/scripts/gala/verify.sh --write --log tmp/task-0367/verify-write.log`: `twins: 77 discovered, 77 transpiled, 77 unchanged (4m49s)`, total 4m52s, exit 0; `git status` showed no `.gala` or generated `.go` changed afterwards.
- Baseline HEAD verifier on the same tree: `twins ... (12m00s)`, total 12m05s. Twins phase is 2.8x faster (12m00s -> 4m14s).
- `python3 server/scripts/gala/test_verify.py`: 6 tests OK (one shared call per package + no per-twin calls; drift/write; missing twin; per-twin RUNTIME; per-twin TRANSPILE with diagnostic and remaining packages verified; CLI/log contract).
- `npm run format:markdown` and `npm run format:markdown:check` pass; `npm run fmt:check` passes; codebase-memory index refreshed with `index_repository` (status `indexed`).
- Pre-existing Markdown lint error in `docs/environments.md` (`sentences-per-line/one` on the `VITE_ENABLE_SIGN_IN` bullet) is unchanged from HEAD and outside this task's files.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: dsh
created: 2026-10-27 10:40
---
Profiling evidence gathered 2026-10-27 with GALA 0.93.0, Go 1.26.7, on this checkout:

- `server/.gala/verify.log`: 77 twins discovered/transpiled/ok in 12m22s; gofmt 0.0s, orphan scan 222 files, `go build ./...` 2.7s. Largest package totals: `routes` 182s for 15 twins, `observability` 110s for 8, `postgres` 77s for 16, `main.gala` 106s for 1, `services/gcloud/tasks.gala` 59s for 1, `services/calendar` 46s for 5.
- `GALA_PROFILE=1` on `server/routes/analytics.gala`: 15.06s total; parse 185ms, analyze 14.865s (98.7%), transform+generate 7ms. Analyze phases: sibling discovery 1.712s, GALA package metadata ~10ms (disk-cache hits), Go packages 6.483s, local Go files 6.576s.
- `GALA_PROFILE=1` on `server/main.gala`: 98.6s, of which 91.7s is local-Go-file analysis; every GALA package metadata load was a cache hit.
- Trial run: all 15 `routes` twins in one `gala transpile-package --scan` invocation took 15.5-20s total (first twin 14.5s, later twins 8-143ms each), and every output was byte-identical to the committed twin (`cmp` clean).
- Subprocess counts from `strace -f -e execve`: a single `gala transpile` resolves Go imports by spawning `go list` per import: 225 calls (91 unique) for `routes/analytics.gala`, 249 (109 unique) for `observability/config.gala`, 1,551 (467 unique) for `main.gala`; one `go list` takes 0.04-0.05s. For comparison, `go list -deps -export timeful/server` returns all 716 packages in 0.36s, which is where the compiler-side fix would start and is out of scope here.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The GALA twin verifier now transpiles each source directory through one shared `gala transpile-package --scan --stdlib-module go.gala.fyi/stdlib --inputs ... --outputs ...` invocation instead of starting a fresh `gala transpile` process per twin, removing the repeated compiler cold start. Twin directories each hold one package, so the analyzer's Go type information is now built once per package (26 invocations for 77 twins instead of 77).

### Changes

- `server/scripts/gala/verify.py`
  - `group_sources()` batches the sorted twins into source directories; `check_package()` runs the shared command with cwd set to the package directory and holds the first twin's progress line across it.
  - `check_twin()` keeps the entire per-twin contract and line format: committed-twin presence (`MISSING`), scratch output presence, banned `martianoff/gala` (`RUNTIME`), byte identity (`DRIFT`) or `--write` regeneration, and `gofmt` eligibility.
  - A twin whose shared invocation produced no output is retried alone with `gala transpile`, so its failure keeps its own path and compiler diagnostic while the rest of the package and the remaining packages continue. The shared failure output is kept in the run log; `--verbose` prints shared commands and shared/retried compiler output. All flags, summary counts, log options, and exit codes are unchanged.
- `server/scripts/gala/test_verify.py` (new): stdlib `unittest` regression check with a temporary git fixture and stub `gala`/`go`/`gofmt` that log every invocation. Cases: one shared call per package with zero per-twin calls, per-twin drift + `--write` repair, `MISSING`, per-twin `RUNTIME`, per-twin `TRANSPILE` attribution with the diagnostic and the rest verified, and the CLI/log contract.
- `server/GALA.md`: new "Verifier invocation model" contract subsection with the measured speedup plus a registry note; `server/README.md`: verifier description and regression-check command; `server/scripts/gala/verify.py` docstring and `--help` describe the model.

### Evidence

- Before (HEAD verifier copied to `tmp/task-0367/verify-baseline.py`, same tree): `twins: 77 discovered, 77 transpiled, 77 ok (12m00s)`, total 12m05s.
- After `server/scripts/gala/verify.sh --check`: `77 ok (4m14s)`, total 4m17s, exit 0. `--write`: `77 discovered, 77 transpiled, 77 unchanged (4m49s)`, total 4m52s, exit 0, and no generated file changed. Twins phase is 2.8x faster.
- Machine: `Eyjafjallajokull`, Linux 6.8.0-147-generic x86_64, Intel Core i7-1165G7, 8 CPUs, GALA 0.93.0, Go 1.26.7.
- `python3 server/scripts/gala/test_verify.py`: 6 tests pass. `npm run format:markdown`, `npm run format:markdown:check`, and `npm run fmt:check` pass; codebase-memory index refreshed.

### Risks / follow-ups

- Per-package failures that leave no output are diagnosed by per-twin retries by design; a package-wide compiler change that stops producing shared outputs would degrade toward the old per-twin speed while keeping correct attribution.
- The remaining dominant cost is the single-file `main.gala` and `services/gcloud/tasks.gala` cold start plus the compiler's per-import `go list` resolution (upstream issue #884), which stays out of scope.
- `docs/environments.md` has a pre-existing `sentences-per-line/one` lint error at HEAD, unrelated to this change.
- TASK-0368 (concurrent package groups) can now build on `group_sources()`/`check_package()`.
<!-- SECTION:FINAL_SUMMARY:END -->
