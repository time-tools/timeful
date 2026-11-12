---
id: TASK-0371
title: Clarify event-page Legend availability labels
status: Done
assignee: []
created_date: '2026-11-12 11:55'
updated_date: '2026-11-12 11:55'
labels:
  - frontend
  - legend
  - copy
dependencies: []
references:
  - e2e/specs/timed-event-respondent-selection-firefox.spec.ts
modified_files:
  - docs/requirements/functional/fr/FR-009.md
  - frontend/src/components/schedule_overlap/ColorLegend.vue
  - frontend/src/components/schedule_overlap/ColorLegend.test.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlapSidebar.test.ts
priority: low
type: enhancement
ordinal: 310000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the event page, the Legend Section described availability colors with bare status words, so the palette did not explain whose availability each color reflects. The event-page legend should read `Available (you or at least one respondent)`, `If needed (you or at least one respondent)`, and `Unavailable (all respondents)`.

The availability-editing legend and the Available/If needed availability toggle keep their existing wording, so the availability editing page is unchanged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The event-page Legend Section items read exactly `Available (you or at least one respondent)`, `If needed (you or at least one respondent)`, and `Unavailable (all respondents)`.
- [x] #2 The availability-editing legend keeps `Available`, `If needed`, and `Unavailable, change in Add/Edit availability`, and the Available/If needed availability toggle stays unchanged.
- [x] #3 Unit coverage asserts both the event-page and availability-editing label sets, and the focused Firefox legend/respondent alignment e2e spec passes.
- [x] #4 FR-009 records the mode-specific legend wording; required frontend checks (lint, fmt:check, typecheck, build, test:unit) pass and changed Markdown is formatted.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The event-page Legend Section now explains whose availability each color reflects: `Available (you or at least one respondent)`, `If needed (you or at least one respondent)`, and `Unavailable (all respondents)`. The availability-editing legend and the Available/If needed toggle keep their existing wording.

**Changes**
- `frontend/src/components/schedule_overlap/ColorLegend.vue`: event-page labels use the respondent-aware wording; `isAddingAvailability` keeps the original editing labels.
- `frontend/src/components/schedule_overlap/ColorLegend.test.ts`: asserts both label sets, including that the aggregate labels do not leak into availability editing.
- `frontend/src/components/schedule_overlap/ScheduleOverlapSidebar.test.ts`: asserts the event-page unavailable label.
- `docs/requirements/functional/fr/FR-009.md`: records the editing-mode wording and the respondent-aware event-page legend labels.

**Verification**
- Focused legend tests: 31 passed.
- Full unit suite: 159 files / 1298 tests passed.
- Frontend lint (0 errors; two pre-existing NewSignUp.test.ts warnings), fmt:check, typecheck, and build passed.
- Root format:markdown:check and lint:markdown passed; `npm run format:markdown` produced no changes.
- Focused Firefox e2e legend/respondent alignment spec passed (1 test); the initial attempt was blocked by the sandbox read-only Docker buildx cache and passed after pointing DOCKER_CONFIG/BUILDX_CONFIG at /tmp.
- `codebase-memory-mcp cli index_repository --repo-path .` refreshed the graph.
<!-- SECTION:FINAL_SUMMARY:END -->
