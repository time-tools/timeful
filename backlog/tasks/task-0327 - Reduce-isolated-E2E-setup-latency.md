---
id: TASK-0327
title: Reduce isolated E2E setup latency
status: Done
assignee:
  - opencode
created_date: '2026-09-23 21:16'
updated_date: '2026-09-23 21:30'
labels:
  - e2e
  - performance
  - docker
dependencies: []
references:
  - e2e/AGENTS.md
  - docs/environments.md
  - compose.test.yaml
  - e2e/isolated-test-stack.ts
documentation:
  - e2e/AGENTS.md
  - docs/environments.md
priority: medium
type: enhancement
ordinal: 317000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Fresh isolated browser E2E runs currently spend about 30 seconds bringing up the test stack, making every verification cycle unnecessarily slow. Reduce that startup latency while retaining the repository's fresh-database isolation and test-only calendar-provider safety guarantees.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A clean isolated E2E stack setup completes materially faster than the current roughly 30-second baseline on the benchmark environment.
- [x] #2 The stack continues to wait for the test-only calendar provider to be healthy before starting the test server, and never enables live provider endpoints.
- [x] #3 Each E2E run still receives a fresh PostgreSQL test database and the existing default teardown behavior remains unchanged.
- [x] #4 Clean-stack and warm-stack setup timings are recorded, and the relevant E2E checks pass.
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
1. Measure the current isolated-stack setup with a no-test Playwright invocation and capture the baseline log.
2. Update the test-only health checks in `compose.test.yaml` so PostgreSQL and `calendar-mock` are polled frequently during startup, while retaining `service_healthy` dependencies and all existing teardown/database-isolation behavior.
3. Validate the rendered Compose configuration and run the identical no-test timing check, then run a focused E2E project to confirm the provider-gated stack still works.
4. Run the repository-required E2E lint, typecheck, formatting, and codebase-memory checks; record clean and warm setup evidence in the task notes and final summary.

Final implementation detail: `postgres-test` keeps its 10-second steady-state interval and adds a 500ms startup interval; `calendar-mock` now explicitly uses a 10-second steady-state interval, 5-second timeout, 5 retries, a 60-second startup period, and a 500ms startup interval. No teardown, database-name, dependency, or provider-override code changed.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Baseline evidence (2026-09-23): a corrected isolated-stack probe recreated the calendar mock and server while retaining the existing PostgreSQL container, then polled `/api/health`; setup completed in 32,738 ms. Temporary services were removed and the existing PostgreSQL container was left untouched. Docker 28.2.2 / Compose 2.36.2 support `start_interval`.

Optimized timing (2026-09-23): direct Compose benchmark against a recreated calendar mock/server measured 4,613 ms clean and 2,910 ms warm. The actual Playwright global setup measured 5,739 ms with the sentinel no-test invocation and 4,615 ms for the focused calendar E2E run. The default `TEST_DB_PERSIST=false` path still performs `down -v`; validation used `true` only to protect the pre-existing volume, and the implementation does not change `e2e/isolated-test-stack.ts`. A separate disposable-project attempt was abandoned after Docker's network address pool was exhausted; shared-network DNS also made that synthetic project invalid, so it was removed without touching the retained project.

Default-path verification: a disposable Compose project with a separate PostgreSQL volume and an explicit unused subnet completed a genuinely fresh setup in 6,399 ms, then `down -v --remove-orphans` removed its containers, network, and volume successfully. The retained `timeful-test` PostgreSQL container/volume remained untouched. This confirms `TEST_DB_PERSIST=false` retains the same fast startup path and only changes teardown.

Validation: `docker compose config --quiet` passed; rendered config reports `start_interval: 500ms` for both `postgres-test` and `calendar-mock`, with `server-test` still requiring `calendar-mock: service_healthy`. `npm run lint`, `npm run typecheck`, and root `npm run fmt:check` passed. E2E `npm run fmt:check` reported pre-existing formatting issues in `specs/schedule-overlap-mobile-touch-firefox.spec.ts`, `specs/timed-event-access-transfer-firefox.spec.ts`, and `specs/timed-event-scheduling-response-selection-firefox.spec.ts`; none are modified by this task, so they were not reformatted.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Reduced isolated E2E startup latency by adding explicit fast startup health polling to the PostgreSQL and test-only `calendar-mock` services in `compose.test.yaml`.
The existing steady-state intervals, `service_healthy` dependency, fresh database generation, provider overrides, and default teardown behavior are unchanged.

## Performance

- Baseline recreated-stack setup: 32,738 ms.
- Optimized recreated-stack setup: 4,613 ms clean and 2,910 ms warm.
- Disposable fresh-volume setup with default-style teardown: 6,399 ms, followed by successful `down -v --remove-orphans` cleanup.
- Actual Playwright global setup: 5,739 ms for the sentinel invocation and 4,615 ms for the focused calendar E2E run.

## Verification

- `docker compose config --quiet` passed; rendered configuration retains `calendar-mock: service_healthy`.
- Focused calendar integration E2E passed: 1 test.
- E2E lint and typecheck passed.
- Root `npm run fmt:check` passed.
- Codebase-memory index refreshed with `codebase-memory-mcp cli index_repository --repo-path .`.
- E2E formatting still reports three pre-existing issues in unrelated spec files; they were not modified.
<!-- SECTION:FINAL_SUMMARY:END -->
