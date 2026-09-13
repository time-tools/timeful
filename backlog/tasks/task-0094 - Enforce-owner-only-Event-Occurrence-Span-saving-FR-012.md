---
id: TASK-0094
title: Enforce owner-only Event Occurrence Span saving (FR-012)
status: To Do
assignee: []
created_date: '2026-08-28 20:35'
updated_date: '2026-09-13 15:05'
labels:
  - backend
  - authorization
  - postgres
dependencies: []
references:
  - server/routes/events.go
  - server/routes/postgres_event_routes.go
  - server/routes/postgres_owner.go
documentation:
  - docs/requirements/functional/fr/FR-012.md
  - docs/requirements/functional/fr/FR-018.md
priority: medium
type: bug
ordinal: 100000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
FR-012 (status: accepted; reworded in TASK-0092) restricts saving, replacing, or clearing the optional Event Occurrence Span to the Event Owner: "Other Event Visitors shall not be able to change it." The application does not enforce this today.

# Verified current behavior (re-verified 2026-09-13 after the MongoDB retirement)
- The PostgreSQL span handlers postgresSaveSchedule and postgresClearSchedule (server/routes/postgres_event_routes.go) delegate to postgresUpdateSchedule, which performs no owner authorization; postgresWritableEvent only rejects deleted or archived events.
- The legacy MongoDB handlers saveTimefulSchedule and clearTimefulSchedule no longer exist: server/db/, the MongoDB runtime, and the MongoDB event store were removed by TASK-0199.

# Scope
- Enforce owner-only Event Occurrence Span mutation for PostgreSQL events per FR-012, reusing the postgresOwnerMutation owner-authorization path established by TASK-0129 (server/routes/postgres_owner.go), following the FR-018 credential model (Event Owner Edit Token or the associated Platform Visitor Identity).
- Add route tests proving the owner-success and non-owner-rejection paths.

# Constraints
- Keep PostgreSQL access in server/postgres/; route handlers must not query the store directly (AGENTS.md backend conventions).
- Run route tests via the isolated Compose overlay: docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test, then run --rm server-route-test; the test database must be timeful-test or carry a timeful-test- prefix (AGENTS.md Server Test Workflow).
- Add Swag annotations if route signatures or responses change, then regenerate swagger and the frontend API types per AGENTS.md.
- This bug was discovered during TASK-0092's VERIFY step; that documentation task deliberately made no code change. FR-012's accepted status means the intended behavior is already the requirement; this task closes the enforcement gap.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Saving, replacing, or clearing an Event Occurrence Span on a PostgreSQL event succeeds only for the Event Owner and is rejected for every other Event Visitor, matching FR-012's owner-only authority (owner authorization follows the FR-018 credential model: the Event Owner Edit Token or the associated Platform Visitor Identity).
- [ ] #2 Route tests cover both the owner-success and non-owner-rejection paths for the span endpoints and pass via the isolated Compose test stack (postgres-test, server-test per compose.test.yaml).
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
created: 2026-09-13 15:05
---
2026-09-13: refreshed for the MongoDB retirement (TASK-0199). Removed the dead Mongo handler references and the legacy-Mongo-endpoints decision, dropped the obsolete Mongo AC, pointed the scope at the existing postgresOwnerMutation path, and switched the test workflow to the postgres-test/server-route-test stack.
---
<!-- COMMENTS:END -->
