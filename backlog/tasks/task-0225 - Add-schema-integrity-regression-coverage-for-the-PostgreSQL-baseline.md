---
id: TASK-0225
title: Add schema-integrity regression coverage for the PostgreSQL baseline
status: To Do
assignee: []
created_date: '2026-09-13 19:35'
labels:
  - postgres
  - tests
dependencies: []
references:
  - server/postgres/migration_helpers_test.go
  - server/postgres/baseline_schema_test.go
  - server/migrations/20260912000000_baseline_schema.sql
modified_files:
  - server/postgres/
priority: medium
type: task
ordinal: 227000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The baseline schema is the single authority for data integrity, but several of its constraints have no regression test, event hard-delete cascade behavior is untested, and the server/postgres test harness replays migration files into TEMP tables with a literal \"CREATE TABLE \" string replacement that can silently exercise the wrong schema. Outcome: tests that fail when a baseline invariant or the harness's schema fidelity regresses.

Exclusions: the FR-119 event-name constraint lifecycle and its convalidated assertion are owned by TASK-0224.02; coordinate rather than duplicating that coverage.

Useful references: server/migrations/20260912000000_baseline_schema.sql, server/postgres/baseline_schema_test.go, server/postgres/migration_helpers_test.go, server/postgres/supporting_indexes_test.go, compose.test.yaml.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Tests assert the currently untested baseline CHECK constraints, at least: postgres_events payload object type and short-id format, access_transfers source/platform-identity XOR, otp_challenges empty-value and attempts checks, postgres_event_responses payload object type, and calendar_preferences constraint behavior.
- [ ] #2 A test deletes a postgres_events row and asserts every dependent row cascades: visitor identities, visitor credentials, responses, signup blocks, signup responses, signup response memberships, attendees, folder_events, and access transfers.
- [ ] #3 A test asserts the application role can perform DML but not DDL, or the equivalent privilege separation is verified against the isolated stack.
- [ ] #4 The temp-table migration replay used by server/postgres tests is checked against the goose-applied schema, or the harness is changed so any divergence fails a test.
- [ ] #5 All new tests run in the isolated Compose overlay and pass.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
