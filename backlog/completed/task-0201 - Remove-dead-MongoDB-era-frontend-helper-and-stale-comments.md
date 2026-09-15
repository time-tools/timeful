---
id: TASK-0201
title: Remove dead MongoDB-era frontend helper and stale comments
status: Done
assignee:
  - opencode
created_date: '2026-09-11 15:21'
updated_date: '2026-09-11 15:42'
labels: []
dependencies: []
priority: low
type: chore
ordinal: 239000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After the TASK-0199 MongoDB retirement, the frontend still exported a dead ObjectID date helper and kept comments and test names that described MongoDB-served events.

Outcome: the dead ObjectID date helper is removed, and comments and test names describe current PostgreSQL / Event Visitor Identity behavior with no migration-era references, without changing any compatibility branch behavior.

Constraints:
- Keep the existing compatibility branches gated on a missing `eventVisitorId`; only wording changes there.
- Do not change transport decoding, boundary types, or payload shapes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 dateFromObjectId, its utils re-export, and its unit test are removed, and no frontend reference remains.
- [x] #2 Comments and test names in frontend/src/views/Event.vue, frontend/src/components/groups/InvitationDialog.vue, and frontend/src/components/schedule_overlap/RespondentsList.test.ts no longer describe MongoDB-served events or migration-era behavior.
- [x] #3 npm run test:unit, npm run lint, npm run fmt:check, npm run typecheck, and npm run build pass.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Plan (researched 2026-09-11)

1. Remove the dead ObjectID helper:
   - Delete `dateFromObjectId` from `frontend/src/utils/dateBoundaryAdapters.ts`.
   - Remove its re-export from the `./dateBoundaryAdapters` block in `frontend/src/utils/index.ts`.
   - Remove its import and `it("converts MongoDB object ids ...")` case from `frontend/src/utils/dateBoundaryAdapters.test.ts`; keep the `convertToUTC` and `convertUTCSlotsToLocalISO` cases.

2. Reword comments and test names that describe MongoDB-served events (behavior unchanged):
   - `frontend/src/views/Event.vue` around line 1287 (`userHasResponded` fallback) and around line 1920 (plugin guest path).
   - `frontend/src/components/groups/InvitationDialog.vue` around line 193 (fallback response path).
   - `frontend/src/components/schedule_overlap/RespondentsList.test.ts` (owner delete test name).

3. Verify no remaining MongoDB wording in `frontend/src`, then run `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` from `frontend/`.

No boundary types, decoding, or payload shapes change; the compatibility branches themselves stay intact per the task constraints.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Removed the dead dateFromObjectId helper, its utils re-export, and its test; reworded the three MongoDB comments; renamed the stale RespondentsList legacy-payload test.

Checks: lint, fmt:check, typecheck, build pass; unit suite 146 files / 1087 tests pass (down one from the removed ObjectID test).

Discovered but not changed (outside confirmed scope): bson-objectid remains imported by Dashboard.vue for ObjectID-timestamp event sorting and by useSignUpForm.ts for client-generated signup block _id values. Recommend a follow-up decision on whether those paths are dead.

No e2e was required: no client-visible behavior changed.

Follow-up polish: renamed the owner-delete test to the response-map wording and confirmed the frontend checks after the rename: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` (146 files, 1087 tests) all pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Removed the dead MongoDB-era ObjectID helper and the stale MongoDB and migration-era wording from the frontend.

### Changes
- Deleted `dateFromObjectId` from `frontend/src/utils/dateBoundaryAdapters.ts` and its re-export from `frontend/src/utils/index.ts`; removed its unit test case. The file now only exports `convertToUTC` and `convertUTCSlotsToLocalISO`.
- Reworded comments in `frontend/src/views/Event.vue` (`userHasResponded` fallback and the plugin guest path) and `frontend/src/components/groups/InvitationDialog.vue` (fallback response path) to describe PostgreSQL / Event Visitor Identity behavior.
- Renamed the stale test `keeps the legacy delete payload for MongoDB respondents` to `keeps the response-map delete payload for events without an Event Visitor Identity`. The compatibility branch itself is unchanged.

### Verification
- `rg -i mongo|dateFromObjectId frontend/src`: no hits.
- `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`: all pass (146 files, 1087 tests; one removed test accounts for the previous 1088).
- No transport, boundary, or payload-shape change; the `!eventVisitorId` compatibility branches were left intact.

### Follow-up (not in scope)
- `bson-objectid` is still imported by `frontend/src/components/home/Dashboard.vue` (ObjectID-timestamp event sort) and `frontend/src/composables/schedule_overlap/useSignUpForm.ts` (client-generated `_id` for new signup blocks). These are MongoDB-era dependencies, but determining whether the code paths are dead requires analysis and was outside this task's confirmed scope.
- The `objectIdLike` guest-name guard in `frontend/src/utils/guestName.ts` is intentionally retained as a name/ID-confusion guard.
<!-- SECTION:FINAL_SUMMARY:END -->
