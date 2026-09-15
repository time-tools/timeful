---
id: TASK-0216
title: >-
  Rename the anonymous-event-compatibility route contract tests for the
  PostgreSQL-only system
status: Done
assignee: []
created_date: '2026-09-12 22:08'
updated_date: '2026-09-15 09:21'
labels:
  - postgres
  - tests
dependencies:
  - TASK-0215
references:
  - server/routes/anonymous_event_contract_test.go
  - server/docs/postgres-event-api-contract.md
priority: low
type: chore
ordinal: 221000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`server/routes/anonymous_event_compatibility_test.go` still carries the two-store migration framing: its `anonymousEventContractStore` doc comment says PostgreSQL can register once its route adapter exists, and `anonymousEventContractStores()` returns a one-element store slice even though PostgreSQL is the only store. Rename the file and helpers and update the comments so the regression suite describes the current PostgreSQL event API contract.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `server/routes/anonymous_event_compatibility_test.go` is renamed to a PostgreSQL event contract name
- [x] #2 The file no longer claims that PostgreSQL can register into the contract later or that other event stores exist
- [x] #3 `anonymousEventContractStores` and related helpers no longer present a multi-store abstraction over the single PostgreSQL store
- [x] #4 Backend route tests still pass under the isolated Compose overlay
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-14 19:01
---
Scope is absorbed by TASK-0233.02 (parent TASK-0233): that subtask renames server/routes/anonymous_event_compatibility_test.go and its single-store helpers for the PostgreSQL-only system. Do not execute this task independently; mark it satisfied when TASK-0233.02 lands.
---

author: opencode
created: 2026-09-15 09:21
---
Satisfied by TASK-0233.02 (Done): `anonymous_event_compatibility_test.go` became `anonymous_event_contract_test.go`, its three test functions dropped the Compatibility framing, the `anonymousEventContractStore`/`anonymousEventContractStores` slice was replaced by `anonymousEventRouter(t)` and `cleanupAnonymousEvent(t, eventID)`, and all seven call sites were updated. Evidence from the .02 finalization: isolated Compose `server-route-test` suite passes every package and the bundled `firefox-desktop` project passes 56 tests with 1 skip.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
TASK-0216 is satisfied by TASK-0233.02.

- `server/routes/anonymous_event_compatibility_test.go` is renamed to `anonymous_event_contract_test.go`, and its functions dropped the Compatibility framing.
- The file no longer claims PostgreSQL can register later or that other event stores exist; `anonymousEventContractStores` and the `anonymousEventContractStore` type are gone, replaced by `anonymousEventRouter(t)` and `cleanupAnonymousEvent(t, eventID)`.
- Backend route tests pass in the isolated Compose overlay (`server-route-test`), and the bundled Firefox desktop e2e project passes 56 tests with 1 skip.
<!-- SECTION:FINAL_SUMMARY:END -->
