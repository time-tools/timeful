---
id: TASK-0270
title: Restore the active Manage access transfer after an event page reload
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-19 15:57'
updated_date: '2026-09-19 16:02'
labels: []
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - frontend/src/composables/transfer/transferBoundary.ts
priority: high
type: bug
ordinal: 270000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the event page, a source-side access transfer only stays "current" in component memory: after creating a transfer link, `EventAccessTransfer.vue` renders the full transfer form (link, copy, status, matching code, approve, cancel) because `currentId` is set. The transfer identifier is persisted to localStorage, but after a page reload the dialog loads the saved identifiers into tracking and fetches their statuses without promoting any of them to the current transfer. As a result, reopening Manage access after a reload shows only the compact dialog with Create transfer link, and the user cannot see the link they created, check its status, approve the matching code, or cancel the transfer they started.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 After an event-page reload, opening Manage access restores the most recent saved pending transfer as the current transfer: the transfer link, Copy transfer link, transfer status, matching-code input, Approve matching code, and Cancel transfer are visible.
- [x] #2 After an event-page reload, a saved approved transfer is restored as the current transfer with its status and Cancel transfer visible, and without the matching-code input.
- [x] #3 Saved terminal or unavailable transfers are not restored as the current transfer, so the dialog keeps the compact Create transfer link form while still showing revocable grant history.
- [x] #4 When several saved transfers are active, the most recently created one is restored.
- [x] #5 In-session behavior stays unchanged: status polling, identifier pruning, stable grant numbering, and cancellation of the active transfer.
- [x] #6 Frontend unit tests cover pending, approved, terminal, and multiple-active restoration after reload.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add regression-first unit tests in `frontend/src/components/event/EventAccessTransfer.test.ts` that seed `timeful.transfers.EVENT123` and render the component fresh (simulating a reload): pending restore, approved restore, terminal/unavailable non-restore, and most-recent-active selection with multiple saved transfers. Confirm they fail against the current component.
2. Fix `frontend/src/components/event/EventAccessTransfer.vue`: keep resolved statuses during `refresh()`, and when no current transfer is set, promote the highest-numbered tracked transfer whose resolved state is `pending` or `approved` to current (`currentId` + `current`). Do not claim terminal/unavailable transfers.
3. Confirm in-session behavior stays intact (start, approve, cancel, polling, pruning, numbering) via the existing unit tests.
4. Extend the Firefox access-transfer source journey with a reload assertion that the approved transfer controls come back after reload; run the focused spec if the isolated stack is available.
5. Run the required frontend checks, `graphify update .`, and finalize the task with per-criterion evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Regression-first: added four unit tests seeded from `timeful.transfers.EVENT123` against a freshly mounted component (reload simulation); the pending, approved, and most-recent-active tests failed before the fix (3 failed / 10 passed), confirming the defect.

Fix scope stayed in `EventAccessTransfer.vue`: `refresh()` now records resolved statuses in a map, and when no current transfer is set it promotes the highest-numbered tracked transfer resolved as `pending` or `approved` to current (`currentId` + `current`). Terminal/unavailable states are never claimed, so the compact create form and revocable grant history remain unchanged.

Selection happens after all status requests settle, so it is deterministic under `Promise.all` and retries on later polls if an initial status request fails transiently.

Verified frontend checks: lint (0 errors; 2 pre-existing unrelated warnings), fmt:check, typecheck, build, and test:unit (149 files / 1130 tests). E2E package lint/fmt:check/typecheck passed.

Extended the existing Firefox access-transfer cancel journey with a real source `page.reload()` restoration step; the isolated `firefox-desktop` run passed 8/8 (2.1m). No requirement change needed: FR-117 already requires source-side cancellation of a pending transfer, which reload previously made unreachable.

No Markdown files changed; managed Backlog Markdown is MCP-authored and excluded from `npm run format:markdown`. No swagger or root `scripts/`/`prettier/` changes. `graphify update .` completed (6032 nodes / 11422 edges).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary
- Reloading the event page no longer collapses Manage access to the compact Create transfer link form while a saved transfer is still active.
- `EventAccessTransfer.vue` now records the statuses resolved during `refresh()` and, when no current transfer is in memory, restores the most recently created saved transfer whose resolved state is `pending` or `approved`, rebuilding the full form (transfer link, copy, friendly status, matching-code input, approve, cancel).
- Terminal or unavailable saved transfers are deliberately not promoted, so the compact form and revocable grant history behave as before.
- In-session behavior is untouched: status polling, pruning, stable grant numbering, and cancellation still use the same paths.

## Verification
- New unit regressions cover pending restore, approved restore, terminal non-restore, and most-recent-active selection; all 13 focused transfer tests pass and the full frontend suite passes (149 files / 1130 tests).
- Required frontend checks pass: lint, fmt:check, typecheck, build; e2e package lint, fmt:check, and typecheck also pass.
- The Firefox `timed-event-access-transfer-firefox.spec.ts` journey now reloads the source page after approval and asserts the restored controls before cancelling; the isolated run passed 8/8 in 2.1 minutes.
- `graphify update .` completed. No swagger, Markdown, or contract-document changes were required.

## Risks
- Restoration happens once the status request settles, so the full form appears after the brief status fetch rather than instantly; status polling retries claiming if that first request fails transiently.
<!-- SECTION:FINAL_SUMMARY:END -->
