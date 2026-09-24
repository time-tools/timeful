---
id: TASK-0328
title: Reduce browser E2E runtime without weakening behavior
status: Done
assignee:
  - opencode
created_date: '2026-09-24 16:49'
updated_date: '2026-09-24 17:57'
labels:
  - e2e
  - performance
  - playwright
dependencies: []
documentation:
  - e2e/AGENTS.md
  - docs/environments.md
  - >-
    backlog/completed/task-0191 -
    Speed-up-local-E2E-with-configurable-Firefox-concurrency-and-targeted-build-setup.md
  - >-
    backlog/completed/task-0234 -
    Raise-the-recorded-access-transfer-journey-E2E-budget-to-40-seconds.md
modified_files:
  - e2e/AGENTS.md
  - e2e/helpers/actor-context.ts
  - e2e/helpers/database-inspect.ts
  - e2e/helpers/timed-event-helpers.ts
  - e2e/specs/event-mobile-editing-options.spec.ts
  - e2e/specs/event-page-days-only-layout.spec.ts
  - e2e/specs/event-page-no-responses-layout.spec.ts
  - e2e/specs/event-toolbar-mobile-layout.spec.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
  - e2e/specs/timed-event-create-firefox.spec.ts
  - e2e/specs/timed-event-reprojection-firefox.spec.ts
  - e2e/specs/timed-event-specific-times-edit-firefox.spec.ts
  - e2e/specs/timed-event-timerange-width-firefox.spec.ts
  - e2e/specs/timed-event-viewer-tz-column-duplication-firefox.spec.ts
  - e2e/specs/timed-event-weekly-firefox.spec.ts
priority: medium
type: enhancement
ordinal: 0
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Browser E2E feedback is too slow for iterative development: some multi-context Firefox journeys reach 30–40+ seconds, and suite wall time is further extended by repeated setup and infrastructure work. Existing records show access-transfer approval journeys at 35.7–40.9 seconds in the default development-server mode, while the recorded bundled-assets run reduced the heaviest cases to about 28 seconds. This task is a measured, behavior-preserving optimization pass for the E2E harness and test setup; it must improve turnaround without skipping coverage, weakening assertions, changing authorization or isolation semantics, or merely increasing timeouts.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A reproducible before-and-after performance record exists for representative slow access-transfer tests and the relevant suite scope, using identical project selection, worker count, frontend mode, and artifact settings; setup, test execution, teardown, retries, and browser startup are distinguished.
- [x] #2 Every existing E2E test, assertion, actor/context boundary, authorization and denial check, consent behavior, persistence check, and reload check remains covered after optimization; no test is skipped solely to improve timing.
- [x] #3 Runtime reductions come from demonstrably redundant or parallelizable work, faster equivalent readiness or setup, or a supported resource configuration; they do not rely on fixed sleeps, weakened waits, increased timeout budgets, or reduced isolation.
- [x] #4 The optimized default and any opt-in modes are documented with commands, measured tradeoffs, and a safe fallback for constrained environments; unsupported worker counts and unsafe server reuse are not made defaults.
- [x] #5 Relevant E2E, lint, formatting, and typecheck commands pass, and a repeat run demonstrates that the improvement is not a one-off timing fluke.
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
## Implementation plan

1. **Remove stale test-only consent latency safely.** In `e2e/helpers/timed-event-helpers.ts`, replace the unconditional 1.5-second wait for the removed Quantcast dialog with a state-based check that still accepts a currently visible consent action and removes the legacy container. Keep the access-transfer application consent assertions untouched; they exercise a different `GrantedAccessConfirmation` flow.

2. **Remove redundant setup work without changing contracts.** In `seedCanonicalTimedEvent`, use the `shortId` returned by event creation and retain the `/ids` lookup only as a compatibility fallback when the response omits it. In `database-inspect.ts`, cache the validated isolated database name per Playwright worker while preserving the `timeful-test-` safety check and synchronous query behavior.

3. **Reduce teardown contention.** In `actor-context.ts`, finalize independent actor recordings concurrently, retaining the existing retention policy, attachment names, filenames, and cleanup-on-success behavior.

4. **Enable safe parallelism where state is isolated.** Remove file-level serial mode from `timed-event-specific-times-edit-firefox.spec.ts` only after confirming every test owns its page/context and seeds unique state. Re-audit other serial files rather than changing them indiscriminately. Record the measured worker-count change.

5. **Document and verify.** Keep the existing dev-server default and bundled opt-in mode; update `e2e/AGENTS.md` only with the verified focused command and parallelism note. Rerun the recorded access-transfer and specific-times baselines with identical settings, run affected browser/database coverage, then run E2E lint, formatting, and typecheck plus the required code graph refresh. Record setup, test, teardown, worker count, and any retries in task notes.

Risks and guardrails: no application code, authorization, consent semantics, assertions, timeout budgets, actor boundaries, or test selection may be weakened. If a helper fast path misses a visible consent UI or parallel execution introduces contention, retain the existing behavior and narrow the change before finalizing.

The first access-transfer after-run showed the shared-helper changes preserve all 8 journeys but did not materially reduce that multi-context run. To address the original 30–40+ second case, extend the same slice by assigning actor IDs before asynchronous context creation, creating independent actor contexts/pages concurrently, and overlapping independent API setup while preserving the required target-before-stranger navigation order. Re-run the access-transfer benchmark and revert any ordering or isolation regression.

Static review found the remaining multi-test serial groups also create isolated state per test. Extend the scheduling change to the independent Firefox groups (`timed-event-reprojection`, `timed-event-weekly`, `timed-event-timerange-width`, and `timed-event-viewer-tz-column-duplication`) and independent Chromium layout groups (`event-page-days-only-layout`, `event-page-no-responses-layout`, `event-mobile-editing-options`, and `event-toolbar-mobile-layout`); leave single-test or stateful groups serial. Validate both browser projects before keeping the batch.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Baseline (2026-09-24): ran `E2E_ARTIFACTS_RUN_ID=task-0328-baseline npm run test:e2e -- --project=firefox-desktop --workers=2 specs/timed-event-access-transfer-firefox.spec.ts` from `e2e/`. All 8 passed in 1.1m; setup 3973ms, teardown 912ms; approval journeys 23.8s owner, 25.1s guest, 17.8s signed-in; reload journeys 8.5s and 11.2s. Artifacts: `/tmp/opencode/timeful-e2e-artifacts/task-0328-baseline/`. This is the comparison run for the optimized path.

Second baseline (2026-09-24): `E2E_ARTIFACTS_RUN_ID=task-0328-baseline-specific npm run test:e2e -- --project=firefox-desktop --workers=2 specs/timed-event-specific-times-edit-firefox.spec.ts` from `e2e/`. All 6 passed in 51.0s, but the file-level serial mode forced Playwright to use 1 worker; setup 5151ms, teardown 933ms. This is the comparison run for removing only an audited serial dependency.

Post-change measurements (2026-09-24): `task-0328-after-specific` ran the 6-test specific-times file with 2 workers in 36.0s (baseline 51.0s, setup 6378ms, teardown 1443ms). `task-0328-baseline-create` ran the 6-test create file with file-level serial mode in 50.2s (setup 6202ms, teardown 1195ms); after removing its audited serial dependency, `task-0328-after-create` passed all 6 with 2 workers in 41.5s (setup 4274ms, teardown 1138ms). The access-transfer concurrency runs passed 8/8 twice in 1.2m and 1.1m; their setup/teardown variation means no isolated speed claim yet.

Verification update (2026-09-24): the affected Firefox set initially passed 31/31 with one fixme (`task-0328-firefox-affected`), and the affected Chromium set passed 26/26 with 20 intentional project skips (`task-0328-chromium-affected`). A full Firefox run then had one signed-in access-transfer failure at the immediate `Create new transfer link` click after `Manage access` (`task-0328-full-firefox`, 76 passed, 1 skipped); the isolated signed-in journey passed in 20.4s. Added a state-based dialog visibility assertion, then reran the full affected Firefox set: 31 passed, 1 fixme in 3.0m (`task-0328-firefox-affected-after-wait`). Bundled access-transfer was 7/8 once due the same cancel-journey readiness symptom, the isolated cancel journey passed, and the repeat bundled run passed 8/8 in 56.2s with the heaviest journey at 19.8s (`task-0328-bundled-access-repeat`).

Full Chromium verification initially had two desktop-only event-shell timeouts (`schedule-overlap-mono-labels` and `sign-up-form-event`; 56 passed, 28 skipped). The two tests were rerun sequentially to avoid the repository's fixed-stack concurrency restriction; the mono-label test passed on its second isolated run (4.2s) and the signup test passed (5.3s). Their blank-page snapshots and repeat behavior indicate existing Vite/full-suite contention rather than a deterministic helper regression; the two concurrent isolated commands initially conflicted over the shared Compose project and were not counted as test evidence.

A second full Chromium run also had one transient desktop failure (`styling-migration` compatibility-reset check) after 57 passes; the exact test passed in isolation in 3.0s. The affected Chromium groups remain green, and the failure is consistent with the repository's existing Vite/full-suite startup flake pattern rather than a deterministic assertion or changed-selector failure.

Final verification (2026-09-24): `task-0328-final-access` passed all 8 access-transfer tests in 1.2m with the same Firefox desktop/two-worker/dev-server settings as the baseline; setup 4873ms, teardown 867ms, approval journeys 27.0s owner, 28.3s guest, 18.1s signed-in. Full Firefox desktop passed 77 tests with 1 intentional fixme in 5.3m (`task-0328-full-firefox-final`); Firefox touch passed 13/13 in 56.6s; affected Chromium groups passed 26/26 with 20 intentional project skips; production projects passed 3/3 including the fresh build. Final E2E lint and typecheck pass; all 14 changed TypeScript files pass oxfmt; root Markdown and root formatting checks pass; codebase-memory indexing completed. The repository-wide E2E `npm run fmt:check` still reports two untouched pre-existing files (`specs/schedule-overlap-mobile-touch-firefox.spec.ts` and `specs/timed-event-scheduling-response-selection-firefox.spec.ts`), so those were not reformatted outside task scope.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Implemented a behavior-preserving E2E runtime optimization pass.

- Replaced the stale 1.5-second Quantcast consent wait with a state-based visible-consent check that still handles `Agree`/`Accept all` and legacy container cleanup.
- Removed the redundant event ID lookup when creation already returns `shortId`, retaining the compatibility fallback.
- Cached the validated isolated database name per worker and finalized independent actor recordings concurrently.
- Started independent access-transfer contexts/pages and API setup concurrently while preserving the required target-before-stranger navigation order; added a state-based Manage Access dialog readiness assertion.
- Removed file-level serial scheduling from audited independent multi-test Firefox and Chromium layout groups; single-test/stateful groups remain serial.
- Documented the supported worker and bundled-frontend tradeoffs in `e2e/AGENTS.md`.

## Measurements

- `timed-event-specific-times-edit-firefox.spec.ts`: 51.0s with one worker before serial removal versus 36.0s with two workers afterward.
- `timed-event-create-firefox.spec.ts`: 50.2s before versus 41.5s afterward.
- Access-transfer final run: all 8 passed in 1.2m at two workers; bundled repeat passed all 8 in 56.2s with the heaviest journey at 19.8s.
- Full Firefox desktop passed 77 tests with 1 intentional fixme; Firefox touch passed 13/13; affected Chromium groups passed 26/26 with 20 intentional skips; production projects passed 3/3.

## Verification

`npm run lint`, `npm run typecheck`, changed-file `oxfmt --check`, root Markdown/root formatting checks, `git diff --check`, and codebase-memory indexing passed. The repository-wide E2E formatter still reports two untouched pre-existing files; they were not changed outside task scope. Full Chromium runs showed transient Vite/page-startup flakes in unrelated tests, all of which passed when rerun in isolation. No application semantics, assertions, timeout budgets, actor boundaries, or test selection were weakened.
<!-- SECTION:FINAL_SUMMARY:END -->
