---
id: TASK-0260
title: Move event archive and delete actions into the Edit event form Danger zone
status: Done
assignee:
  - opencode
created_date: '2026-09-18 15:54'
updated_date: '2026-09-18 16:37'
labels: []
dependencies: []
references:
  - frontend/src/components/event/EventOwnerActions.vue
  - frontend/src/components/NewEvent.vue
  - frontend/src/components/NewDialog.vue
  - frontend/src/views/Event.vue
  - frontend/src/components/EventItem.vue
  - frontend/src/utils/services/EventService.ts
  - e2e/specs/timed-event-owner-authority-firefox.spec.ts
  - docs/requirements/functional/fr/FR-115.md
  - docs/requirements/README.md
priority: medium
type: enhancement
ordinal: 260000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The event page header mixes routine actions (Edit, Copy link, access transfer) with destructive lifecycle actions (Archive, Delete), so destructive controls sit next to frequently used ones. Moving them into the Edit event form groups them where owners already manage the event and reduces accidental header clicks.

Confirmed decisions: only the event page moves these actions; the dashboard event list keeps its archive/unarchive and delete actions; a standalone Unarchive action stays on the event page for archived events because archived events hide the Edit event button; archiving or unarchiving from the form closes the form and refreshes the event page; delete stays confirmation-gated and lands on the home page.

Requirements capture is part of this task: add FR-124 for the Danger zone placement and interaction, and extend FR-115 so the archived event page keeps an Unarchive entry point outside the unavailable Event Settings editing form.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/requirements/functional/fr/FR-124.md` exists with README-conformant front matter (`components: frontend`, `status: proposed`) and states: in the event page Edit event form, the Event Owner sees archive/unarchive and delete controls under a `Danger zone` on separate rows; archiving or unarchiving closes the form; delete requires confirmation; the dashboard event list is out of scope.
- [x] #2 FR-115 gains a clause that an Archived Event's page exposes Unarchive outside the unavailable Event Settings editing form; the requirements README index gains an FR-124 row with a stable link; each Markdown sentence stays on one physical source line.
- [x] #3 In edit mode, the Edit event form shows the Danger zone with Archive/Unarchive event and Delete event on separate rows; create mode shows no Danger zone.
- [x] #4 The Danger zone appears only under the same ownership conditions that gate today's event-page Archive and Delete actions.
- [x] #5 Archive and Unarchive use the existing archive endpoint; after either action the form closes and the event page reflects the new state.
- [x] #6 Delete opens the existing confirmation copy; confirming deletes the event and navigates to the home page, while cancelling changes nothing.
- [x] #7 The event page header no longer renders Archive/Unarchive or Delete actions; a standalone Unarchive stays available on the event page while the event is archived and still works after reload.
- [x] #8 Dashboard event list archive/unarchive and delete actions are unchanged.
- [x] #9 Unit regression coverage: Danger zone rendering and visibility, archive and unarchive flows, delete confirm and cancel, archived-event standalone Unarchive, and absence of header Archive/Delete; affected `NewEvent.test.ts` and `Event.test.ts` expectations are updated.
- [x] #10 E2E: `e2e/specs/timed-event-owner-authority-firefox.spec.ts` is updated so the owner journeys archive, unarchive, and delete through the form while visitor denial assertions still hold; the affected Firefox spec is run.
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
## Approach

Preserve today's permission gate (`event.eventVisitorId && event.canManageEvent`) for every moved/retained lifecycle action, and keep the dashboard event list (`EventItem.vue`) untouched.

1. `frontend/src/composables/event/eventOwnership.ts`: add `canManageEventAsCurrentViewer(event)` wrapping the existing `eventVisitorId && canManageEvent` gate so the event-page action and the form Danger zone cannot drift.
2. `frontend/src/components/event/EventOwnerActions.vue`: slim to a standalone `Unarchive event` button shown only while the event is archived and the viewer can manage it; remove the active-archive path, delete dialog, `_delete` use, and `deleted` emit. Keep `changed` refresh.
3. `frontend/src/views/Event.vue`: drop the now-unused `@deleted` listener; keep mounting `EventOwnerActions` in the header.
4. `frontend/src/components/NewEvent.vue`: in edit mode only, render a `Danger zone` after Advanced options with `Archive`/`Unarchive event` and `Delete event` on separate full-width rows. Archive/unarchive reuse `POST /events/{id}/archive`; on success emit `refresh-event` so `NewDialog` closes the form and `Event.vue` refreshes the page. Delete reuses the existing confirmation copy, then `DELETE /events/{id}`, closes the dialog, and pushes `/`. Errors reuse the existing `showError` copy.
5. Requirements: add `docs/requirements/functional/fr/FR-124.md` (frontend, proposed) for the Danger zone placement and interaction; extend `FR-115` with the archived-event page Unarchive clause; add the FR-124 index row.
6. Unit tests: `eventOwnership.test.ts` for the new helper; `NewEvent.test.ts` for Danger zone rendering/visibility, archive, unarchive, and delete confirm/cancel; `Event.test.ts` for header absence and the archived standalone Unarchive.
7. E2E: update `timed-event-owner-authority-firefox.spec.ts` so archive/delete happen through the form and the visitor denial stays; update the `mode === "owner"` assertion in `timed-event-access-transfer-firefox.spec.ts` to see `Archive event` inside the opened form. Run both Firefox specs on the isolated stack.
8. Checks: `npm run lint`, `fmt:check`, `typecheck`, `build`, `test:unit` in `frontend/`; `npm run format:markdown` for changed Markdown; `graphify update .` after code changes.

## Risks

- Nested `v-dialog` inside the edit dialog must still render and gate on `loading`.
- The edit form closes on archive; ensure `refresh-event` here does not carry save-specific payload that would confuse `Event.vue`.
- Archived events keep `canManageEvent true` and `canEditSettings false` (server `owner_test.go`), which is what makes the standalone Unarchive reachable.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-18 progress: implementation complete; verification partially run. Next session finishes e2e and finalization.

Implemented: added `canManageEventAsCurrentViewer(event)` in `frontend/src/composables/event/eventOwnership.ts` (wraps `eventVisitorId && canManageEvent`) and used it in both lifecycle surfaces. `EventOwnerActions.vue` now renders only a standalone `Unarchive event` button for archived events the viewer can manage; its delete dialog, active-archive path, `_delete` use, and `deleted` emit are gone. `Event.vue` dropped the `@deleted` listener. `NewEvent.vue` gained a `Danger zone` after Advanced options (edit mode plus managed event only) with `Archive`/`Unarchive event` and `Delete event` on separate full-width rows, plus the moved `Delete event?` confirmation dialog; archive/unarchive `POST /events/{id}/archive` then emits `refresh-event` to close the form and refresh the page, delete `DELETE /events/{id}` then closes and routes to `/`, and both reuse the previous `showError` copy.

Requirements: added `docs/requirements/functional/fr/FR-124.md` (frontend, proposed), extended FR-115 with the archived-page Unarchive clause, and added the FR-124 row to `docs/requirements/README.md`.

Tests: new `frontend/src/components/event/EventOwnerActions.test.ts`; Danger zone render/visibility/archive/unarchive/delete-confirm coverage in `NewEvent.test.ts`; helper coverage in `eventOwnership.test.ts`; header-absence source assertion in `Event.test.ts`; both Firefox e2e specs updated (owner-authority journeys through the form; access-transfer opens the form before asserting `Archive event`).

Checks passed: `npm run test:unit` (149 files, 1115 tests), `npm run lint` (0 errors; 2 pre-existing NewSignUp.test.ts warnings), `npm run typecheck`, `npm run fmt:check`, `npm run build`. Markdown reformatted with root `npm run format:markdown`.

Remaining: run the isolated-stack e2e specs `cd e2e && npm run test:e2e -- --project=firefox-desktop specs/timed-event-owner-authority-firefox.spec.ts specs/timed-event-access-transfer-firefox.spec.ts` (external Go cache volumes already created; stream full output, never pipe through tail/head). Then run `graphify update .`, verify ACs #1-#10 with evidence, check DoD items (swagger and root fmt:check are N/A), write the final summary, and mark the task Done. Do not archive to `backlog/completed/`.

2026-09-18 review/finalization: supersedes the earlier notes. Archive/unarchive now uses the `archiveEvent` service; delete emits a dedicated `deleted` event that `NewDialog` closes via `exitDialog`, so it bypasses the unsaved-changes gate; FR-115 now names the Event Owner and no longer calls the archived form read-only; the `Event.test.ts` header-absence check is a rendered assertion instead of a source grep; the unit total is 149 files / 1116 tests. E2E run: owner-authority 2/2 at default workers; access-transfer 8/8 at `--workers=1`.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-18 16:36
---
2026-09-18 review (opencode): reviewed the implementation against the current codebase, applied the agreed fixes, and verified them.

Fixes applied:
- Delete now emits a dedicated `deleted` event from `NewEvent`; `NewDialog` handles it with `exitDialog`, so confirming delete no longer routes through `handleDialogInput` and cannot raise the unsaved-changes dialog on the way to `/`.
- `NewEvent` archive/unarchive now calls the `archiveEvent` service instead of a raw `post`.
- FR-115's archived-page clause now names the Event Owner and states unarchive happens without opening the Event Settings editing form.
- `Event.test.ts` header-absence coverage is now a rendered assertion (mounts the event view with `EventOwnerActions` unstubbed, asserts no Archive/Delete for an active event and Unarchive for an archived one) instead of a source grep.
- Added a `NewDialog` regression test for the deletion close path and updated `NewEvent` archive/delete expectations.

Verification:
- `frontend` checks: `test:unit` 149 files / 1116 tests pass; `lint` 0 errors (2 pre-existing NewSignUp.test.ts warnings); `fmt:check`, `typecheck`, `build` pass.
- Markdown: root `format:markdown` clean; `graphify update .` run.
- E2E firefox-desktop: `timed-event-owner-authority-firefox.spec.ts` 2/2 pass at default workers, including archive/unarchive/delete through the form and the visitor denial assertions; `timed-event-access-transfer-firefox.spec.ts` 8/8 pass with `--workers=1` (the default 2-worker run hit the documented heavy-spec timeout flake in the new dialog-open step and in two unrelated tests).

Remaining: AC #1-#10 are still unchecked and the task stays In Progress; finalization (AC evidence, final summary, Done) has not been done. Changes are staged but not committed.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Moved the event-page archive/unarchive and delete lifecycle actions from the header into a Danger zone in the edit-event form. `canManageEventAsCurrentViewer` in `eventOwnership.ts` wraps the existing `eventVisitorId && canManageEvent` gate and is used by both the form Danger zone (`NewEvent.vue`) and the retained standalone header Unarchive (`EventOwnerActions.vue`). Archive/unarchive calls the `archiveEvent` service, closes the form via `refresh-event`, and refreshes the event page; delete keeps the existing confirmation copy, then `DELETE /events/{id}` and routes home through a dedicated `deleted` event that `NewDialog` closes via `exitDialog`, bypassing the unsaved-changes prompt. The event-page header no longer renders Archive/Delete; the dashboard event list is untouched. Requirements: added FR-124 and extended FR-115 with the archived-page unarchive clause, plus the README index row. Verification: frontend lint (0 errors), fmt:check, typecheck, build, and test:unit (149 files / 1116 tests) pass; Markdown format check passes; `graphify update .` run. E2E firefox-desktop: owner-authority 2/2 at default workers, covering archive/unarchive/delete through the form and the visitor denial; access-transfer 8/8 with `--workers=1` (the default 2-worker run hit the documented heavy-spec timeout flake).
<!-- SECTION:FINAL_SUMMARY:END -->
