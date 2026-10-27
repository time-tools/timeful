---
id: TASK-0368
title: Run the GALA twin verifier's package groups concurrently with bounded memory
status: Done
assignee:
  - dsh
created_date: '2026-10-27 10:40'
updated_date: '2026-10-27 12:12'
labels:
  - gala
  - tooling
  - performance
dependencies:
  - TASK-0367
references:
  - server/scripts/gala/verify.py
  - server/scripts/gala/verify.sh
  - >-
    backlog/tasks/task-0349 -
    Advance-the-GALA-translation-loop-cursor-and-land-its-runtime-free-twins.md
documentation:
  - server/GALA.md
  - .agents/skills/gala-loop/SKILL.md
priority: medium
type: task
ordinal: 372300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After the per-twin compiler cold start is removed, the verifier still processes packages one at a time. Each compiler process spends most of its wall time waiting on serial Go subprocesses while using well under two cores, so independent packages can overlap on the development host. Wall time still gates every translation iteration, and each compiler process needs several hundred MB, so the concurrency has to be bounded and the per-twin reporting has to stay deterministic. This follows the batching task and preserves every check it establishes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Independent package groups run concurrently under a bounded worker count that the command line can configure, with a default that keeps peak memory within the development machine's budget.
- [x] #2 Per-twin progress lines, the problem list, the summary counts, and the exit code are deterministic and complete regardless of worker count, and no group's report is dropped or interleaved with another's.
- [x] #3 The run log is complete and attributable with concurrency enabled, and `--verbose` output can still be traced to its twin.
- [x] #4 The batching semantics and every check from the predecessor task are unchanged, including `--write` behavior.
- [x] #5 On the current 77-twin tree, the task records wall time and peak memory at the default worker count versus a single worker, with identical check results.
- [x] #6 The new option and its default are documented in `--help`, the verifier docstring, and the `server/GALA.md` verification prose.
- [x] #7 A regression check that does not need a full real-compiler run covers the bounded concurrency and deterministic reporting (a stub compiler executable is acceptable).
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
   - `GroupRecorder` buffers one package group's report events; `GroupResult` owns that group's counts, problems, `gofmt` targets, twin count, and failed flag.
   - `check_group()` and `check_twin()` operate on a `GroupResult` instead of shared verifier state, so groups are independent; `merge()` folds a replayed result into the run totals strictly in sorted group order.
   - `run_groups()` submits groups to a `ThreadPoolExecutor` bounded by `--jobs N` (default `min(4, CPU count)`), yields results in submission order, and prints a dim heartbeat naming the still-running groups every 15 s while it waits.
   - `-j/--jobs` uses a positive-integer parser; `--verbose` commands render as `(cd server/<pkg> && gala ...)` so concurrency keeps every command attributable; all existing flags, counts, and exit codes are unchanged.
2. `server/scripts/gala/test_verify.py` — the stub compiler logs start/end timestamps and pid per invocation and supports per-input delays. Two new cases cover the worker bound (peak overlap) and deterministic reporting/logging across `--jobs 1` vs `--jobs 4` with the first group forced to finish last; `--jobs 0`, negative, and non-integer values are usage errors.
3. Documentation: `--help`, the module docstring, a `server/GALA.md` "Concurrent package groups" subsection with the measured wall time and peak RSS, `server/README.md`, and the `verify.sh` header.

### Evidence

- `python3 server/scripts/gala/test_verify.py`: 8 tests OK.
- Full `--check --jobs 1`: `twins: 77 discovered, 77 transpiled, 77 ok (4m50s)`, total 4m52 s, exit 0, sampled peak tree RSS 780 MiB (gala processes 577 MiB).
- Full `--check --jobs 4`: 77/77/77 `ok` in `2m43s`, total 2m45 s, exit 0, sampled peak tree RSS 1,329 MiB (gala processes 1,262 MiB); normalized reports identical except the log path and heartbeat lines, and the twins phase is 1.8x faster.
- `--write --jobs 4`: `77 discovered, 77 transpiled, 77 unchanged (2m08s)`, total 2m10 s, exit 0, and no generated file changed.
- Log: 77 progress lines in stdout order, every line timestamped, 10 heartbeat lines naming the active groups; `--verbose` commands carry their package directory.
- Machine: Linux 6.8.0-147-generic x86_64, Intel Core i7-1165G7, 8 CPUs, 31 GiB RAM, GALA 0.93.0, Go 1.26.7. Sampling method: a 0.1 s `/proc` sampler summed the verifier tree's `VmRSS`.
- `npm run format:markdown`, `npm run format:markdown:check`, and `npm run fmt:check` pass; `codebase-memory-mcp cli index_repository --repo-path .` reports `status: indexed`.

### Findings outside scope

- The single-twin `main.gala` group is the critical path (~2m under four-way contention), so the default bound buys about 1.8x on this tree rather than 4x; raising the worker count would add memory without shortening that group materially.
- Concurrent compilers share `server/.gala/cache`; its writes are atomic and two full concurrent passes were byte-identical, but a compiler change that breaks that invariant would surface as spurious drift or transpile failures.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Started 2026-10-27. TASK-0367 is Done (shared per-package invocation, 77 twins in 4m14s twins phase); this task builds on `group_sources()`/`check_package()`. Environment: GALA 0.93.0, Go 1.26.7, 8 CPUs, 31 GB RAM, `GALA_HOME` set to the checkout's `.local/gala`.

## Completion 2026-10-27

- Refactored the verifier so a package group's work and report are self-contained (`GroupRecorder`/`GroupResult`) and replayed in sorted group order, then added bounded `ThreadPoolExecutor` execution with `-j/--jobs` (default `min(4, CPU count)`). Existing flags, the shared invocation, the per-twin retry, every check, and `--write` are unchanged.
- Added stub-compiler coverage: start/end timestamps and pids per invocation, per-input delays, a peak-overlap test for the worker bound, a deterministic stdout+log comparison across `--jobs 1` and `--jobs 4` with a forced out-of-order completion, and invalid `--jobs` usage errors. `python3 server/scripts/gala/test_verify.py` -> 8 tests OK.
- Full real-compiler evidence: `--check --jobs 1` -> 77 ok, 4m50s twins, 4m52s total, 780 MiB peak tree RSS; `--check --jobs 4` -> 77 ok, 2m43s twins, 2m45s total, 1,329 MiB peak tree RSS; `--write --jobs 4` -> 77 unchanged (2m08s) and no generated file changed. The normalized reports are identical (only the log path and the timeout heartbeats differ).
- Docs updated in `server/GALA.md` (new subsection with measurements), `server/README.md`, the module docstring, `--help`, and the `verify.sh` header. Markdown and root formatting checks pass; the codebase-memory index was refreshed.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: dsh
created: 2026-10-27 10:40
---
Follows TASK-0367. Batch measurement on `routes` (15 twins, one shared analysis) was 15.5-20s of wall time for the whole package, so after batching the remaining wall time is spread over independent package groups. Memory is the constraint: a single twin compiler process used roughly 0.4-0.5 GiB RSS on this host (8 cores, 31 GB), and the heaviest twins (`main.gala`, `services/gcloud/tasks.gala`) run alone, so the default worker count should be conservative and the reporting must stay deterministic.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The GALA twin verifier now verifies independent source directories concurrently under a bounded worker count while keeping its report deterministic.

### Changes

- `server/scripts/gala/verify.py`
  - `GroupRecorder` buffers one package group's report; `GroupResult` owns that group's counts, problems, twin count, and `gofmt` targets.
  - `check_group()`/`check_twin()` no longer mutate shared verifier state; `merge()` folds finished groups into the run totals and replays each group's report strictly in sorted group order, so per-twin lines, the problem list, summary counts, and the exit code are identical for every `--jobs` value and no group's report interleaves with another's.
  - `run_groups()` uses a `ThreadPoolExecutor` bounded by `-j/--jobs` (default `min(4, CPU count)`) and prints dim 15 s heartbeats naming the still-running groups while it waits.
  - Every existing flag, the shared per-package invocation, the per-twin retry, and the `--write` behavior are unchanged; `--verbose` commands are printed as `(cd server/<pkg> && gala ...)` so they stay attributable with concurrency.
- `server/scripts/gala/test_verify.py`: the stub compiler logs per-invocation start/end timestamps and pids and supports per-input delays. New cases assert the worker bound via peak overlap and identical normalized stdout and log across `--jobs 1` and `--jobs 4` with the first group forced to finish last; `--jobs 0`, `--jobs two`, and `-j -1` are usage errors.
- Documentation: `--help`, the module docstring, `server/GALA.md` ("Concurrent package groups"), `server/README.md`, and the `verify.sh` header.

### Evidence

- `python3 server/scripts/gala/test_verify.py`: 8 tests OK (fixture-only, no real compiler).
- Full `--check --jobs 1`: `twins: 77 discovered, 77 transpiled, 77 ok (4m50s)`, total 4m52 s, exit 0, sampled peak tree RSS 780 MiB.
- Full `--check --jobs 4`: the same 77/77/77 `ok` in `2m43s`, total 2m45 s, exit 0, sampled peak tree RSS 1,329 MiB; the twins phase is 1.8x faster and the normalized reports are identical (only the log path and heartbeat lines differ).
- Full `--write --jobs 4`: `77 discovered, 77 transpiled, 77 unchanged (2m08s)`, total 2m10 s, exit 0; `git status` shows no changed `.gala` or generated `.go` file.
- Run log: 77 progress lines in stdout order, every line timestamped, 10 dim heartbeat lines naming the active groups; `--verbose` commands carry `(cd server/<pkg> && ...)`.
- Machine: Linux 6.8.0-147-generic x86_64, Intel Core i7-1165G7, 8 CPUs, 31 GiB RAM, GALA 0.93.0, Go 1.26.7.
- `npm run format:markdown`, `npm run format:markdown:check`, and `npm run fmt:check` pass; `codebase-memory-mcp cli index_repository --repo-path .` reports `status: indexed`.
- No Swagger annotations or e2e specs changed, so those checks are not applicable; the full `verify.sh` pass above is the integration check.

### Risks / follow-ups

- The single-twin `main.gala` group remains the critical path (~2m under four-way contention), so the default worker count buys about 1.8x on this tree rather than 4x; raising it further would add memory without shortening that group materially.
- Concurrent compiler processes share `server/.gala/cache`; its writes are atomic and concurrently generated twins were byte-identical across two full passes, but a compiler change that breaks that invariant would show up as spurious drift or transpile failures.
- `--jobs` must be at least 1; there is no `0 = auto` spelling.
<!-- SECTION:FINAL_SUMMARY:END -->
