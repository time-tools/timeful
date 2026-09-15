---
id: TASK-0198
title: >-
  Fix models.Event JSON round-trip so PostgreSQL payloads keep duration and
  legacy schedule fields
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-10 18:21'
updated_date: '2026-09-11 14:53'
labels:
  - postgresql
  - persistence
  - bug
dependencies: []
references:
  - >-
    backlog/tasks/task-0190.05.03 -
    Implement-PostgreSQL-group-response-mutation-and-calendar-derived-availability.md
  - server/models/event.go
  - server/routes/postgres_event_routes.go
  - server/routes/postgres_group.go
  - server/postgres/codec.go
priority: high
type: bug
ordinal: 219000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`models.Event.MarshalJSON` (server/models/event.go:125) wraps `Event` in an anonymous struct that embeds `*eventJSON` and re-declares `Duration`, `Dates`, `TimeIncrement`, `HasSpecificTimes`, `Times`, and `StartOnMonday`. The explicit wrapper fields shadow the promoted embedded fields, so those six fields always serialize from the nil wrapper values and are omitted for every event where `DaysOnly` is not true. Confirmed by marshaling a populated `models.Event` (`Duration`/`TimeIncrement` set): the JSON omits `duration` and `timeIncrement`.

This is not only an API-response concern: PostgreSQL event persistence uses the same custom marshaller. `postgresCreateEvent`/edit store `json.Marshal(event)` into `postgres_events.payload` (server/routes/postgres_event_routes.go:1015, and :563/:770), and reads decode it back with `json.Unmarshal` (server/postgres/codec.go:22). As a result the persisted payload loses `duration`, so `postgresMutateGroupResponse` computes the manual availability window from `eventModel.Duration` as zero (server/routes/postgres_group.go:490) and the legacy day-window replacement no longer spans the event duration. The route test seeded the payload directly, which hid the gap. `postgresEventPayload` (server/routes/postgres_event_routes.go:298) also marshals through this method for API responses.

Required outcome:
- Separate persistence encoding from the API-specific encoding so a PostgreSQL event payload round-trips all persisted fields, including `duration`, and live group creation/edit routes persist the manual availability window.
- Decide and document whether the six legacy schedule fields are intentionally omitted from the timed-event API; persistence must not lose fields regardless of that decision.
- Add regression coverage: a `models.Event` JSON round-trip test asserting `duration` (and the other persisted fields) survive marshal/unmarshal, and a route-level test that a live-created PostgreSQL group retains its manual availability window rather than defaulting to zero.
- Preserve the existing API contract and do not change browser-plugin `window.postMessage` payload shapes.

References:
- server/models/event.go:125 (MarshalJSON)
- server/routes/postgres_event_routes.go:298, 563, 770, 1015 (callers)
- server/postgres/codec.go:22 (decode)
- server/routes/postgres_group.go:490 (manual availability window)
- TASK-0190.05.03 Implementation Notes, "Discovered, not fixed (out of scope)"
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Marshaling a populated models.Event preserves duration, dates, timeIncrement, hasSpecificTimes, times, and startOnMonday values instead of dropping them.
- [x] #2 A PostgreSQL event payload written by live create/edit routes round-trips all persisted fields, and a live-created group's manual availability window uses the event duration rather than defaulting to zero.
- [x] #3 Regression tests cover the models.Event JSON round-trip and the live group manual availability window.
- [x] #4 The intended timed-event API contract for legacy schedule columns is explicitly decided and documented, and existing required frontend/backend checks pass.
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
Separate PostgreSQL persistence encoding from the API-specific event encoding, then finalize TASK-0190.

1. Decide and document the timed-event API contract for the six legacy schedule fields (duration, dates, timeIncrement, hasSpecificTimes, times, startOnMonday) in `server/models/event.go`.
2. Give persistence an encoding that never drops stored fields: either an explicit `MarshalPersistenceJSON`-style method or a route-local helper that marshals the unshadowed struct, and use it in `postgresCreateEvent` and the edit path so `postgres_events.payload` keeps `duration`.
3. Keep `postgresEventPayload` (API response) on the documented API encoding.
4. Add regression tests: a models-level persistence round-trip test asserting the six fields survive, and a route-level test that a live-created PostgreSQL group keeps its manual availability window.
5. Verify: isolated backend suite, frontend required checks if the API contract changes, Markdown formatting; then finalize TASK-0198 and TASK-0190 with evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (session 2026-09-11)

### Encoding split (AC#1, AC#4)

- `models.Event.MarshalJSON` is now the persistence encoding: it emits every persisted field, including `duration`, `dates`, `timeIncrement`, `hasSpecificTimes`, `times`, and `startOnMonday`, with `omitempty` on the six so nil columns stay absent and stored timed payloads keep their shape.

- `models.Event.MarshalAPIJSON` carries the previous timed-event API projection: it deliberately suppresses the six legacy columns for every event whose `daysOnly` flag is not true and keeps the null `attendees` key for non-group payloads.

- `postgresEventPayload` and `postgresDashboardEvent` now call `MarshalAPIJSON`, so the HTTP contract is unchanged while `json.Marshal(event)` at the create, edit, and schedule-persistence sites round-trips all fields.

- The contract decision is recorded in `server/docs/postgres-anonymous-event-compatibility.md` under JSONB Payloads.

### Live group duration (AC#2)

- `postgresGroupDurationHours` derives the availability-group duration from the canonical slot-generation window, wrapping past midnight like the editor.

- `postgresCreateEvent` sets it on every group create, and `postgresEditEvent` re-derives it on edit and preserves the stored duration when the edit omits generation settings.

- `postgresMutateGroupResponse` therefore computes the manual-availability day window from the event duration instead of zero.

### Regression coverage (AC#3)

- `server/models/event_test.go`: `TestEventMarshalJSONPreservesLegacyScheduleColumns` round-trips all six columns, and the suppression test is retargeted at `MarshalAPIJSON`.

- `server/routes/postgres_group_response_test.go`: `TestPostgresGroupLiveCreateDerivesManualAvailabilityWindow` creates a group through the live route, asserts the stored duration is 8, proves the 8-hour window replaces a stored 13:00 day with a 09:00 payload day, and asserts a live edit re-derives 3 hours from a 09:00-12:00 window.

### Verification

- `go build ./...` and `go vet ./...`: pass.

- Focused PostgreSQL route tests (group manual availability, group response lifecycle, signed-in lifecycle): pass.

- Full isolated backend suite (`docker compose ... run --rm server-route-test`): every package ok (log `/tmp/opencode/task-0198-backend.log`).

- Frontend `lint`, `fmt:check`, `typecheck`, `build`, `test:unit` (146 files, 1088 tests): pass.

- Firefox desktop e2e with `E2E_FRONTEND=bundled`: 56 passed, 1 skipped, 0 failed (log `/tmp/opencode/task-0198-e2e-firefox.log`).

- `npm run format:markdown:check`: pass; `graphify update .`: completed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Separated the PostgreSQL persistence encoding from the timed-event API projection and derived the availability-group duration from the canonical slot window.

## Changes
- `models.Event.MarshalJSON` now emits the full persisted payload, including the six legacy schedule columns when set; `models.Event.MarshalAPIJSON` keeps the previous timed-event API suppression, and the two API response sites (`postgresEventPayload`, `postgresDashboardEvent`) use it. The contract decision is documented in `server/docs/postgres-anonymous-event-compatibility.md`.
- `postgresGroupDurationHours` derives the group duration from the canonical slot-generation window and is applied on live create and edit, preserving the stored duration when an edit omits generation settings; `postgresMutateGroupResponse` now gets a real manual-availability window.

## Verification
- Models round-trip and live-group route regression tests pass, including the 8-hour window replacing a stored day and a 09:00-12:00 edit re-deriving 3 hours.
- Full isolated backend suite: every package ok; frontend lint/fmt/typecheck/build/unit pass; Firefox desktop e2e 56 passed, 1 skipped, 0 failed; Markdown check and graph update pass.

## Follow-up
TASK-0190 can now be finalized: its event-storage acceptance criterion no longer has the known payload-loss gap.
<!-- SECTION:FINAL_SUMMARY:END -->
