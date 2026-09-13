---
id: TASK-0137
title: Validate event name length in the event create and update APIs (FR-119)
status: To Do
assignee: []
created_date: '2026-09-02 11:18'
updated_date: '2026-09-13 15:06'
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

- Reject event names longer than 100 characters on POST /events and PUT /events/:id with an appropriate 4xx response, without persisting the event or the change.
- Names of exactly 100 characters remain valid.
- Enforce the cap in the PostgreSQL event handlers postgresCreateEvent and postgresEditEvent (server/routes/postgres_event_routes.go), keeping the handler/service-layer conventions used by the existing event routes; PostgreSQL access stays in server/postgres/ (server/db/ no longer exists after the MongoDB retirement).
- Add route tests for both boundaries (over-limit rejection and exactly-100 acceptance) following the Server Test Workflow: start postgres-test via compose.test.yaml and run --rm server-route-test against the timeful-test database.
- If handler annotations change, regenerate swagger per Backend Conventions (swag init, then npm run gen:api).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The event create and update APIs reject an event name longer than 100 characters with a 4xx response and do not persist the change, per FR-119.
- [ ] #2 An event name of exactly 100 characters is accepted.
- [ ] #3 Route tests cover the over-limit rejection and the 100-character acceptance boundary and run in the isolated Compose test stack.
- [ ] #4 Swag annotations and regenerated swagger artifacts stay consistent if handler responses change.
- [ ] #5 Relevant server checks (go build/test or the scoped route suite) pass.
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
<!-- COMMENTS:END -->
