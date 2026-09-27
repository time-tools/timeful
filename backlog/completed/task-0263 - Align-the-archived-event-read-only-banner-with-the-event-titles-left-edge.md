---
id: TASK-0263
title: Align the archived-event read-only banner with the event title's left edge
status: Done
assignee: []
created_date: '2026-09-19 13:01'
updated_date: '2026-09-19 13:11'
labels: []
dependencies: []
references:
  - frontend/src/views/Event.vue
  - e2e/specs/timed-event-owner-authority-firefox.spec.ts
  - e2e/specs/event-page-no-responses-layout.spec.ts
  - e2e/helpers/visual-gap-helpers.ts
priority: medium
type: bug
ordinal: 263000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The event page renders the "This event is archived and read-only." banner as a v-alert directly under the page wrapper in frontend/src/views/Event.vue (currently around lines 4-11), outside the centered content column that holds the event header. The event title renders inside that column, under the tw:mx-auto tw:max-w-5xl and tw:mx-4 wrappers in #event-header. As a result the banner's left edge does not line up with the event title's left edge, and the mismatch is visible on both desktop and phone widths. Align the archived read-only banner's left edge with the event title's left edge at all supported breakpoints using a layout-based change, without absolute positioning, negative margins, or one-off overrides. Keep the banner's existing copy, visibility condition (event.eventVisitorId && event.isArchived), and tonal info styling unchanged, and do not alter read-only enforcement. Add regression coverage for the alignment, or record a concrete reason it cannot be covered at the unit or e2e layer.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a wide desktop viewport, the archived read-only banner's left edge aligns with the event title's left edge (within 1px).
- [x] #2 On a phone-width viewport, the archived read-only banner's left edge aligns with the event title's left edge (within 1px).
- [x] #3 The banner keeps its existing text, visibility condition, and tonal info styling, and read-only behavior is unchanged.
- [x] #4 Regression coverage verifies the left-edge alignment, or the task records a concrete reason layout cannot be covered at the unit or e2e layer.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. frontend/src/views/Event.vue: move the archived read-only v-alert from the top-level page wrapper into the existing `tw:mx-auto tw:max-w-5xl tw:flex-1` content column, directly before the `v-if="!isSettingSpecificTimes"` header wrapper, and add `tw:mx-4` so the alert box shares the title's horizontal inset. Keep the text, `event.eventVisitorId && event.isArchived` condition, `type="info"`, `variant="tonal"`, and `tw:mb-4` unchanged. The column's existing `tw:mt-4` means the alert's top gap grows by 16px while the title's position is unchanged; the title/title-adjacent layout tests only cover non-archived events.
2. frontend/src/views/Event.test.ts: add a regression test that mounts an archived event state and asserts the banner renders, carries `tw:mx-4`, and lives in the same `.tw\:max-w-5xl` column ancestor as `#event-header`.
3. e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts (new): seed a canonical timed event via `seedCanonicalTimedEvent(page.request, ...)`, archive it with `POST /api/events/:eventId/archive`, open it, and assert the `.v-alert` left edge equals the `#event-header` title left edge within 1px at the desktop viewport, then repeat after `page.setViewportSize({ width: 375, height: 900 })` using `expect.poll` so responsive re-render is awaited.
4. Run frontend lint, fmt:check, typecheck, build, test:unit; run the new spec under firefox-desktop; run `graphify update .`.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Moved the archived read-only banner into the event page's centered content column so its left edge matches the event title's left edge. frontend/src/views/Event.vue now renders the v-alert inside the `tw:mx-auto tw:max-w-5xl tw:flex-1` column, directly before the `v-if="!isSettingSpecificTimes"` header wrapper, with the added `tw:mx-4` inset; text, visibility condition, type/variant, and `tw:mb-4` are unchanged, and the title's position is unaffected. The banner's top gap grows by the column's existing 16px `tw:mt-4` while the title position stays the same.

Regression coverage:
- Unit: frontend/src/views/Event.test.ts "renders the archived read-only banner inside the event content column" asserts the banner text, the `tw:mx-4` inset, and that the banner and #event-header share the same `.tw\:max-w-5xl` column ancestor. Verified it fails against the previous layout (`expected [ 'tw:mb-4' ] to include 'tw:mx-4'`).
- E2E: new e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts seeds and archives a canonical timed event, then asserts the `.v-alert` and title left edges match within 1px at the desktop viewport and after resizing to 375x900. Verified it fails against the previous layout with a 224px desktop mismatch and passes with the fix.

Checks run:
- frontend: lint (0 errors; 2 pre-existing vue/one-component-per-file warnings in NewSignUp.test.ts), fmt:check, typecheck, build, test:unit (149 files, 1118 tests) all pass.
- e2e: new alignment spec plus timed-event-owner-authority-firefox.spec.ts and event-page-no-responses-layout.spec.ts (chromium-desktop) pass (5 passed).
- root `npm run format:markdown:check` passes; `graphify update .` run.
- No swagger, scripts/, prettier/, or contract documents changed.
<!-- SECTION:FINAL_SUMMARY:END -->
