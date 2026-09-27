---
id: TASK-0268
title: Outline all Legend Section square indicators in shared grey
status: Done
assignee: []
created_date: '2026-09-19 15:31'
updated_date: '2026-09-19 15:35'
labels: []
dependencies: []
documentation:
  - docs/requirements/functional/fr/FR-125.md
priority: medium
type: enhancement
ordinal: 268000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Legend Section's square indicators currently use mixed outline treatments: most use a light grey `#bdbdbd` outline, the Scheduled event indicator's outline matches its blue fill, and the Disabled, outside indicator's outline matches its dark-grey fill, so those two squares show no visible border. Owner confirmed every square indicator should share one clearly visible grey outline, set to `#999999` (the existing grid-line grey), so the legend reads uniformly and no indicator blends into its background. The durable behavior is recorded as FR-125.

Owner decisions:
- All Legend Section square outlines use `#999999`, including the Scheduled event indicator.
- The Disabled, outside the event dates indicator keeps its dark-grey fill; only its outline becomes `#999999`.
- The Disabled, collapsed indicator keeps its dashed outline (same `#999999` grey).
- Indicator fills and labels are unchanged, and this is legend-only: grid cell and scheduled-event block treatments and the respondents-list status squares stay as they are.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every square indicator in the Legend Section renders an outline in the shared #999999 grey, including the Scheduled event indicator, and the Disabled, collapsed indicator keeps its dashed outline in the same grey.
- [x] #2 No Legend Section square indicator's outline color matches its fill, and no indicator retains the previous border-outline-neutral or border-scheduled-event outline treatment.
- [x] #3 A regression test in frontend/src/components/schedule_overlap/ColorLegend.test.ts asserts the shared outline on every indicator and the absence of the previous outline colors.
- [x] #4 FR-125 exists at docs/requirements/functional/fr/FR-125.md with the standard front matter and is listed in docs/requirements/README.md.
- [x] #5 Indicator fill colors and labels are unchanged; grid cell and scheduled-event block treatments and respondents-list status squares are unchanged.
- [x] #6 Frontend checks npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit pass.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Implementation plan (researched against the current system):

1. Author `docs/requirements/functional/fr/FR-125.md` (`status: proposed`, component `frontend`) requiring every Legend Section square indicator, including the Scheduled event indicator, to render an outline in the same grey that is discernible against its fill; add the index row to `docs/requirements/README.md`.
2. Update `frontend/src/components/schedule_overlap/ColorLegend.vue`: replace the five `tw:border-outline-neutral` indicator classes and the Scheduled event indicator's `tw:border-scheduled-event` with `tw:border-(--timeful-grid-line-color)` (`#999999`). Leave the collapsed indicator's dashed `var(--timeful-grid-line-color)` CSS and every fill class unchanged.
3. Update `frontend/src/components/schedule_overlap/ColorLegend.test.ts`: refresh the Scheduled event class expectation and add a regression test that every indicator renders the shared grey outline (collapsed via `.color-legend-indicator--collapsed`) and that no indicator retains `tw:border-outline-neutral` or `tw:border-scheduled-event`.
4. Run `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` in `frontend/`; format changed Markdown with root `npm run format:markdown`; run `graphify update .`.
5. Finalize TASK-0268 with per-criterion evidence and a final summary.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: ColorLegend.vue now applies `tw:border-(--timeful-grid-line-color)` to the six solid square indicators, including the Scheduled event indicator; the Disabled, collapsed indicator keeps its dashed `var(--timeful-grid-line-color)` outline. Fills and labels unchanged; git diff confirms no other component files were touched.

Verification evidence: ColorLegend.test.ts 9/9, including the new regression test that all seven indicators share the grid-line outline and that `tw:border-outline-neutral`/`tw:border-scheduled-event` no longer appear. Full frontend suite 149 files / 1126 tests pass; lint 0 errors (2 pre-existing warnings in untouched NewSignUp.test.ts); fmt:check, typecheck, and build pass. The built CSS contains `.tw\:border-\(--timeful-grid-line-color\){border-color:var(--timeful-grid-line-color)}`. Root format:markdown and lint:markdown pass; `graphify update .` run.

DoD #3 (e2e) left unchecked: no e2e spec asserts legend indicator styling, and the two legend-touching layout specs only measure the `Legend` heading bounding box, which a border-color change cannot affect; frontend/AGENTS.md required checks do not include e2e. DoD #5, #7, and #8 are not applicable: no Swagger annotations, scripts/ or prettier/ files, or contract documents changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Applied a single shared grey outline to every square indicator in the schedule-overlap Legend Section and recorded the durable behavior as FR-125.

Changes:
- frontend/src/components/schedule_overlap/ColorLegend.vue: the six solid indicators (Available, If needed, Unavailable, Disabled inside, Disabled outside, Scheduled event) now use `tw:border-(--timeful-grid-line-color)` (#999999); the Scheduled event indicator no longer uses `border-scheduled-event`, and the Disabled, collapsed indicator keeps its dashed `var(--timeful-grid-line-color)` outline. Fills and labels are unchanged; grid cells, scheduled-event blocks, and respondents-list status squares are untouched.
- frontend/src/components/schedule_overlap/ColorLegend.test.ts: refreshed the Scheduled event expectation and added a regression test asserting every indicator shares the grid-line grey outline and that the previous outline classes are gone (9/9 pass).
- docs/requirements/functional/fr/FR-125.md (new, status proposed) and docs/requirements/README.md: requirement that every Legend Section square indicator, including the Scheduled event indicator, renders an outline in the same grey that is discernible against its fill.

Verification: frontend lint (0 errors; 2 pre-existing warnings in untouched NewSignUp.test.ts), fmt:check, typecheck, build, and test:unit (149 files / 1126 tests) pass; the built CSS resolves the new utility to `border-color: var(--timeful-grid-line-color)`; root format:markdown and lint:markdown pass; `graphify update .` run. E2E not run: no spec asserts legend indicator styling, the legend-touching layout specs only measure the `Legend` heading box that a border-color change cannot affect, and the required frontend check set does not include e2e.
<!-- SECTION:FINAL_SUMMARY:END -->
