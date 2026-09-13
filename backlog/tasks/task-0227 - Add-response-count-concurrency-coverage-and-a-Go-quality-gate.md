---
id: TASK-0227
title: Add response-count concurrency coverage and a Go quality gate
status: To Do
assignee: []
created_date: '2026-09-13 19:36'
labels:
  - postgres
  - tests
  - ci
dependencies: []
references:
  - server/postgres/repository.go
  - server/routes/auth_otp_test.go
  - .github/workflows/backend-ci.yml
priority: medium
type: chore
ordinal: 229000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The response-count adjustments that back num_responses are only tested sequentially, and backend CI runs go test ./... -count=1 without the race detector or any Go static gate. The post-migration quality audit also found pre-existing gofmt drift in server/routes/auth_otp_test.go.

Outcome: race-detector coverage for the concurrency-sensitive response-count path and a Go quality gate that catches formatting and vet regressions in CI. This is distinct from TASK-0179 (dependency bumps); do not change dependencies here.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A test races concurrent response creation and deletion for one event and asserts num_responses equals the number of stored responses afterwards.
- [ ] #2 Backend CI runs the affected packages with the race detector, or the full server suite with -race, with the chosen scope and runtime recorded.
- [ ] #3 gofmt -l server/ is clean, including the current drift in server/routes/auth_otp_test.go.
- [ ] #4 Backend CI fails on gofmt drift and on go vet findings, or the repository records why no static gate is configured.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
