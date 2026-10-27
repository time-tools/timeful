---
id: TASK-0367
title: Remove the per-twin compiler cold start from the GALA twin verifier
status: To Do
assignee: []
created_date: '2026-10-27 10:40'
updated_date: '2026-10-27 10:40'
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
- [ ] #1 The verifier pays the compiler cold start at most once per package instead of once per twin (one shared invocation covering several packages also satisfies this), while every twin is still checked independently for missing output, drift, a banned `martianoff/gala` import, and `gofmt` eligibility.
- [ ] #2 Check mode on the current tree reports every committed twin `ok` with zero drift, and `--write` leaves every committed twin byte-identical and reports them unchanged.
- [ ] #3 A deliberately modified `.gala` twin is reported as `DRIFT` in check mode and regenerated only in `--write` mode; a twin whose `.go` file is missing is reported as `MISSING`; a twin that does not transpile is reported individually with its path and compiler diagnostic while the rest of its package and the remaining packages are still verified.
- [ ] #4 The command-line and log contract is unchanged: per-twin progress lines, `--check`, `--write`, `--verbose`, `--log FILE`, `--no-log`, the summary counts, and the exit codes behave as they do today.
- [ ] #5 A regression check that does not need a full real-compiler run covers the shared-invocation behavior and the per-twin failure attribution (a stub compiler executable in a temporary fixture is acceptable).
- [ ] #6 A full `server/scripts/gala/verify.sh` pass on the current tree is green, and the task records the twins-phase wall time before and after the change with the command and machine used.
- [ ] #7 The verifier's help/docstring and the verification contract in `server/GALA.md` describe the shared invocation model and the measured speedup; changed Markdown is formatted.
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
