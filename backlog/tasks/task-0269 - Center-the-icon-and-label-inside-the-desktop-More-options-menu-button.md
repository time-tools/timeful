---
id: TASK-0269
title: Center the icon and label inside the desktop More options menu button
status: Done
assignee: []
created_date: '2026-09-19 15:40'
updated_date: '2026-09-19 15:49'
labels:
  - frontend
  - styling
dependencies: []
priority: medium
type: bug
ordinal: 269000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the desktop event page, the More options menu activator is a full-width outlined button, but its menu activator class includes `tw:justify-between`, which pushes the tune icon to the left edge and the "More options" label to the right edge. Expected: the icon and label render as one centered group inside the button.

Affected activator bindings in `frontend/src/views/Event.vue`: the non-editing desktop header menu (`menu-activator-class="desktop-event-header-control desktop-event-header-options__menu-button tw:justify-between tw:w-full"`, around line 527) and the desktop editing header menu (same class string, around line 643). The shared activator button lives in `frontend/src/components/schedule_overlap/EventOptions.vue` and already has a centered default content layout; the mobile `ToolRow.vue` activator uses `tw:w-fit` and must stay a compact, left-aligned fit-content button.

Acceptance Criteria:
--------------------------------------------------
- [ ] #1 Both desktop More options activators center the tune icon and label as a group inside the full-width button instead of spreading them to the edges.
- [ ] #2 Regression coverage in `frontend/src/views/Event.test.ts` fails on the old `tw:justify-between` activator class and passes with the centered class.
- [ ] #3 The mobile timed-event toolbar More options button (`ToolRow.vue`) keeps its compact fit-content layout and existing behavior unchanged.
- [ ] #4 Required frontend checks pass from `frontend/`: lint, fmt:check, typecheck, build, test:unit.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Both desktop More options activators center the tune icon and label as a group inside the full-width button instead of spreading them to the edges.
- [x] #2 Regression coverage in frontend/src/views/Event.test.ts fails on the old tw:justify-between activator class and passes with the centered class.
- [x] #3 The mobile timed-event toolbar More options button (ToolRow.vue) keeps its compact fit-content layout and existing behavior unchanged.
- [x] #4 Required frontend checks pass from frontend/: lint, fmt:check, typecheck, build, test:unit.
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
1. In `frontend/src/views/Event.test.ts`, add regression assertions in the desktop display-options test and the desktop editing-options test that the More options `menuactivatorclass` contains `tw:justify-center` and does not contain `tw:justify-between` (these fail before the fix).
2. In `frontend/src/views/Event.vue`, replace `tw:justify-between` with `tw:justify-center` on both desktop More options activator bindings (non-editing and editing headers), leaving `tw:w-full` so the button stays full-width with its content centered.
3. Leave `frontend/src/components/schedule_overlap/EventOptions.vue` and `ToolRow.vue` untouched: the activator button already centers its content by default, and the mobile button intentionally stays `tw:w-fit`.
4. Run the targeted Event unit test, then the required frontend checks (lint, fmt:check, typecheck, build, test:unit) and `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Root cause: the desktop More options activator is a full-width v-btn whose activator class forced `tw:justify-between`, so content aligned to the two edges. Vuetify's `.v-btn` base rule already sets `justify-content: center`; switching the activator class to `tw:justify-center` makes the centering explicit and testable while `tw:w-full` keeps the button full width.

E2E flake observed: chromium-desktop and chromium-mobile focused runs initially failed with blank event pages because dev-server module requests got `net::ERR_NETWORK_CHANGED` (confirmed in trace.zip console logs). Retries of the same selections passed (desktop 12/12, mobile 4/4). This is environmental, not caused by the activator class change.

The mobile ToolRow activator (`tw:w-fit`) and EventOptions.vue were intentionally not modified; existing ToolRow unit assertions for `menu-activator-class="tw:w-fit"` and the mobile e2e layout spec remain green.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Problem

On the desktop event page the More options menu activator is a full-width outlined button, but its activator class included `tw:justify-between`, pushing the tune icon to the left edge and the "More options" label to the right edge instead of rendering them as one centered group.

## Fix

`frontend/src/views/Event.vue`: both desktop `menu-activator-class` bindings (non-editing header at line 527 and desktop editing header at line 643) now use `tw:justify-center` instead of `tw:justify-between`, keeping `tw:w-full`. The shared activator in `frontend/src/components/schedule_overlap/EventOptions.vue` was left untouched: it already renders its content centered, and the mobile `ToolRow.vue` activator keeps `tw:w-fit` compact left alignment.

## Tests

- Regression coverage added in `frontend/src/views/Event.test.ts` for both desktop bindings: activator class must contain `tw:justify-center` and must not contain `tw:justify-between`. Both assertions fail on the old class and pass after the fix.
- Required frontend checks from `frontend/`: lint (0 errors; 2 pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, build, test:unit (149 files, 1126 tests) all pass.
- Focused e2e: `chromium-desktop specs/event-page-days-only-layout.spec.ts specs/event-page-no-responses-layout.spec.ts` 12 passed; `chromium-mobile specs/event-toolbar-mobile-layout.spec.ts` 4 passed.
- `graphify update .` run; no code-graph topology changes detected.

## Risks / notes

- The first e2e attempts failed with blank pages because the dev server modules failed to load with `net::ERR_NETWORK_CHANGED` (environment flake, confirmed in traces); the same selections passed on retry, and the failures were unrelated to the one-class change.
<!-- SECTION:FINAL_SUMMARY:END -->
