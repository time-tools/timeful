---
id: TASK-0368
title: Run the GALA twin verifier's package groups concurrently with bounded memory
status: To Do
assignee: []
created_date: '2026-10-27 10:40'
updated_date: '2026-10-27 10:40'
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
- [ ] #1 Independent package groups run concurrently under a bounded worker count that the command line can configure, with a default that keeps peak memory within the development machine's budget.
- [ ] #2 Per-twin progress lines, the problem list, the summary counts, and the exit code are deterministic and complete regardless of worker count, and no group's report is dropped or interleaved with another's.
- [ ] #3 The run log is complete and attributable with concurrency enabled, and `--verbose` output can still be traced to its twin.
- [ ] #4 The batching semantics and every check from the predecessor task are unchanged, including `--write` behavior.
- [ ] #5 On the current 77-twin tree, the task records wall time and peak memory at the default worker count versus a single worker, with identical check results.
- [ ] #6 The new option and its default are documented in `--help`, the verifier docstring, and the `server/GALA.md` verification prose.
- [ ] #7 A regression check that does not need a full real-compiler run covers the bounded concurrency and deterministic reporting (a stub compiler executable is acceptable).
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
Follows TASK-0367. Batch measurement on `routes` (15 twins, one shared analysis) was 15.5-20s of wall time for the whole package, so after batching the remaining wall time is spread over independent package groups. Memory is the constraint: a single twin compiler process used roughly 0.4-0.5 GiB RSS on this host (8 cores, 31 GB), and the heaviest twins (`main.gala`, `services/gcloud/tasks.gala`) run alone, so the default worker count should be conservative and the reporting must stay deterministic.
---
<!-- COMMENTS:END -->
