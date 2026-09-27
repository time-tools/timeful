---
id: TASK-0292
title: >-
  Fix the no-response desktop "Collapse disabled times" overflow in the event
  header
status: Done
assignee:
  - opencode
created_date: '2026-09-20 21:37'
updated_date: '2026-09-20 21:47'
labels: []
dependencies: []
references:
  - e2e/specs/event-page-no-responses-layout.spec.ts
  - frontend/src/views/Event.vue
modified_files:
  - e2e/specs/event-page-no-responses-layout.spec.ts
  - frontend/src/views/Event.vue
priority: medium
type: bug
ordinal: 292000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On a timed event page with no responses on desktop, the header shows Add availability, the inline "Collapse disabled times" switch, and Schedule event as a right-aligned stack that shares the narrow half of the 22rem header action column (about 10.75rem / 172px). The compact switch plus its label need roughly 14rem on one line, and the switch's dedicated CSS forces `inline-size: fit-content` on the selection control and `flex: 0 0 auto` on the label, so the content overflows the column instead of wrapping. The visual result is a "Collapse disabled times" row wider than the buttons above and below it, extending past the grid right edge.

The existing e2e geometry check in `e2e/specs/event-page-no-responses-layout.spec.ts` compares only the switch root box with the button box, so it does not catch the content overflow.

Approved direction: keep the narrow shared column and let the label wrap to two lines, with the break controlled before "disabled times" via an owned nowrap span so the noun phrase stays intact. Do not shorten the label, do not widen the shared column, and do not change the mobile or More options switch behaviour.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 In the desktop no-response state of a timed event page, the rendered content of the inline "Collapse disabled times" control (toggle and label) stays inside the horizontal bounds of the Add availability and Schedule event controls, with no overflow past the right edge of the header action column
- [x] #2 The "Collapse disabled times" label wraps to two lines inside the shared narrow action column instead of clipping or overflowing, with the break before "disabled times" so the noun phrase stays intact
- [x] #3 The toggle and wrapped label form a left-aligned text cluster centered in the same column as the Add availability and Schedule event controls, while the three controls keep their equal width and shared right alignment in the no-response desktop header
- [x] #4 The e2e no-response layout spec asserts the switch content containment so the overflow cannot regress
- [x] #5 Frontend lint, fmt:check, typecheck, build, and test:unit pass, and the affected e2e specs pass on chromium-desktop and chromium-mobile/firefox-desktop where their projects run them
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
1. Add a regression assertion to the chromium-desktop geometry block of `e2e/specs/event-page-no-responses-layout.spec.ts`: extend the existing `#collapse-disabled-times-toggle` page.evaluate to return the combined left/right bounds of the toggle wrapper and label, then assert those bounds stay within the Add availability button box (1px tolerance). Confirm it fails on the current code.

2. In `frontend/src/views/Event.vue`, mark up the desktop inline switch label (around line 580) as `Collapse` followed by a nowrap span around `disabled times`, so the label can wrap only before the noun phrase.

3. In the non-scoped style block, adjust `.desktop-event-header-options__collapse-disabled-times-switch` (around lines 2763-2780): let the selection control shrink to the column (`inline-size: 100%; min-inline-size: 0`) and let the label wrap (`flex: 1 1 auto; min-inline-size: 0; line-height: 1.25; overflow-wrap: break-word; white-space: normal`). Keep the 2.5rem minimum height and vertical centering; do not touch `desktop-event-header-single-column` or the shared compact switch CSS.

4. Verify: focused e2e on chromium-desktop (geometry + containment) and firefox-desktop, frontend lint/fmt:check/typecheck/build/test:unit, visual check of the two-line wrap, then `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation deviation from the draft plan: the label uses `flex: 0 0 auto` with `inline-size: min-content` instead of `flex: 1 1 auto`, so the toggle + label cluster stays centered in the column rather than filling it; the wrapped text is left-aligned (compact switch default) per review.

The firefox-desktop project only matches `timed-event-*firefox.spec.ts`, so the no-responses layout spec does not run there; Firefox coverage came from the add-availability hint spec and mobile coverage from the mobile editing options spec.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The desktop no-response event header no longer lets the inline `Collapse disabled times` switch overflow its narrow 172px action column.

- `frontend/src/views/Event.vue`: the desktop inline switch label now reads `Collapse` followed by a `tw:whitespace-nowrap` span around `disabled times`, so wrapping breaks before the noun phrase instead of splitting it. The dedicated `.desktop-event-header-options__collapse-disabled-times-switch` rules now let the label shrink and wrap (`inline-size: min-content`, `max-inline-size: 100%`, `line-height: 1.25`, `overflow-wrap: break-word`, `white-space: normal`) and let the selection control fill the column (`inline-size: 100%; min-inline-size: 0`), so the toggle and wrapped label form a left-aligned cluster centered by the existing `justify-content: center` on `.v-selection-control`.
- `e2e/specs/event-page-no-responses-layout.spec.ts`: the chromium-desktop geometry check now measures the combined toggle-wrapper and label bounds and asserts they stay inside the Add availability button box, catching content overflow that the root-box-only assertions missed.

## Scope notes

- The shared `desktop-event-header-single-column` half-column width is unchanged, so Add availability, the collapse switch, and Schedule event keep their equal width and shared right alignment.
- Mobile row layouts and the More options menu switch are untouched; their e2e specs still pass.

## Verification

- Regression-first: the new containment assertion failed before the fix (`collapseContentBounds.left` 1030.5 vs button left 1043) and passes after.
- E2E: `event-page-no-responses-layout.spec.ts` 3 passed on chromium-desktop (with the new containment assertion plus the existing containment/centering/width assertions); `event-mobile-editing-options.spec.ts` 4 passed on chromium-mobile and `timed-event-add-availability-hint-firefox.spec.ts` 3 passed on firefox-desktop.
- Frontend: lint, fmt:check, typecheck, build, and test:unit pass (154 files / 1232 tests).
- E2E package: lint and typecheck pass; `fmt:check` flags only the pre-existing, untouched `specs/timed-event-access-transfer-firefox.spec.ts`.
- `graphify update .` completed (6165 nodes, 11614 edges).
- No Swagger annotations, scripts, Markdown, or contract documents changed.
<!-- SECTION:FINAL_SUMMARY:END -->
