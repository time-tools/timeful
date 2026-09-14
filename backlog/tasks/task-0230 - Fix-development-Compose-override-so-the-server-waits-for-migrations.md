---
id: TASK-0230
title: Fix development Compose override so the server waits for migrations
status: To Do
assignee: []
created_date: '2026-09-14 11:46'
updated_date: '2026-09-14 11:46'
labels:
  - compose
  - development
  - migrations
dependencies: []
references:
  - compose.development.yaml
  - compose.yaml
  - docs/environments.md
priority: medium
type: bug
ordinal: 232000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`compose.development.yaml` intends to specialize the development `server` service: depend only on the one-shot `postgres-migrate` service and mount only the `server_logs` volume. Both keys use `!reset` with a replacement value, and Docker Compose v2.36.2 resolves each to null. Evidence: `docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml config` reports `server.depends_on: null` and `server.volumes: null`.

Consequences observed 2026-09-14 while recreating the development database: the documented development command `docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml up --build postgres server` never starts `postgres-migrate`, so on a fresh database the server starts against an empty schema, crash-loops until the database happens to be migrated, and migrations must be run manually (`run --rm postgres-migrate`). The dev server also loses its `server_logs` volume mount.

Outcome: the development overlay resolves to exactly the intended `depends_on` (`postgres-migrate` with `service_completed_successfully`, and no `frontend-artifacts` dependency) and `volumes` (`server_logs:/app/logs`, and no `frontend_dist` mount), and the documented development command applies all pending migrations before the server starts on a fresh development volume. Base `compose.yaml` and the staging, production, and test overlays must resolve unchanged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Development resolved config (compose.yaml + compose.development.yaml) has server.depends_on equal to postgres-migrate with condition service_completed_successfully, and server.volumes equal to server_logs:/app/logs.
- [ ] #2 On a fresh development volume, the documented development command starts postgres, runs postgres-migrate to completion, and only then starts the server; all goose migrations apply with no manual migrate step.
- [ ] #3 Base compose.yaml and the staging, production, and test overlays resolve to their prior server dependencies and volumes.
- [ ] #4 Any docs describing the development command stay accurate or are updated to the verified behavior.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Fix lead (verified with a probe override on 2026-09-14): `!reset` clears the key to null; `!override` replaces the base value instead of merging. Changing both keys to `!override` in compose.development.yaml resolves to exactly the intended dev config:

```yaml
services:
  server:
    volumes: !override
      - server_logs:/app/logs
    depends_on: !override
      postgres-migrate:
        condition: service_completed_successfully
```

Probe result from `docker compose --env-file .env.development -f compose.yaml -f /tmp/opencode/compose.dev-fix-probe.yaml config`: `server.depends_on: {"postgres-migrate": {"condition": "service_completed_successfully", "required": true}}` and `server.volumes: [{"type": "volume", "source": "server_logs", "target": "/app/logs"}]`. The probe file was temporary and is not part of the repo.

Verification: compare `docker compose config --format json` server sections for dev (compose.yaml + compose.development.yaml), base compose.yaml, and the staging/production/test overlay combinations; then on a scratch project or after a volume backup, run the documented dev command against a fresh volume and confirm postgres-migrate exits successfully before the server starts and `goose_db_version` reaches `20260913000002`.
<!-- SECTION:PLAN:END -->
