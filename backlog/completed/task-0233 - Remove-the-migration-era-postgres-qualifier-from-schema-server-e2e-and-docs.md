---
id: TASK-0233
title: 'Remove the migration-era postgres qualifier from schema, server, e2e, and docs'
status: Done
assignee: []
created_date: '2026-09-14 19:00'
updated_date: '2026-09-15 12:08'
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
- [x] #1 The schema, migration chain, and all four recreated databases expose only events and event_responses with no postgres-prefixed relation, constraint, or index name.
- [x] #2 No postgres-prefixed Go or e2e file name remains, and domain symbols and test titles no longer carry the qualifier.
- [x] #3 Live docs, swagger artifacts, and generated frontend API types use the new object names, and hand-written frontend prose no longer qualifies events or responses as PostgreSQL-specific.
- [x] #4 Legitimate PostgreSQL product and infrastructure naming is preserved as listed in the confirmed decisions.
- [x] #5 Every required check passes for the touched areas, including the isolated backend suite, the Firefox desktop e2e project, and the frontend required checks.
- [x] #6 Historical ADRs, handoffs, completed task records, and requirements records are unchanged.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Finalization verification (2026-09-15)

- Databases: development and isolated test were recreated by TASK-0233.01; staging and production were recreated by TASK-0233.01.01 on release 03c34a37 from the edited chain (goose through 20260915000000). Catalog queries show only events and event_responses, zero postgres-prefixed relations/constraints/indexes, and events_name_length convalidated = t in all four databases.
- Backups: pre-recreate custom-format dumps for staging and production were restored into scratch databases and reconciled by per-table counts and md5 row-content digests (MATCH), then retained at ~/timeful-backups/renamed-staging-20260915T115455Z/ and ~/timeful-backups/renamed-production-20260915T115506Z/.
- Residual gates on the current tree: no live reference to postgres_events or postgres_event_responses outside docs/design/architecture/adr/**, backlog/handoffs/**, backlog/completed/**, graphify-out/**, and the backlog workflow records; no postgres-prefixed Go or e2e file name outside the kept server/postgres package (server/docs/postgres-event-api-contract.md and server/docs/postgres-data-boundaries.md keep qualified titles as recorded in TASK-0233.04).
- Historical records: `git diff --name-only 5c57b4b5^..e1d279d5 -- docs/design/architecture/adr backlog/handoffs backlog/completed graphify-out` is empty.
- Preserved infrastructure naming: server/postgres package and pgstore alias referenced by 37 Go files, POSTGRES_* variables in 12 Go files, compose postgres/postgres-test/postgres-migrate services and postgres_data volumes unchanged.
- Checks: isolated Compose backend suite passed in TASK-0233.01 (twice) and TASK-0233.04; bundled firefox-desktop e2e passed with only the documented access-transfer 30s-budget flakes (accepted by the user, passing in isolation); frontend lint/fmt:check/typecheck/build/test:unit passed in TASK-0233.04; root `npm run format:markdown:check` exit 0 on the current tree.
- Deployment: deployed revision 03c34a37e981c1953138e6f52816030d7d9c3b43 on branch main; production /api/health {"status":"ok"} plus HTTP 200 on https://timeful.fun/ and https://timeful.fun/api/health; staging verified healthy on both the local binding and https://staging.timeful.fun/api/health before being stopped per user approval.

No repository changes were made during this finalization session; only Backlog records changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
TASK-0233 is complete: the migration-era postgres qualifier is removed from the schema, server, e2e, and live docs, and all four databases run the renamed chain.

Schema and databases:
- The baseline and incremental migrations create events and event_responses with unprefixed constraints and indexes (including event_responses_visitor_event_fk); the three postgres-named migration files were renamed with unchanged goose version prefixes. Development and isolated test were recreated by TASK-0233.01.
- TASK-0233.01.01 recreated staging and production on release 03c34a37 after verified custom-format backups (staging sha256 be40d713..., production sha256 20b97eeb..., both scratch-restore reconciled and retained under ~/timeful-backups/), with the user-approved no-carry-over decision recorded. Both databases apply goose through 20260915000000, expose only events/event_responses, contain zero postgres-prefixed relations/constraints/indexes, and have events_name_length validated.

Code and text:
- TASK-0233.02 renamed the route/test files and domain symbols and removed redundant qualifiers while preserving the server/postgres package, pgstore, POSTGRES_*, pgx/pgconn, and the postgresError wrapper; TASK-0233.03 renamed the e2e specs and helper and switched their SQL to the new tables; TASK-0233.04 dropped the two PostgreSQL-only swag descriptions, regenerated server/docs and frontend/src/types/api.ts through the documented commands, and cleaned the hand-written frontend prose.

Verification:
- Residual greps find no live postgres_events/postgres_event_responses reference outside the allowed historical paths and no postgres-prefixed Go or e2e file name outside the server/postgres package; the rename commits 5c57b4b5^..e1d279d5 touched no ADR, handoff, completed-task, or graphify path.
- The isolated Compose backend suite passed in TASK-0233.01 and TASK-0233.04; the bundled firefox-desktop project passed 54-56 tests with the documented access-transfer 30s-budget flakes accepted by the user and passing in isolation; frontend lint, fmt:check, typecheck, build, and test:unit passed in TASK-0233.04; npm run format:markdown:check passes on the current tree.
- Production and staging /api/health are healthy on the deployed revision 03c34a37 (production also HTTP 200 on https://timeful.fun/ and /api/health; staging verified healthy, then stopped for host memory).
- Follow-ups outside this cleanup remain owned elsewhere: TASK-0228 owns the docs/ci.md stale anonymous-event-creation flag sentence, and the PostgreSQL-qualified titles of server/docs/postgres-event-api-contract.md and postgres-data-boundaries.md stay as recorded observations.
<!-- SECTION:FINAL_SUMMARY:END -->
