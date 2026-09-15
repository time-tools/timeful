---
id: TASK-0230
title: Fix development Compose override so the server waits for migrations
status: Done
assignee:
  - opencode
created_date: '2026-09-14 11:46'
updated_date: '2026-09-14 11:55'
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
- [x] #1 Development resolved config (compose.yaml + compose.development.yaml) has server.depends_on equal to postgres-migrate with condition service_completed_successfully, and server.volumes equal to server_logs:/app/logs.
- [x] #2 On a fresh development volume, the documented development command starts postgres, runs postgres-migrate to completion, and only then starts the server; all goose migrations apply with no manual migrate step.
- [x] #3 Base compose.yaml and the staging, production, and test overlays resolve to their prior server dependencies and volumes.
- [x] #4 Any docs describing the development command stay accurate or are updated to the verified behavior.
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation notes (2026-09-14): compose.development.yaml now uses `!override` for both `server.volumes` and `server.depends_on`, replacing the inherited values rather than clearing them to null with `!reset`.

Resolved-config evidence: dev server.depends_on = {"postgres-migrate": {"condition": "service_completed_successfully", "required": true}} and server.volumes = [{"type": "volume", "source": "server_logs", "target": "/app/logs"}]. Base, test, staging, and production overlays still resolve to frontend-artifacts + postgres-migrate dependencies and frontend_dist:ro + server_logs volumes; compose.development.yaml is not part of those combinations.

Live verification on scratch project `timeful-dev-verify` with alternate host ports 15432/13002 left the running development stack untouched. Compose log order was postgres Healthy -> postgres-migrate Started -> Exited -> server Started. postgres-migrate exit=0 and finished 11:51:24.535Z; server started 11:51:24.945Z with RestartCount=0; /api/health returned {"status":"ok"} (HTTP 200); goose_db_version max = 20260913000002; the server container mounted only server_logs at /app/logs. Scratch stack, volumes, and images were removed afterward.

Cold image-cache limitation found during verification (outside this task's acceptance criteria): `up --build postgres server` does not build the newly required postgres-migrate dependency image, so the documented command fails with `No such image: ...postgres-migrate:latest` on a cold image cache. Follow-up filed with user approval as TASK-0231.

No unit or e2e tests cover the development Compose overlay; verification used `docker compose config` comparisons across all overlay combinations plus the live scratch-stack run. No Markdown files were changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Changed `compose.development.yaml` to use `!override` instead of `!reset` for `server.volumes` and `server.depends_on`, so the development overlay replaces the base values instead of resolving both keys to null. Docker Compose v2.36.2 now resolves development `server` to depend only on `postgres-migrate` (`service_completed_successfully`) and to mount only `server_logs:/app/logs`, with no `frontend-artifacts` dependency and no `frontend_dist` mount.

Verification:
- `docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml config --format json` resolves dev `server.depends_on` to `{"postgres-migrate": {"condition": "service_completed_successfully", "required": true}}` and `server.volumes` to `server_logs:/app/logs`.
- Base `compose.yaml` and the test, staging, and production overlays resolve unchanged: `frontend-artifacts` plus `postgres-migrate` dependencies and `frontend_dist:ro` plus `server_logs` volumes.
- Live run on an isolated scratch project with fresh volumes and alternate host ports: postgres became healthy, `postgres-migrate` exited 0 after applying migrations `20260912000000` through `20260913000002`, and only then the server started (migrate finished `11:51:24.535Z`; server started `11:51:24.945Z`), with `/api/health` returning HTTP 200, `RestartCount=0`, and only the `server_logs` mount.
- Follow-up discovered during verification and filed with user approval: TASK-0231 covers the documented development command on a cold image cache, where `up --build postgres server` does not build the newly required `postgres-migrate` dependency image.

No Markdown files changed, and no unit or e2e tests cover the development Compose overlay; verification used resolved-config comparison for every overlay plus the live scratch-stack run.
<!-- SECTION:FINAL_SUMMARY:END -->
