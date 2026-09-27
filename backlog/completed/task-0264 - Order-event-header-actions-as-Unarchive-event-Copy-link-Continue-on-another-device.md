---
id: TASK-0264
title: >-
  Order event header actions as Unarchive event, Copy link, Continue on another
  device
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 13:18'
updated_date: '2026-09-19 13:20'
labels:
  - frontend
dependencies: []
modified_files:
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
priority: medium
type: enhancement
ordinal: 264000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the event page, the metadata action row currently orders the secondary actions Continue on another device, Unarchive event, Copy link (after Edit event when the viewer can edit). The requested reading order puts the lifecycle action before the share action and the cross-device handoff action last: Unarchive event, Copy link, Continue on another device. Edit event remains the leading action. Visibility rules for each action are unchanged: Unarchive event only for archived events the current viewer can manage, Copy link hidden for groups, Continue on another device only for events with a visitor id and id.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 In the event header action row, Unarchive event renders before Copy link, and Copy link renders before Continue on another device
- [x] #2 Edit event remains the first action in the row when the viewer can edit event metadata
- [x] #3 Existing visibility rules are unchanged: Unarchive event only for archived manageable events, Copy link hidden for groups, Continue on another device only when the event has a visitor id and id
- [x] #4 Unit coverage asserts the rendered order of the header action buttons and the frontend required checks pass
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
1. In frontend/src/views/Event.vue reorder the #event-header-button-row children so the rendered order is: Edit event (unchanged, conditional on canEditMetadata), EventOwnerActions (Unarchive event), Copy link, EventAccessTransfer (Continue on another device). Visibility conditions on each action are untouched.
2. Add a unit test in frontend/src/views/Event.test.ts that mounts the event view for an archived, manageable event with the real EventOwnerActions and EventAccessTransfer and asserts the button order inside #event-header-button-row.
3. Run the frontend required checks (lint, fmt:check, typecheck, build, test:unit) and `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Moved `EventOwnerActions` and `#copy-link-btn` ahead of `EventAccessTransfer` in `frontend/src/views/Event.vue`; Edit event stays first and its `canEditMetadata` condition is unchanged.

Added order assertions in `frontend/src/views/Event.test.ts` reusing `scheduleGateStubs` with real `EventOwnerActions` and `EventAccessTransfer`. Archived events hide Edit event via `canEditEventMetadata`, so the archived row shows exactly the three requested actions.

Verification: Event.test.ts 67 passed; full unit suite 1120 passed; lint/fmt:check/typecheck/build green; graphify update run.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Reordered the event page header action buttons to Unarchive event, Copy link, Continue on another device.

## Change
- `frontend/src/views/Event.vue`: `#event-header-button-row` now renders Edit event (unchanged, shown only when metadata editing is allowed), then `EventOwnerActions` (Unarchive event), then `#copy-link-btn` (Copy link), then `EventAccessTransfer` (Continue on another device). Visibility conditions and styling are untouched.

## Tests
- `frontend/src/views/Event.test.ts`: added two DOM order tests. The archived manageable event row renders `["Unarchive event", "Copy link", "Continue on another device"]`; the editable row renders `["Edit event", "Copy link", "Continue on another device"]`.
- Verified: `npx vitest run src/views/Event.test.ts` (67 passed), then frontend required checks — lint (0 errors; 2 pre-existing warnings in `NewSignUp.test.ts`), fmt:check, typecheck, build, and test:unit (149 files / 1120 tests, +2 new) all pass.
- `graphify update .` run. No Markdown, swagger, `scripts/`, `prettier/`, or contract-affecting changes; no e2e coverage required for this DOM order change.
<!-- SECTION:FINAL_SUMMARY:END -->
