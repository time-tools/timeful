---
id: TASK-0137
title: Validate event name length in the event create and update APIs (FR-119)
status: Done
assignee:
  - opencode
created_date: '2026-09-02 11:18'
updated_date: '2026-09-13 18:16'
labels:
  - backend
dependencies: []
references:
  - docs/requirements/functional/fr/FR-119.md
  - TASK-0136
  - server/routes/postgres_event_routes.go
priority: medium
type: task
ordinal: 150300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up to TASK-0136. FR-119 (docs/requirements/functional/fr/FR-119.md) requires that the event create and update APIs reject an event name longer than 100 characters, but the backend currently accepts names of any length.

Add server-side validation of the event name length on event creation and update so the backend enforces the FR-119 cap:

- Count the limit in Unicode code points (runes), matching the respondents.ValidateGuestName/MaxGuestNameLength convention (server/respondents/identity.go:12,63), so names of exactly 100 non-ASCII characters remain valid.
- Reject event names longer than 100 code points on POST /events and PUT /events/:id with a 400 response carrying the stable event-name-too-long error token, without persisting the event or the change.
- Enforce the cap in the PostgreSQL event handlers postgresCreateEvent and postgresEditEvent (server/routes/postgres_event_routes.go), keeping the handler/service-layer conventions used by the existing event routes; PostgreSQL access stays in server/postgres/ (server/db/ no longer exists after the MongoDB retirement).
- Add route tests for both boundaries plus a non-ASCII boundary case and a no-persistence assertion, following the Server Test Workflow: start postgres-test via compose.test.yaml and run --rm server-route-test against the timeful-test database.
- Add @Failure 400 for the new rejection to both handler annotations and regenerate swagger per Backend Conventions (swag init, then npm run gen:api).

Out of scope: the NewSignUp.vue sign-up event form still submits an uncapped name field; the client-side cap and rejection UX are tracked by TASK-0220.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The event create and update APIs reject an event name longer than 100 Unicode code points with a 400 response carrying the event-name-too-long token and do not persist the event or the change, per FR-119.
- [x] #2 An event name of exactly 100 characters is accepted, including a non-ASCII name of 100 code points.
- [x] #3 Route tests in the isolated Compose test stack cover the over-limit rejection, the 100-character acceptance boundary (including a non-ASCII case), and that a rejected change was not persisted.
- [x] #4 POST /events and PUT /events/{eventId} document @Failure 400 and the regenerated server/docs/swagger.* and frontend/src/types/api.ts stay consistent with the handler responses.
- [x] #5 The scoped server-route-test suite passes on the isolated Compose test stack.
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
1. Add a routes-package event-name length check (MaxEventNameLength = 100, utf8.RuneCountInString) returning the stable event-name-too-long token.
2. Enforce it in postgresCreateEvent and postgresEditEvent before any repository work so rejected requests cannot persist.
3. Add @Failure 400 to both handler annotations and regenerate swagger (swag init, npm run gen:api).
4. Add route tests for 101-rune rejection (ASCII and non-ASCII), 100-rune acceptance (ASCII and non-ASCII) on create and update, plus no-persistence assertions through pgstore.Pool and a follow-up GET.
5. Run the backend suite through the isolated compose.test.yaml server-route-test service, then run frontend checks if the regenerated api.ts changes.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation notes (2026-09-13):
- server/errs/errors.go: added the stable EventNameTooLong = "event-name-too-long" token.
- server/models/event.go: added MaxEventNameLength = 100.
- server/routes/postgres_event_routes.go: postgresCreateEvent and postgresEditEvent now reject utf8.RuneCountInString(name) > models.MaxEventNameLength with 400 responses.Error{Error: errs.EventNameTooLong} before any repository work, so rejected requests cannot persist; both routes gained @Failure 400 annotations.
- server/routes/postgres_event_name_length_test.go: four new route tests cover ASCII and non-ASCII 101-rune rejection with token and no-persistence (CREATE: count rows by name; PUT: re-read stored name) plus ASCII and non-ASCII 100-rune acceptance on create and update.
- Regenerated server/docs/docs.go, server/docs/swagger.json, server/docs/swagger.yaml, and frontend/src/types/api.ts; the only API diff is the new 400 responses.

Evidence:
- Focused: go test ./routes/ -run 'NamesLongerThan100|NamesAt100' -v -count=1 -> 4 tests / 8 subtests PASS.
- Full isolated stack: docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test -> every package ok (timeful/server/routes 3.259s).
- Frontend: npm run lint, fmt:check, typecheck, build, and test:unit (146 files / 1093 tests) all pass.
- No Markdown files changed, so format:markdown was not needed; no e2e was required for this backend-only change.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-13 15:06
---
2026-09-13: refreshed for the MongoDB retirement (TASK-0199). Replaced the server/db/ access rule with server/postgres/, named the PostgreSQL handlers that need the check, and switched the test workflow to postgres-test/server-route-test.
---

created: 2026-09-13 18:10
---
2026-09-13 review (review-task skill):

Findings:
- Relevance confirmed: postgresCreateEvent (server/routes/postgres_event_routes.go:1148) and postgresEditEvent (server/routes/postgres_event_routes.go:667) validate only non-empty names; no length cap exists.
- References verified: FR-119 exists, TASK-0136 is Done, both handlers exist, compose.test.yaml defines postgres-test and server-route-test, and server/db/ is gone.
- Important: the 100-character unit was unspecified; a byte-based len(name) would reject valid 100-character non-ASCII names and fail AC #2. Decision: count Unicode code points (runes), matching respondents.ValidateGuestName/MaxGuestNameLength (server/respondents/identity.go:12,63).
- Minor: AC #4 was conditional; swagger currently exposes POST /events as {201} and PUT /events/{eventId} as {200,403,404}. Decision: require @Failure 400 and regeneration.
- Minor: acceptance criteria now require non-ASCII boundary coverage and a no-persistence assertion.
- Clean dimensions: scope is one backend change, DoD matches project defaults, no active task modifies server/routes/postgres_event_routes.go, dependencies satisfied, FR-119 and terminology consistent.
- Scope: the uncapped sign-up event name field (frontend/src/components/NewSignUp.vue:476, :500; shared nameRules at frontend/src/composables/event/useEventEditorState.ts:229, excluded by TASK-0136) is out of scope for this backend task and tracked by TASK-0220.

Decisions applied: description and acceptance criteria updated; status, priority, assignee, and milestone unchanged.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Enforced the FR-119 100-code-point event-name cap on POST /events and PUT /events/{eventId}.

Changes:
- Added the stable errs.EventNameTooLong ("event-name-too-long") token and models.MaxEventNameLength = 100.
- postgresCreateEvent and postgresEditEvent now reject names whose utf8.RuneCountInString exceeds the cap with 400 and the token, before any repository work, so neither a new event nor an edit persists.
- Added @Failure 400 to both handler annotations and regenerated server/docs/docs.go, swagger.json, swagger.yaml, and frontend/src/types/api.ts (only the new 400 responses changed).
- Added server/routes/postgres_event_name_length_test.go with ASCII and non-ASCII coverage of the 101-rune rejection (including no-persistence assertions via a CREATE row count and a PUT re-read) and the 100-rune acceptance boundary on both create and update.

Validation:
- Focused route tests: 4 tests / 8 subtests PASS with -run 'NamesLongerThan100|NamesAt100'.
- Full isolated stack: server-route-test via compose.test.yaml -> all packages ok.
- Frontend: lint, fmt:check, typecheck, build, test:unit (146 files / 1093 tests) all pass.
<!-- SECTION:FINAL_SUMMARY:END -->
