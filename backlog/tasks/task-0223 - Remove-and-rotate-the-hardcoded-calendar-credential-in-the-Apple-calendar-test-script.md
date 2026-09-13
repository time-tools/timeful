---
id: TASK-0223
title: >-
  Remove and rotate the hardcoded calendar credential in the Apple calendar test
  script
status: To Do
assignee: []
created_date: '2026-09-13 19:33'
labels:
  - security
  - cleanup
dependencies: []
references:
  - server/scripts/20240721_apple_calendar_test/main.go
modified_files:
  - server/scripts/20240721_apple_calendar_test/main.go
priority: high
type: bug
ordinal: 222000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
server/scripts/20240721_apple_calendar_test/main.go embeds an iCloud account email and app-specific password as string literals. The file is tracked in git, compiled by `go build ./...`, and was found during the post-migration quality audit. The credential is therefore exposed both in the working tree and in git history.

Outcome: no credential literal remains in the repository, and the leaked credential is revoked and replaced. The script is a 2024 manual CalDAV debugging tool with no callers, so the implementer may delete it or keep it with environment-provided values; either way the credential must be rotated as an operator action because history retains the old value.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 No hardcoded account email, password, or app-specific password remains in server/scripts/20240721_apple_calendar_test/; if the script is kept, its values come from environment variables and appear in no tracked file.
- [ ] #2 The leaked app-specific password is revoked and replaced, and the rotation is recorded in a task comment.
- [ ] #3 go build ./... and go vet ./... pass.
- [ ] #4 The task notes that git history still contains the old credential and that rotation is the mitigation.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
