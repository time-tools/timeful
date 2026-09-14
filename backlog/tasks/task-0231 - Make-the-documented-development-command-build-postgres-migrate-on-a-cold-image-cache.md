---
id: TASK-0231
title: >-
  Make the documented development command build postgres-migrate on a cold image
  cache
status: To Do
assignee: []
created_date: '2026-09-14 11:54'
labels:
  - compose
  - development
dependencies: []
references:
  - compose.development.yaml
  - compose.yaml
  - docs/environments.md
  - server/README.md
priority: medium
type: bug
ordinal: 233000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After TASK-0230, the development overlay's `server` depends on the one-shot `postgres-migrate` service, so `docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml up --build postgres server` now creates `postgres-migrate` as well. Compose `up --build` only builds images for the services named on the command line, so on a cold image cache (fresh clone or after pruning the `timeful-development-*` images) the command fails while creating that container: `Error response from daemon: No such image: timeful-development-postgres-migrate:latest`. The failure is loud and the server never starts, unlike the pre-TASK-0230 silent unmigrated server.

Reproduction (verified 2026-09-14 with Docker Compose v2.36.2 on a scratch Compose project): with no built `postgres-migrate` image, `up --build postgres server` builds only `server`, then fails at `Container ...-postgres-migrate-1 Creating` with the missing-image error. Adding `postgres-migrate` to the command's explicit service list, or documenting an explicit build step such as `docker compose ... build postgres-migrate`, resolves it.

Outcome: a developer with no pre-built development images can follow only the documented development commands from a fresh clone and get postgres, a completed `postgres-migrate`, and then the server, with no undocumented manual image build.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Following only the documented development commands from a cold image cache (no pre-built timeful-development-* images), the documented development command starts postgres, runs postgres-migrate to completion, and then starts the server with all goose migrations applied.
- [ ] #2 docs/environments.md and server/README.md describe the verified development flow accurately, including any explicit build step that is required.
- [ ] #3 On a fresh development volume the server still starts only after postgres-migrate exits successfully, preserving the TASK-0230 behavior.
- [ ] #4 Base compose.yaml and the staging, production, and test overlays resolve unchanged.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
