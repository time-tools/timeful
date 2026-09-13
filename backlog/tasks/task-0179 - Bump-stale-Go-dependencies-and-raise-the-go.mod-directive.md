---
id: TASK-0179
title: Bump stale Go dependencies and raise the go.mod directive
status: To Do
assignee: []
created_date: '2026-09-08 12:11'
updated_date: '2026-09-13 15:06'
labels:
  - server
dependencies: []
references:
  - server/go.mod
  - server/Dockerfile
  - server/utils/utils.go
priority: medium
type: chore
ordinal: 185000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
server/go.mod declares `go 1.20` while the Docker toolchain is golang:1.26.4-alpine (server/Dockerfile:5), and several modules are years stale (2026-09-08 audit; dependency list re-checked 2026-09-13). Raise the directive and bump stale modules; replace unmaintained ones or record why they stay. No route or behavior change.

Notable stale or unmaintained direct dependencies:
- github.com/brianvoe/sjwt v0.5.1 (JWT handling in server/utils/utils.go)
- github.com/gin-gonic/gin v1.9.1 plus gin-contrib/cors and gin-contrib/sessions
- github.com/bwmarrin/discordgo v0.27.1
- github.com/swaggo/files v1.0.1 and swaggo/gin-swagger v1.6.0
- github.com/jonyTF/go-webdav v0.5.2 (fork) and github.com/emersion/go-ical pinned to a 2024 pseudo-version

Already retired and no longer in server/go.mod: go.mongodb.org/mongo-driver (removed by TASK-0199.09.03) and gopkg.in/gomail.v2, whose server/utils/mail_utils.go was removed when OTP mail moved to Listmonk.

Bump within API compatibility. If swag changes, keep `swag init --parseDependency` and `npm run gen:api` output stable or update them together. Verify with the isolated PostgreSQL test stack (compose.test.yaml overlay) rather than a host database.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The go.mod go directive is raised to the language level supported by the Docker toolchain (no older than golang:1.26.4 features)
- [ ] #2 Stale direct dependencies are bumped to current stable, with explicit recorded keep-decisions for anything not bumped (the sjwt decision must be explicit)
- [ ] #3 go build and go vet are clean; if swag is bumped, swag regen plus frontend npm run gen:api produce no unintended diffs
- [ ] #4 Route tests pass via the isolated compose.test.yaml PostgreSQL stack (server-route-test), with no route or behavior change
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
created: 2026-09-13 15:06
---
2026-09-13: refreshed for the MongoDB retirement (TASK-0199). Removed mongo-driver and gomail from the stale-dependency list (both already retired), dropped the dead mail_utils.go reference, and replaced the Mongo test-stack verification with the postgres-test/server-route-test workflow.
---
<!-- COMMENTS:END -->
