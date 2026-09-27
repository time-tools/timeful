---
id: TASK-0306
title: >-
  Match the mobile scheduling Cancel and Clear buttons to the desktop red
  destructive treatment
status: Done
assignee:
  - opencode
created_date: '2026-09-21 16:41'
updated_date: '2026-09-21 16:51'
labels: []
dependencies: []
references:
  - frontend/src/views/Event.vue
documentation:
  - frontend/AGENTS.md
priority: medium
type: bug
ordinal: 306000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On mobile, the scheduling-mode Cancel and Clear buttons in the event action bar render with blue text and blue borders, while the equivalent desktop header buttons render with the destructive red treatment. The same destructive actions therefore look non-destructive on phones and the two viewports disagree. The desktop scheduling Cancel and Clear buttons are the reference appearance; the mobile editing Cancel button already follows the red destructive treatment.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a phone viewport in scheduling mode, the Cancel and Clear buttons in the mobile event action bar render with the same destructive red treatment as the desktop header scheduling Cancel and Clear buttons, with no remaining blue text or blue border on either control.
- [x] #2 The mobile scheduling Cancel and Clear buttons keep their current labels, visibility conditions (Clear only when a saved Timeful schedule exists), actions, and disabled behavior.
- [x] #3 Regression coverage asserts the mobile scheduling Cancel and Clear controls use the red destructive treatment and not the blue treatment, at the layer that can observe the rendered styling.
- [x] #4 npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit pass in frontend/.
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
1. Add the regression test first in `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` (firefox-touch project, 375x900, hasTouch). Seed a canonical timed event with `seedCanonicalTimedEvent(page.request, ...)`, then save a Timeful schedule with `page.request.put('/api/events/<eventId>/schedule', { startDate, endDate })` so the owner creation cookies are shared with the browser. Open the event, enter scheduling mode with the mobile "Reschedule" button, and assert the mobile Cancel and Clear buttons compute red text and a red outlined border (and therefore not the current blue). Run it once to confirm it fails against current code.
2. Fix `frontend/src/views/Event.vue` in the mobile `isScheduling` branch: replace `tw:border-blue tw:text-blue` with `tw:text-red` on Cancel and Clear, matching the desktop reference. Vuetify's outlined variant uses `border: thin solid currentColor` (`VBtn.css`), so the desktop class set alone yields the red border.
3. Extend the mobile scheduling unit tests in `frontend/src/views/Event.test.ts` to assert Cancel and Clear carry `tw:text-red` and not `tw:border-blue`/`tw:text-blue`, so the fast unit layer pins the class contract.
4. Run the required frontend checks (`npm run lint`, `fmt:check`, `typecheck`, `build`, `test:unit`), the touch e2e spec with `--project=firefox-touch`, and `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Root cause: the mobile `isScheduling` branch hard-coded `tw:border-blue tw:text-blue` while the desktop reference used `tw:text-red`. Because Vuetify's outlined variant sets `border: thin solid currentColor`, mirroring the desktop text-red utility also colors the border red, so no extra border class is needed.

Test-first: the new firefox-touch e2e regression test was run before the fix and failed with computed color `rgb(0, 107, 232)` (blue); after the one-class fix it passes with `rgb(219, 22, 22)`.

The pre-fix e2e run also confirmed the API setup path works: seeding with `page.request` plus `PUT /api/events/<eventId>/schedule` gives the browser owner authority and makes Clear visible via `hasSavedTimefulSchedule`.

Verification: firefox-touch spec 11/11 passed; frontend test:unit 1254 passed; lint/fmt:check/typecheck/build passed; `graphify update .` run. E2E `npm run fmt:check` fails only on pre-existing committed formatting in `specs/timed-event-access-transfer-firefox.spec.ts`, unrelated to this change.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Problem

On a phone viewport, the scheduling-mode Cancel and Clear buttons in the mobile event action bar rendered blue (`tw:border-blue tw:text-blue`), while the equivalent desktop header buttons rendered the destructive red treatment (`tw:text-red`). The same destructive actions looked non-destructive on phones.

## Fix

`frontend/src/views/Event.vue` mobile `isScheduling` branch: Cancel and Clear now use `tw:text-red` (Clear keeps `tw:ml-2`). Vuetify's outlined variant draws its border with `border: thin solid currentColor` (`VBtn.css`), so the desktop class set alone yields both red text and a red border; the explicit blue border utility is gone.

## Tests

- Regression coverage added at the layer that observes rendered styling: a new `firefox-touch` e2e test in `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` seeds a canonical timed event with an owner API-saved Timeful schedule, enters scheduling mode via the mobile "Reschedule" button, and asserts computed `color` and `border-top-color` are `rgb(219, 22, 22)` on both Cancel and Clear. It failed before the fix with computed blue `rgb(0, 107, 232)` and passes after.
- `frontend/src/views/Event.test.ts` mobile scheduling tests now assert Cancel and Clear carry `tw:text-red` and not `tw:border-blue`/`tw:text-blue`.
- Full `firefox-touch` spec: 11/11 pass.
- Required frontend checks from `frontend/`: lint (0 errors; 2 pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, build, and test:unit (157 files, 1254 tests) pass.
- `graphify update .` run.

## Risks / notes

- E2E package `npm run fmt:check` reports a pre-existing oxfmt issue in `specs/timed-event-access-transfer-firefox.spec.ts` (committed, untouched by this change); the new test file is correctly formatted. E2E lint and typecheck pass.
- No Markdown, Swagger, contract, or scripts changes are part of this task.
<!-- SECTION:FINAL_SUMMARY:END -->
