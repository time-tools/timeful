---
id: TASK-0214
title: Move remaining Go-side data logic into PostgreSQL
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-12 22:06'
updated_date: '2026-09-13 14:21'
labels: []
dependencies: []
priority: medium
type: task
ordinal: 214000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
# Context

The MongoDB driver and runtime are fully retired.
This task captures a 2026-09-13 audit of Go code that still performs set-based data manipulation in memory around PostgreSQL and could instead be expressed in SQL.
No Mongo driver imports remain; only dead compatibility artifacts are left.
Phases are tracked as subtasks and should land in order to minimize churn in shared files.

# Audit summary by phase

## Subtask 1 - read paths (no schema change)

- `server/routes/analytics.go:137-154` and `:224-237` issue one 30-day count query per calendar day in a Go loop; a single `generate_series`/`unnest` query can return the range.
- `server/postgres/dailylogs.go:133-160` pads missing reporting days by copying slices; `generate_series($1::date, $2::date, '1 day') LEFT JOIN daily_user_logs` returns padded days directly.
- `server/routes/postgres_event_routes.go:279-291` calls `visitor.controls` per response (one EXISTS per row); `server/routes/postgres_group.go:129-153` loads all responses for an existence test; `:81-95` scans attendees with EqualFold.
- `server/routes/respondent_identity.go:55-88` performs one account lookup per signup response; join `accounts` into `ListSignupResponses`.
- `server/postgres/calendar.go:113-134` adds identity and account-id pre-queries to nearly every calendar call; `:296-300` already shows the merged `UPDATE ... FROM` pattern.
- `server/postgres/transfers.go:78-84` issues two queries for one credential row; `server/postgres/identities.go:115-121` locks then re-reads the event.

## Subtask 2 - write paths (no schema change)

- `server/postgres/attendees.go:41-59`, `server/routes/postgres_group.go:208-257`, and `server/routes/postgres_event_routes.go:1121-1135` resolve accounts and write attendees row by row; set-based `unnest`/`INSERT ... SELECT` fits.
- `server/accounts/calendar.go:137-168` upserts sub-calendars one at a time, re-reads, then deletes one at a time.
- `server/postgres/repository.go:165-179` rewrites the full event row and payload after Go mutates `num_responses` (`server/routes/postgres_event_routes.go:826,854`, `server/routes/postgres_group.go:235,561,630`); targeted `num_responses = num_responses +- 1` updates are enough.
- `server/routes/postgres_owner.go:186-204` rewrites the full row to set one boolean.
- `server/postgres/folders.go:93-123` assembles a dynamic SET list in Go; a static `COALESCE` update works.

## Subtask 3 - signup block membership

- `event_signup_responses.block_ids TEXT[]` (`server/migrations/20260912000000_baseline_schema.sql:234-247`) has no FK and forces Go dedupe plus per-block `count(*)` in `server/postgres/signup.go:350-398`.
- `ReplaceSignupBlocks` (`server/postgres/signup.go:111-189`) loops per-block INSERT/UPDATE.
- A join table with FKs and set-based capacity/replace logic is the structural fix.

## Subtask 4 - indexes and typed UUID binds

- Missing indexes: `postgres_event_responses (event_id, created_at, id)`, `event_attendees lower(email)` and ordering, `postgres_events created_at` for analytics, `daily_user_log_members (daily_user_log_id, first_seen_position)`.
- `::text` casts on UUID columns defeat indexes: `server/postgres/signup.go:131,254,335,355,384`, `server/postgres/identities.go:58,87`, `server/postgres/transfers.go:35`, `server/postgres/accounts.go:263`.

## Subtask 5 - Mongo-era cleanup

- Dead `bson` tags at `server/services/calendar/google_calendar.go:37-39`.
- Unused helpers: `server/utils/response_utils.go`, legacy `Encode`/`Decode`/`Encrypt`/`Decrypt` and `EscapeRegExp` in `server/utils/utils.go`.
- `GuestNameObjectIDLike` naming at `server/respondents/identity.go:20` is a holdover.

# Constraints

- Keep PostgreSQL access in `server/postgres/`; routes and services must not access PostgreSQL directly.
- Add schema changes as new goose migrations in `server/migrations/`; the baseline is immutable.
- Preserve JSONB payload round-trip, absent-versus-null semantics, blind-availability privacy, EVCC authorization precedence, credential encryption, and guest-name normalization in Go.
- Verify with the isolated test overlay in `compose.test.yaml`.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All phase subtasks are complete and verified.
- [x] #2 No event, response, folder, signup, calendar, or transfer wire shape changes.
- [x] #3 New SQL lives in server/postgres/ and schema changes use goose migrations.
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
# Parent finalization plan

All five phase subtasks are Done and committed (`e23ddd09`..`b8eed8c1`). The parent has no remaining implementation scope, so this plan covers aggregate verification and finalization only.

## Findings so far

- TASK-0214.01 read paths, TASK-0214.02 write paths, TASK-0214.03 signup block join table, TASK-0214.04 indexes and typed UUID binds, and TASK-0214.05 Mongo-era cleanup all report complete with per-criterion evidence in their task records.
- A pre-commit review of TASK-0214.04 (findings in `backlog/handoffs/handoff-2026-09-13T09-51-42Z.md`) was fully fixed before commit; the residual-symbol checks show no `padDailyUserLogs`/`ListDailyUserLogs`, no `EqualFold` in `routes/postgres_group.go`, no `response_id::text`, no dead `Encode`/`Decode`/`Encrypt`/`Decrypt`/`EscapeRegExp`, no `bson` tags, and no SQL statements outside `server/postgres/`.
- Wire-shape diff `e23ddd09~1..HEAD`: only the `bson` tag removal in `services/calendar/google_calendar.go` changed any JSON tag line; no `server/docs` or frontend generated API diffs.
- Schema changes are the goose migrations `20260913000000_signup_response_blocks.sql` and `20260913000001_supporting_indexes.sql`.

## Steps

1. Run the full PostgreSQL-backed suite on HEAD with the isolated overlay.
2. Run the Firefox desktop E2E suite with `E2E_FRONTEND=bundled` (CI-equivalent per `e2e/AGENTS.md`) as aggregate behavior evidence for event, response, folder, signup, calendar, and transfer flows.
3. Check parent acceptance criteria and Definition of Done against that evidence, append notes, write the Final Summary, and set the parent Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verification baseline for every subtask:

- Create the external Go cache volumes once: `docker volume create timeful-test-go-build-cache timeful-test-go-mod-cache`.
- Start test PostgreSQL: `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test`.
- Run backend tests: `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test`.
- Browser E2E when a signup, event, folder, calendar, or transfer flow changes: `cd e2e && npm run test:e2e -- --project=firefox-desktop`.
- After code changes, run `graphify update .` to keep the knowledge graph current.

Parent finalization (2026-09-13):

- All five phase subtasks are Done and committed in order: TASK-0214.01 (e23ddd09), TASK-0214.02 (806ead25), TASK-0214.03 (dc530d1f), TASK-0214.04 (ac302aef), TASK-0214.05 (b8eed8c1). Each task record carries per-criterion evidence; the TASK-0214.04 review findings recorded in backlog/handoffs/handoff-2026-09-13T09-51-42Z.md were fixed and committed before this finalization.
- AC #1: verified all five subtask records (every acceptance criterion and DoD item checked with recorded evidence) and reran the full suite on HEAD.
- AC #2: `git diff e23ddd09~1..HEAD` shows no swagger annotation changes, no server/docs or frontend generated-API diffs, and no JSON tag changes other than the dead `bson` tag removal in server/services/calendar/google_calendar.go (the `json` tags themselves are unchanged). Residual-symbol checks find no leftover read/write-loop helpers, no EqualFold attendee scan, no response_id::text cast, and no dead MongoDB-era helpers.
- AC #3: a SQL-placement search finds no SELECT/INSERT/UPDATE/DELETE statements in routes, services, accounts, respondents, or utils; all new statements live in server/postgres/. Schema changes are the goose migrations 20260913000000_signup_response_blocks.sql and 20260913000001_supporting_indexes.sql.
- Backend suite: `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test` — all packages ok (postgres 6.6s, routes 3.8s, accounts, respondents, utils, and the rest ok).
- E2E: `E2E_FRONTEND=bundled npm run test:e2e -- --project=firefox-desktop` — 56 passed, 1 skipped, 0 failed (5.7m), covering access-transfer, calendar integration, group availability, signup blocks, owner authority, dashboard/folders, and visitor identities.
- Markdown: `npm run format:markdown:check` passes from the repo root; this finalization changed no repository Markdown files.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Closed out the multi-phase move of remaining Go-side data logic into PostgreSQL. All five phase subtasks landed in order and are verified together on HEAD.

### Phase outcomes

- **TASK-0214.01 — read paths** (`e23ddd09`): analytics day-spine queries, one `ListActiveUserDays` day-spine query, batched response authorization (`controlsBatch`), batched signup account profiles, `EventVisitorHasResponse` EXISTS, merged calendar identity liveness, single-query transfer credential, merged `LockEvent`.
- **TASK-0214.02 — write paths** (`806ead25`): set-based attendee upsert/delete and departed-response cleanup, set-based calendar sub-calendar sync, targeted `AdjustEventResponseCount`/`SetEventArchived`/`SetEventDeleted`, static `COALESCE` folder update.
- **TASK-0214.03 — signup block membership** (`dc530d1f`): `event_signup_response_blocks` join table with cascading FKs and position-ordered membership, set-based capacity checks and `ReplaceSignupBlocks`.
- **TASK-0214.04 — indexes and typed UUID binds** (`ac302aef`): supporting-indexes migration, typed UUID binds replacing `::text` casts, `HasNonDeclinedAttendeeEmail` closing the deferred attendee-email check, forced-plan tests running the production SQL constants.
- **TASK-0214.05 — Mongo-era cleanup** (`b8eed8c1`): dead `bson` tags and unused utils removed, `GuestNameAccountIDLike` rename.

### Verification

- `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test`: all packages pass.
- `E2E_FRONTEND=bundled npm run test:e2e -- --project=firefox-desktop`: 56 passed, 1 skipped, 0 failed (5.7m).
- Wire-shape audit of `e23ddd09~1..HEAD`: no swagger, generated-API, or JSON tag changes except the dead `bson` tag removal.
- SQL placement: no SQL outside `server/postgres/`; both schema changes are goose migrations.
- `npm run format:markdown:check` passes from the repo root.
<!-- SECTION:FINAL_SUMMARY:END -->
