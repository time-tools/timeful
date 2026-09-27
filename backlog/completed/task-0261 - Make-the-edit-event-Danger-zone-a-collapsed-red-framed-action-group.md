---
id: TASK-0261
title: Make the edit-event Danger zone a collapsed red-framed action group
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 12:28'
updated_date: '2026-09-19 12:35'
labels: []
dependencies: []
references:
  - frontend/src/components/NewEvent.vue
  - docs/requirements/functional/fr/FR-124.md
  - e2e/specs/timed-event-owner-authority-firefox.spec.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
priority: low
type: enhancement
ordinal: 261000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In the event page Edit event form, make the Danger zone a collapsible section that starts collapsed. When expanded, present Archive/Unarchive event and Delete event as full-width red-outlined buttons, each on its own row inside a single red rounded-corner frame with no divider lines between rows, and with leading icons (archive for Archive/Unarchive, trash can for Delete). Keep the existing visibility gating, archive/unarchive flow, delete confirmation, and requirement alignment with FR-124.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 In the event page Edit event form, the Danger zone starts collapsed and expands through its heading toggle.
- [x] #2 When expanded, Archive/Unarchive event and Delete event render as full-width red-outlined buttons on separate rows inside one red rounded-corner frame without dividers between the rows.
- [x] #3 The Danger zone heading and actions appear only under the existing edit-mode and Event Owner management gate.
- [x] #4 Archive, unarchive, and delete-confirmation behavior is unchanged.
- [x] #5 FR-124 states the collapsed-by-default Danger zone and expand-to-reveal interaction.
- [x] #6 NewEvent unit tests cover the collapsed default, the expanded frame and rows, and the leading icons while retaining archive, unarchive, and delete-flow coverage.
- [x] #7 Firefox owner-authority and access-transfer e2e specs expand the Danger zone before using its actions and pass.
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
1. `frontend/src/components/NewEvent.vue`: add a collapsed-by-default `showDangerZoneActions` ref; keep the `.danger-zone` gate; wrap the actions in `ExpandableSection` (label `Danger zone`, same scroll behavior as Email reminders); add a `danger-zone-frame` with rounded red border; render both actions as full-width red-outlined buttons with leading archive/trash icons.
2. `docs/requirements/functional/fr/FR-124.md`: state the collapsed-by-default heading and expand-to-reveal controls.
3. `frontend/src/components/NewEvent.test.ts`: local model-aware ExpandableSection stub plus a rendering icon stub for the Danger-zone tests; cover collapsed default, expanded frame/buttons/icons, and expand before the existing archive, unarchive, and delete flows.
4. `e2e/specs/timed-event-owner-authority-firefox.spec.ts` and `e2e/specs/timed-event-access-transfer-firefox.spec.ts`: assert the collapsed state and expand the Danger zone before asserting or using its actions.
5. Checks: frontend lint, fmt:check, typecheck, build, test:unit; `npm run format:markdown` for FR-124; Firefox e2e for the two touched specs; `graphify update .`; finalize the task with evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-19 implementation and verification: `NewEvent.vue` renders the Danger zone in `ExpandableSection` (collapsed by default through `showDangerZoneActions`) with a `danger-zone-frame` (rounded red border) holding red-outlined, icon-led Archive/Unarchive and Delete rows. `NewEvent.test.ts` adds a model-aware ExpandableSection stub plus a color-capturing button stub, asserts the collapsed default, frame classes, red color, and per-button icons, and expands before the archive/unarchive/delete flows. E2E: owner-authority asserts the actions are hidden while collapsed and expands before archive/delete in both tests; access-transfer expands before asserting `Archive event`. Checks: frontend lint/fmt:check/typecheck/build/test:unit pass; root `format:markdown` and `format:markdown:check` pass; `graphify update .` run; e2e owner-authority 2/2 and access-transfer 8/8 pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Made the Edit event form Danger zone collapsible and reframed. It starts collapsed and expands through its `Danger zone` heading toggle (the existing ExpandableSection pattern used by Email reminders). When expanded, Archive/Unarchive event and Delete event render as full-width red-outlined buttons with leading archive/trash-can icons, one per row, inside a single red-bordered rounded frame with no divider lines. Existing visibility gating, archive/unarchive, and delete-confirmation behavior are unchanged. FR-124 now states the collapsed-by-default heading and expand-to-reveal controls; the frame and icons stay out of the requirement as incidental visual structure. Verification: frontend lint (0 errors; 2 pre-existing NewSignUp.test.ts warnings), fmt:check, typecheck, build, and test:unit (149 files / 1116 tests) pass; root format:markdown check and fmt:check pass; `graphify update .` run. E2E firefox-desktop: owner-authority 2/2 at default workers and access-transfer 8/8 with E2E_FRONTEND=bundled; both include the new expand step, and owner-authority asserts the actions stay hidden while collapsed. No swagger, contract-document, or `scripts/`/`prettier/` changes.
<!-- SECTION:FINAL_SUMMARY:END -->
