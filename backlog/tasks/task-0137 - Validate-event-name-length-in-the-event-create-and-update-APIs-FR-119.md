---
id: TASK-0137
title: Validate event name length in the event create and update APIs (FR-119)
status: To Do
assignee: []
created_date: '2026-09-02 11:18'
updated_date: '2026-09-13 18:10'
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
- [ ] #1 The event create and update APIs reject an event name longer than 100 Unicode code points with a 400 response carrying the event-name-too-long token and do not persist the event or the change, per FR-119.
- [ ] #2 An event name of exactly 100 characters is accepted, including a non-ASCII name of 100 code points.
- [ ] #3 Route tests in the isolated Compose test stack cover the over-limit rejection, the 100-character acceptance boundary (including a non-ASCII case), and that a rejected change was not persisted.
- [ ] #4 POST /events and PUT /events/{eventId} document @Failure 400 and the regenerated server/docs/swagger.* and frontend/src/types/api.ts stay consistent with the handler responses.
- [ ] #5 The scoped server-route-test suite passes on the isolated Compose test stack.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

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
