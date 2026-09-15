---
id: TASK-0233
title: 'Remove the migration-era postgres qualifier from schema, server, e2e, and docs'
status: To Do
assignee: []
created_date: '2026-09-14 19:00'
labels:
  - cleanup
  - postgres
  - naming
  - schema
dependencies: []
references:
  - server/migrations/20260912000000_baseline_schema.sql
  - docs/postgres-operations.md
  - server/docs/postgres-event-api-contract.md
  - backlog/backlog.md
documentation:
  - docs/terminology/glossary.md
  - docs/postgres-operations.md
priority: medium
type: chore
ordinal: 232000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
MongoDB is retired and PostgreSQL is the sole authoritative store, but the migration-era "postgres" qualifier survives in the schema (postgres_events, postgres_event_responses), in route-level Go symbols and file names, in e2e spec and helper names, and in live docs and generated API text. It implies an alternative store that no longer exists and is inconsistent with the rest of the schema, which uses unprefixed domain names.

Outcome: the event tables are events and event_responses; related constraints and indexes drop the postgres_ prefix; migration files, Go/e2e file names, and internal identifiers drop the qualifier; live docs, swagger output, and frontend prose use the new names. Genuine PostgreSQL product and infrastructure names stay.

Confirmed decisions:
- Names: postgres_events -> events, postgres_event_responses -> event_responses; constraint and index names drop the postgres_ prefix (for example postgres_events_name_length -> events_name_length); postgres_response_owner_event_fk -> event_responses_visitor_event_fk.
- Migration strategy: edit the baseline and existing incremental migrations in place (no new rename migration), rename migration files without changing goose version prefixes, and recreate development, isolated test, staging, and production databases from the edited chain. Back up staging and production first per docs/postgres-operations.md and retain the backups.
- Keep: the server/postgres package and pgstore alias, POSTGRES_* variables, compose services (postgres, postgres-test, postgres-migrate, postgres_data), psql, pgx/pgconn usage, and the postgresError SQLSTATE wrapper. Do not substitute "store" for "postgres".
- Historical records stay as written: docs/design/architecture/adr/**, backlog/handoffs/**, backlog/completed/**, and graphify-out/**.
- Out of scope: docs/requirements/** wording (requirements workflow), glossary alignment (TASK-0087), and TASK-0228 owns the stale anonymous-event-creation flag sentence in docs/ci.md. The anonymous-event-compatibility route test rename (TASK-0216) is absorbed by the Go subtask.

Risks: because applied migrations are edited rather than versioned by content, any database left unrecreated silently keeps the old schema while new code expects the new names; pre-rename server binaries are incompatible with the renamed schema, so code and migration must ship together and rollback requires restoring a backup plus the previous release. Recreating staging and production destroys data accrued since the last recreate unless restored from backup.

See subtasks for the per-layer work orders.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The schema, migration chain, and all four recreated databases expose only events and event_responses with no postgres-prefixed relation, constraint, or index name.
- [ ] #2 No postgres-prefixed Go or e2e file name remains, and domain symbols and test titles no longer carry the qualifier.
- [ ] #3 Live docs, swagger artifacts, and generated frontend API types use the new object names, and hand-written frontend prose no longer qualifies events or responses as PostgreSQL-specific.
- [ ] #4 Legitimate PostgreSQL product and infrastructure naming is preserved as listed in the confirmed decisions.
- [ ] #5 Every required check passes for the touched areas, including the isolated backend suite, the Firefox desktop e2e project, and the frontend required checks.
- [ ] #6 Historical ADRs, handoffs, completed task records, and requirements records are unchanged.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
