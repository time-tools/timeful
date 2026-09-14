---
id: TASK-0216
title: >-
  Rename the anonymous-event-compatibility route contract tests for the
  PostgreSQL-only system
status: To Do
assignee: []
created_date: '2026-09-12 22:08'
updated_date: '2026-09-14 19:01'
labels:
  - postgres
  - tests
dependencies:
  - TASK-0215
references:
  - server/routes/anonymous_event_compatibility_test.go
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
- [ ] #1 `server/routes/anonymous_event_compatibility_test.go` is renamed to a PostgreSQL event contract name
- [ ] #2 The file no longer claims that PostgreSQL can register into the contract later or that other event stores exist
- [ ] #3 `anonymousEventContractStores` and related helpers no longer present a multi-store abstraction over the single PostgreSQL store
- [ ] #4 Backend route tests still pass under the isolated Compose overlay
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
created: 2026-09-14 19:01
---
Scope is absorbed by TASK-0233.02 (parent TASK-0233): that subtask renames server/routes/anonymous_event_compatibility_test.go and its single-store helpers for the PostgreSQL-only system. Do not execute this task independently; mark it satisfied when TASK-0233.02 lands.
---
<!-- COMMENTS:END -->
