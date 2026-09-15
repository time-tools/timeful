---
id: TASK-0126
title: Enforce FR-118 past-date rules in the event editor date picker
status: To Do
assignee: []
created_date: '2026-09-01 10:34'
updated_date: '2026-09-13 15:22'
labels:
  - date-picker
  - temporal
  - requirements
dependencies:
  - TASK-0125
documentation:
  - docs/requirements/functional/fr/FR-118.md
  - frontend/AGENTS.md
  - docs/terminology/glossary.md
priority: medium
type: enhancement
ordinal: 134300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement [FR-118](docs/requirements/functional/fr/FR-118.md) in the frontend event editor.

Current behavior and starting points:
- `frontend/src/composables/event/useEventEditorState.ts` (minCalendarDate computed, near line 236) sets the minimum only in create mode using the viewer's device date (`Temporal.Now.plainDateISO()`), and applies no restriction in edit mode.
- `frontend/src/components/DatePicker.vue` passes the minimum to Vuetify `v-date-picker` via `:min` (near line 23) and implements custom drag selection (`getSelectableDateFromNode`, `addRemoveDate`).
- Call sites: `frontend/src/components/NewEvent.vue` and `frontend/src/components/NewSignUp.vue` bind `:min-calendar-date` from the composable.
- The editor state already holds the form/event timezone (via `useOwnedTimezone`).

Constraints:
- Derive "current date in a timezone" with `Temporal` from the temporal-polyfill using the selected IANA timezone, in one centralized helper; follow frontend ADR-002/ADR-004/ADR-005 semantics and treat Temporal values with value semantics.
- Do not remove or alter already-selected past dates as a side effect; no destructive stripping, and do not break FR-050/FR-051/FR-052 active-slot behavior on save.
- Frontend-only enforcement; no server or boundary payload changes.

Known verification risks:
- Confirm Vuetify `v-date-picker` with `:min` does not purge model values dated before the minimum; if it does, preserve existing past dates without weakening the selection restriction.
- Confirm the custom drag-selection path cannot pick disabled past cells even though the native Vuetify cells are disabled.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 In event-creation forms (new-event form and sign-up event flow), the date picker does not allow selecting a date before the current date in the form's selected timezone, and changing the form's selected timezone re-derives the earliest selectable date
- [ ] #2 While editing an existing event, Event Picked Dates that fall before the current date in the Event Timezone remain selected in the date picker and are preserved through save
- [ ] #3 While editing, a date before the current date in the Event Timezone cannot be picked by click or by drag-select
- [ ] #4 The current-date-in-a-timezone derivation uses one centralized Temporal-based helper, not ad hoc conversions at call sites
- [ ] #5 Unit tests cover the earliest-selectable-date derivation for create vs edit mode and for timezone-change reactivity
- [ ] #6 An e2e regression spec under e2e/ covers keeping existing past dates while editing
- [ ] #7 npm run lint, npm run typecheck, npm run build, and npm run test:unit pass in frontend/
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-13 15:22
---
2026-09-13: updated paths for the root e2e package move (TASK-0164); repo-tracked browser specs now live under e2e/.
---
<!-- COMMENTS:END -->
