---
id: TASK-0231
title: >-
  Make the documented development command build postgres-migrate on a cold image
  cache
status: Done
assignee:
  - opencode
created_date: '2026-09-14 11:54'
updated_date: '2026-09-14 12:17'
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
- [x] #1 Following only the documented development commands from a cold image cache (no pre-built timeful-development-* images), the documented development command starts postgres, runs postgres-migrate to completion, and then starts the server with all goose migrations applied.
- [x] #2 docs/environments.md and server/README.md describe the verified development flow accurately, including any explicit build step that is required.
- [x] #3 On a fresh development volume the server still starts only after postgres-migrate exits successfully, preserving the TASK-0230 behavior.
- [x] #4 Base compose.yaml and the staging, production, and test overlays resolve unchanged.
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
## Plan

1. Reproduce the cold-cache failure on an isolated scratch Compose project (`timeful-dev-verify`, host ports 15432/13002, fresh volume, no pre-built images): run `up --build postgres server` and confirm it fails creating `postgres-migrate` with `No such image`.
2. Test candidate fixes on the scratch project and pick the documented flow:
   - A: add `postgres-migrate` to the explicit service list: `up --build postgres postgres-migrate server`.
   - B: document an explicit build step: `build postgres-migrate` before `up --build postgres server`.
   Prefer the single-command flow if it starts postgres, completes migrations, and starts the server on both cold and warm caches.
3. Update the Development command blocks in `docs/environments.md` and `server/README.md`, and add a short note that Compose `up --build` builds only the services named on the command line, so `postgres-migrate` must stay in the list for a cold image cache.
4. Verify the documented command from a cold cache on a fresh volume: postgres healthy -> postgres-migrate exit 0 -> server starts; `/api/health` 200; `goose_db_version` max equals the latest migration; then verify a warm-cache re-run.
5. Confirm AC#4 by resolving base compose.yaml and the test/staging/production overlays with `docker compose config --quiet` (no compose files change, so their resolution is expected to be byte-identical in the server sections).
6. Run `npm run format:markdown` from the repo root for changed Markdown; docs-only change, so unit/e2e are exempt per the project DoD.
7. Record per-criterion evidence, update notes, and finalize the task.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation (2026-09-14): chose the explicit-service-list fix over a separate `build postgres-migrate` step because it is a single command and builds both required images on a cold cache, while a warm re-run simply rebuilds and re-runs the one-shot migration (goose is idempotent) without restarting the server. Updated `docs/environments.md` (Development block plus one note sentence) and `server/README.md` (command plus one note sentence). No compose files changed, so AC#4 holds trivially and was confirmed with `docker compose config --quiet` for base, dev, test, staging, and production (blank values in the staging/production example env files stubbed in scratch-only files under /tmp/opencode).

Verification environment note: this host has exhausted its default Docker address pools, so the scratch project used `/tmp/opencode/compose.dev-verify-net.yaml` with subnet 10.231.0.0/24. That override is scratch-only and not part of the repository. Isolated project name and ports (15432/13002) kept the running `timeful-development` and `timeful-test` stacks untouched; all scratch containers, volumes, network, and images were removed after verification.

Evidence: reproduction failure and fixed cold run in /tmp/opencode/task-0231-cold-run.log (both images Built, postgres Healthy -> postgres-migrate Exited -> server Started); health HTTP 200; goose max 20260913000002 equals newest migration; migrate finish 12:14:30.517Z before server start 12:14:30.907Z; format and lint Markdown commands exit 0. `graphify update .` ran after the docs change.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Updated the documented development command in `docs/environments.md` and `server/README.md` to name `postgres-migrate` explicitly:

`docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml up --build postgres postgres-migrate server`

Compose `up --build` builds images only for services named on the command line, so on a cold image cache the TASK-0230 dependency was created without ever building its image, failing with `No such image: timeful-development-postgres-migrate:latest`. Both docs now carry the corrected command plus a one-line note explaining why `postgres-migrate` must stay in the list. No compose files changed.

Verification (Docker Compose v2.36.2, isolated scratch project `timeful-dev-verify-0231` with alternate host ports, fresh volume, and a scratch-only network override because this host's default Docker address pools are exhausted):
- Before fix: `up --build postgres server` built only `server` and failed at `postgres-migrate` creation with the missing-image error.
- After fix, cold image cache (both scratch images deleted first): command exited 0; log shows `postgres-migrate Built` and `server Built`, then postgres Healthy -> postgres-migrate Started/Exited -> server Started. Server health returned HTTP 200 and `goose_db_version` max applied was `20260913000002`, matching the newest migration on disk.
- Fresh-volume ordering (TASK-0230 behavior preserved): postgres-migrate exited 0 at 12:14:30.517Z and the server started at 12:14:30.907Z.
- Warm-cache re-run: postgres-migrate re-ran and exited 0; the already-running server was left running.
- AC#4: `git diff` for compose.yaml and the development/staging/production/test overlays is empty; `docker compose config --quiet` passes for base, dev, test, staging, and production combinations.
- Markdown: `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` all exit 0. Docs-only change, so unit and e2e tests are exempt per the project Definition of Done.

Scratch containers, volumes, network, and images were removed afterward; the running development and test stacks were not touched.
<!-- SECTION:FINAL_SUMMARY:END -->
