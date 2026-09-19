---
id: TASK-0267
title: Move the Unarchive event action into the archived read-only banner
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 14:47'
updated_date: '2026-09-19 15:09'
labels:
  - frontend
dependencies: []
modified_files:
  - frontend/src/views/Event.vue
  - frontend/src/components/event/EventOwnerActions.vue
  - frontend/src/components/event/EventOwnerActions.test.ts
  - frontend/src/views/Event.test.ts
  - e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts
priority: medium
type: enhancement
ordinal: 267000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the event page, the archived read-only banner ("This event is archived and read-only.") and the Unarchive event action are currently separate: the banner renders when event.eventVisitorId && event.isArchived, and the green outlined Unarchive event button renders in the header action row when the viewer can manage the archived event. Group the archive state and its remedy so a user who learns the event is read-only sees the action in the same place.

Move the Unarchive event control into the banner, and change its treatment from the header's green outlined style to a solid green primary action (tw:bg-green tw:text-white) so it reads as the banner's call to action. Remove the standalone control from the event header action row. Keep EventOwnerActions as the single implementation of the unarchive action and keep its capability gating unchanged.

Layout: do not use the v-alert #append slot; Vuetify's alert grid keeps append beside the text at all widths, which squeezes the sentence into a narrow column on phones. Render the action inside the alert's default content with a responsive flex wrapper so it stacks below the sentence on phones and sits to the right of it from the sm breakpoint up.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Unarchive event control renders inside the archived read-only banner for an archived event the current viewer can manage
- [x] #2 The Unarchive event control no longer renders in the event header action row (#event-header-button-row)
- [x] #3 The banner keeps its existing text "This event is archived and read-only.", its event.eventVisitorId && event.isArchived visibility condition, and its tonal info styling
- [x] #4 A viewer who cannot manage the archived event sees the banner without any unarchive action
- [x] #5 The Unarchive event control keeps the archive-up icon and its behavior: archiveEvent(id, false), disabled while the request is in flight, store error message on failure, and a changed emit on success
- [x] #6 At phone widths the action stacks below the banner sentence inside the strip, and at the sm breakpoint and wider it sits to the right of the sentence
- [x] #7 Unit coverage asserts action placement in the banner and capability gating, and the required frontend checks pass
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
1. Move the Unarchive event control into the archived read-only banner in frontend/src/views/Event.vue, inside the alert's default content with a responsive flex wrapper; remove it from #event-header-button-row.
2. Change EventOwnerActions.vue to the solid green primary treatment (tw:bg-green tw:text-white) while keeping the archive-up icon, capability gating, busy state, error handling, and changed emit unchanged.
3. Update frontend/src/components/event/EventOwnerActions.test.ts treatment coverage and frontend/src/views/Event.test.ts placement, order, and capability coverage.
4. Run frontend lint, fmt:check, typecheck, build, and test:unit; run the focused Firefox e2e specs for the archived banner and owner authority; run graphify update .
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Moved `EventOwnerActions` from `#event-header-button-row` into the archived `v-alert` default content in `frontend/src/views/Event.vue`, wrapped with a responsive flex layout: column/items-start on phones, row/items-center/justify-between from `sm`. Chose the default-slot layout over the alert's `#append` slot because Vuetify's alert grid keeps append beside the text at every width and would squeeze the sentence on phones.

Changed `EventOwnerActions.vue` to the solid primary treatment (`variant="flat"`, `tw:bg-green tw:text-white`) and removed the per-element `tw:text-green` classes; the `v-icon` inherits white through `currentColor`.

Test updates: `Event.test.ts` now asserts the action lives inside the `v-alert` element and not in the header row, plus a new capability test; `EventOwnerActions.test.ts` asserts the solid treatment. The archived e2e spec gained a placement test using bounding boxes and `expect.poll` (beside at desktop, below at 375x900).

Verification: focused vitest 74 passed; full unit suite 1125 passed; lint/fmt:check/typecheck/build green; focused `firefox-desktop` e2e 4 passed; e2e lint/fmt:check/typecheck green; `graphify update .` run.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @opencode
created: 2026-09-19 15:09
---
Follow-up requested after completion: the banner action now stacks below the sentence at all widths, not only on phones. `frontend/src/views/Event.vue` uses `tw:flex tw:flex-col tw:items-start tw:gap-2` with no `sm:` row modifiers, and the e2e placement test now asserts the button is below the sentence at both the desktop viewport and 375x900. This supersedes the desktop beside-placement in AC #6. Frontend lint/fmt:check/typecheck/build and 1125 unit tests pass; the focused archived-banner firefox-desktop spec passes (2 tests); e2e lint/fmt:check/typecheck pass; `graphify update .` run.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Change
- `frontend/src/views/Event.vue`: the archived read-only `v-alert` now renders "This event is archived and read-only." plus `EventOwnerActions` inside a responsive flex wrapper (`tw:flex tw:flex-col tw:items-start tw:gap-2 tw:sm:flex-row tw:sm:items-center tw:sm:justify-between`). The standalone `EventOwnerActions` was removed from `#event-header-button-row`. Banner text, `event.eventVisitorId && event.isArchived` visibility, and tonal info styling are unchanged.
- `frontend/src/components/event/EventOwnerActions.vue`: the unarchive control changed from the green outlined header treatment to a solid primary action (`variant="flat"`, `tw:bg-green tw:text-white`). The archive-up icon, capability gating, busy disable, store error handling, and `changed` emit are unchanged. This intentionally supersedes TASK-0265's outlined treatment for this control.

## Tests
- `frontend/src/views/Event.test.ts`: updated the lifecycle/header test (Unarchive is inside `v-alert`, not `#event-header-button-row`), added a placement/capability test (manager sees the action in the banner; a viewer without `canManageEvent` sees the sentence with no action), updated the archived order test to `["Copy link", "Manage access"]` with the banner holding `["Unarchive event"]`, and updated the treatment test to solid green for the banner action while Manage access stays outlined.
- `frontend/src/components/event/EventOwnerActions.test.ts`: treatment test now asserts `variant="flat"` plus `tw:bg-green tw:text-white`.
- `e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts`: added a placement test asserting the action is beside the sentence at the desktop viewport and below it at 375x900.
- Verified: focused unit run 74 passed; full `test:unit` 149 files / 1125 tests passed; frontend lint (0 errors; 2 pre-existing `vue/one-component-per-file` warnings in `NewSignUp.test.ts`), fmt:check, typecheck, build green; e2e `firefox-desktop` focused specs 4 passed; e2e lint/fmt:check/typecheck green.
- `graphify update .` run.
- No swagger, `scripts/`, `prettier/`, or contract-affecting changes; no hand-authored Markdown changed (only the Backlog-generated task record), so DoD items #4, #5, #7, #8 are not applicable.
<!-- SECTION:FINAL_SUMMARY:END -->
