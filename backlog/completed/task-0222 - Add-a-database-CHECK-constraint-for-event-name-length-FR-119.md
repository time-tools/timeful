---
id: TASK-0222
title: Add a database CHECK constraint for event name length (FR-119)
status: Done
assignee:
  - opencode
created_date: '2026-09-13 18:19'
updated_date: '2026-09-13 18:21'
labels:
  - backend
dependencies: []
references:
  - docs/requirements/functional/fr/FR-119.md
  - TASK-0137
  - server/migrations/20260912000000_baseline_schema.sql
priority: medium
type: task
ordinal: 225000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up to TASK-0137. PostgreSQL is the authoritative store for events, but postgres_events.name has no length guard; the FR-119 cap is enforced only in the create/update handlers.

Add defense-in-depth at the storage boundary:

- New incremental goose migration server/migrations/20260913000002_postgres_events_name_length.sql adding CHECK (name <> '' AND char_length(name) <= 100) NOT VALID, plus a Down that drops it. char_length counts Unicode code points, matching utf8.RuneCountInString and models.MaxEventNameLength.
- NOT VALID keeps legacy over-limit or empty rows from failing the migration while every new insert and update is checked. Legacy data cleanup and a later VALIDATE CONSTRAINT stay out of scope.
- Tests in server/postgres using the existing migration test harness (newMigrationTestRepository / expectSavepointError): constraint exists with convalidated = false, rejects empty and 101-code-point ASCII/non-ASCII names, accepts 100-code-point ASCII/non-ASCII names, and the Down migration drops the constraint.
- Verify with the isolated stack: run postgres-test-migrate to apply the migration to the real test DB, then run the full server-route-test suite.

Out of scope: cleaning existing rows and validating the constraint in deployed environments.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A new incremental goose migration adds postgres_events_name_length CHECK (name <> '' AND char_length(name) <= 100) NOT VALID with a matching Down drop.
- [x] #2 server/postgres tests prove the constraint exists with convalidated = false, rejects empty names and 101-code-point ASCII and non-ASCII names, accepts 100-code-point ASCII and non-ASCII names, and the Down migration drops it.
- [x] #3 The migration applies to the isolated test database via compose.test.yaml and the full server-route-test suite passes.
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
1. Add server/migrations/20260913000002_postgres_events_name_length.sql with the NOT VALID CHECK and Down drop.
2. Add server/postgres/name_constraint_test.go covering constraint presence (convalidated = false), rejections (empty, 101 ASCII, 101 non-ASCII), acceptance (100 ASCII, 100 non-ASCII), and the Down drop.
3. Apply the migration to the isolated test database with run --rm postgres-test-migrate, then run the full server-route-test suite.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation notes (2026-09-13):
- Migration: server/migrations/20260913000002_postgres_events_name_length.sql adds postgres_events_name_length CHECK (name <> '' AND char_length(name) <= 100) NOT VALID with a Drop in the Down section. The migration comment records the later operational step: after legacy rows are audited and cleaned, run ALTER TABLE postgres_events VALIDATE CONSTRAINT postgres_events_name_length.
- Tests: server/postgres/name_constraint_test.go uses newMigrationTestRepository and expectSavepointError. applyMigrationData rewrites CREATE TABLE to CREATE TEMP TABLE, so the new ALTER TABLE targets the transaction-scoped shadow table and never mutates shared test data. The Down test calls applyMigrationDown on the new file and then inserts a 101-code-point name to prove the guard is gone.
- Runtime note: if a future writer bypasses the handlers, a violation surfaces as raw SQLSTATE 23514 (constraint postgres_events_name_length), not the event-name-too-long API token; the API remains the primary gate.

Evidence:
- Focused: docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test go test ./postgres/ -run 'PostgresEventsNameLengthConstraint' -v -count=1 -> both tests and 5 subtests PASS.
- Migration applied to the real isolated test DB: docker compose ... run --rm --build postgres-test-migrate -> goose OK 20260913000002_postgres_events_name_length.sql; live query shows convalidated = f and CHECK (((name <> ''::text) AND (char_length(name) <= 100))) NOT VALID.
- Full suite: docker compose ... run --rm server-route-test -> every package ok (timeful/server/postgres 4.661s, timeful/server/routes 2.205s).
- No frontend, swagger, or Markdown changes, so gen:api, frontend checks, and format:markdown were not needed; no e2e was required.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the FR-119 event-name guard at the storage boundary.

Changes:
- server/migrations/20260913000002_postgres_events_name_length.sql: ALTER TABLE postgres_events ADD CONSTRAINT postgres_events_name_length CHECK (name <> '' AND char_length(name) <= 100) NOT VALID, with a Down that drops the constraint. char_length counts Unicode code points; NOT VALID keeps legacy rows from failing the migration while new inserts and updates are checked.
- server/postgres/name_constraint_test.go: proves the constraint exists with convalidated = false and both predicates, rejects empty and 101-code-point ASCII/non-ASCII names with SQLSTATE 23514, accepts 100-code-point ASCII/non-ASCII names, and verifies the Down migration drops the guard.

Validation:
- Applied to the real isolated test database with run --rm --build postgres-test-migrate (the migrator image bakes migrations, so a build was required); goose reported OK 20260913000002 and the live constraint shows convalidated = false with the expected definition.
- Focused package tests: TestPostgresEventsNameLengthConstraint (5 subtests) and TestPostgresEventsNameLengthConstraintDown PASS.
- Full isolated server-route-test suite: all packages ok, including timeful/server/postgres and timeful/server/routes.
- No frontend, swagger, or Markdown changes were needed.
<!-- SECTION:FINAL_SUMMARY:END -->
