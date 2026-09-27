---
id: TASK-0262
title: Add the unarchive icon to the standalone Unarchive event control
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 12:47'
updated_date: '2026-09-19 12:51'
labels: []
dependencies: []
references:
  - frontend/src/components/event/EventOwnerActions.vue
  - frontend/src/components/event/EventOwnerActions.test.ts
  - frontend/src/components/NewEvent.vue
  - frontend/src/views/Event.vue
  - docs/requirements/functional/fr/FR-115.md
priority: low
type: enhancement
ordinal: 262000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Archived events expose a standalone Unarchive event control in the event page header because the edit-event form is unavailable to them. The edit-event Danger zone lifecycle controls and the neighboring header controls render leading icons, so the standalone Unarchive control reads as visually inconsistent. Align it with that leading-icon pattern without changing its visibility gating, label, or behavior.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The standalone Unarchive event control renders the unarchive icon to the left of the "Unarchive event" label, matching the leading-icon pattern of the existing archive controls.
- [x] #2 The change does not alter the control's label text, visibility gating, busy/disabled behavior, or archive request.
- [x] #3 Unit coverage asserts the leading unarchive icon renders for an archived managed event, and existing Unarchive event label assertions keep passing.
- [x] #4 The frontend required checks pass: lint, fmt:check, typecheck, build, and test:unit.
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

1. `frontend/src/components/event/EventOwnerActions.vue`: import `MdiArchiveArrowUpOutline` from `~icons/mdi/archive-arrow-up-outline` (the icon `NewEvent.vue` already uses for its archived control) and render it as a leading `v-icon` inside the button, with the label wrapped in a `tw:ml-1` span to match the neighboring header controls and the Danger zone rows. Visibility gating, `busy` handling, and the `archiveEvent` call stay unchanged.
2. `frontend/src/components/event/EventOwnerActions.test.ts`: assert the leading icon component renders for an archived managed event, mirroring the `NewEvent.test.ts` icon assertions. Existing label and behavior tests stay as-is.
3. Requirements: iconography is not requirement-scoped; FR-115/FR-124 behavior wording, including the archived-page unarchive clause, is unchanged.
4. Checks: frontend `lint`, `fmt:check`, `typecheck`, `build`, `test:unit`; run the owner-authority Firefox spec because it clicks the `Unarchive event` control by exact accessible name; then `graphify update .`.

## Risks

- The icon must not change the button's accessible name used by e2e (`getByRole("button", { name: "Unarchive event", exact: true })`); the equivalent Archive control already carries an icon under the same exact-name assertion, so risk is low.
- The `v-btn` test stub renders slots, so the icon assertion works with the existing stub setup.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-19: implemented the leading icon in `frontend/src/components/event/EventOwnerActions.vue` (`MdiArchiveArrowUpOutline` from `~icons/mdi/archive-arrow-up-outline` plus a `tw:ml-1` label span, mirroring `NewEvent.vue`). Added an icon-order regression test in `EventOwnerActions.test.ts` with `passThroughStub` for `v-icon`. Checks: frontend lint 0 errors (2 pre-existing NewSignUp.test.ts warnings), fmt:check clean, typecheck pass, build pass, test:unit 149 files / 1117 tests pass. E2E: `e2e/specs/timed-event-owner-authority-firefox.spec.ts` 2/2 pass on the isolated stack, including the exact accessible-name `Unarchive event` journey. `graphify update .` run. No Markdown, swagger, or contract documents changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Gave the standalone `Unarchive event` control on the archived event page the same leading-icon treatment as the other lifecycle and header controls. `EventOwnerActions.vue` now renders `MdiArchiveArrowUpOutline` in a leading `v-icon` with the label in a `tw:ml-1` span, matching `NewEvent.vue`'s archived Danger zone control; visibility gating, `busy` handling, and the `archiveEvent` request are unchanged. `EventOwnerActions.test.ts` gained a regression test asserting the icon component renders before the label, using a pass-through `v-icon` stub so the DOM order is checkable. Requirements were left unchanged because iconography is not requirement-scoped (FR-124 likewise records only behavior). Verification: frontend `lint` (0 errors, 2 pre-existing NewSignUp.test.ts warnings), `fmt:check`, `typecheck`, `build`, and `test:unit` (149 files / 1117 tests) pass; `e2e` Firefox `timed-event-owner-authority-firefox.spec.ts` passes 2/2, including the archive/unarchive journey that clicks the control by exact accessible name, confirming the icon does not alter it; `graphify update .` run.
<!-- SECTION:FINAL_SUMMARY:END -->
