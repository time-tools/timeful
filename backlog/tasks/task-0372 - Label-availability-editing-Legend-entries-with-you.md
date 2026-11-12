---
id: TASK-0372
title: Label availability editing Legend entries with (you)
status: Done
assignee: []
created_date: '2026-11-12 12:10'
updated_date: '2026-11-12 12:10'
labels:
  - frontend
  - legend
  - copy
dependencies: []
modified_files:
  - docs/requirements/functional/fr/FR-009.md
  - frontend/src/components/schedule_overlap/ColorLegend.vue
  - frontend/src/components/schedule_overlap/ColorLegend.test.ts
priority: low
type: enhancement
ordinal: 311000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the event availability-editing page, the Legend Section's availability entries used bare status labels, so they did not clarify that the colors reflect the current respondent's own availability. While adding or editing availability, the Legend Section items should read `Available (you)`, `If needed (you)`, and `Unavailable (you)`. The pale-red editing entry previously read `Unavailable, change in Add/Edit availability`; this task replaces that guidance text with `Unavailable (you)`.

The read-mode event-page Legend Section labels remain `Available (at least one respondent)`, `If needed (at least one respondent)`, and `Unavailable (all respondents)`.

This follows TASK-0371, which deliberately left the availability-editing legend unchanged. The legend is shared with the Event Owner's edit-owned-guest flow, so the static `(you)` labels also appear while editing another respondent's availability.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 In availability editing, the Legend Section items read exactly `Available (you)`, `If needed (you)`, and `Unavailable (you)`.
- [x] #2 The read-mode Legend Section labels remain `Available (at least one respondent)`, `If needed (at least one respondent)`, and `Unavailable (all respondents)`.
- [x] #3 Unit coverage asserts the availability-editing label set and that read-mode availability labels do not appear while editing.
- [x] #4 FR-009 records the editing-mode labels, and the required frontend checks plus the code knowledge graph refresh pass.
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Update the `isAddingAvailability` branches in `ColorLegend.vue` to `Available (you)`, `If needed (you)`, and `Unavailable (you)`.
2. Update `ColorLegend.test.ts` to assert the editing label set and keep the read-mode label assertions.
3. Record the editing-mode labels in FR-009.
4. Run the focused and full frontend checks, Markdown checks, and the code knowledge graph refresh.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation is complete. The editing-mode `Available`/`If needed`/`Unavailable, change in Add/Edit availability` items now read `Available (you)`, `If needed (you)`, and `Unavailable (you)`; the read-mode event-page labels are unchanged.

Focused legend/sidebar tests passed (31 tests), the full unit suite passed (159 files / 1298 tests), and frontend `lint`, `fmt:check`, `typecheck`, and `build` passed. Root Markdown format and lint checks passed, and `npm run format:markdown` produced no changes. `codebase-memory-mcp cli index_repository --repo-path .` refreshed the graph.

E2e coverage: no e2e spec asserts these legend labels, so no e2e run was required.

The shared Legend also reads `(you)` during the Event Owner's edit-owned-guest flow; this is recorded in the task description as an explicit scope note.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The availability-editing Legend Section now labels the current respondent's states `Available (you)`, `If needed (you)`, and `Unavailable (you)`, replacing the bare labels and the `Unavailable, change in Add/Edit availability` guidance. The read-mode event-page labels are unchanged.

**Changes**
- `frontend/src/components/schedule_overlap/ColorLegend.vue`: editing-mode labels use `(you)`; read-mode labels remain `Available (at least one respondent)`, `If needed (at least one respondent)`, and `Unavailable (all respondents)`.
- `frontend/src/components/schedule_overlap/ColorLegend.test.ts`: asserts the editing label set and that the read-mode labels do not appear while editing.
- `docs/requirements/functional/fr/FR-009.md`: records the editing-mode labels.

**Verification**
- Focused legend/sidebar tests: 31 passed.
- Full unit suite: 159 files / 1298 tests passed.
- Frontend lint (0 errors; two pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, and build passed.
- Root Markdown format and lint checks passed; `npm run format:markdown` produced no changes.
- `codebase-memory-mcp cli index_repository --repo-path .` refreshed the graph.
- No e2e spec asserts these labels, so no e2e run was required.

**Note**
- The shared Legend also reads `(you)` in the Event Owner's edit-owned-guest flow, as recorded in the task description.
<!-- SECTION:FINAL_SUMMARY:END -->
