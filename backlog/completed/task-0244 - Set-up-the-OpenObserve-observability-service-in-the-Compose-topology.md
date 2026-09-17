---
id: TASK-0244
title: Set up the OpenObserve observability service in the Compose topology
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-16 12:35'
updated_date: '2026-09-16 17:25'
labels: []
dependencies:
  - TASK-0243
references:
  - 'https://github.com/openobserve/openobserve'
  - 'https://openobserve.ai/docs/architecture/'
  - >-
    https://openobserve.ai/docs/administration/configuration/environment-variables/
documentation:
  - docs/design/architecture/adr/ADR-022.md
  - docs/requirements/quality/qr/QR-015.md
  - docs/requirements/quality/qr/QR-016.md
  - docs/requirements/quality/qr/QR-017.md
  - docs/requirements/quality/qr/QR-018.md
  - docs/requirements/quality/qr/QR-019.md
  - docs/environments.md
  - DEPLOYMENT.md
  - compose.yaml
  - compose.development.yaml
  - compose.edge.yaml
priority: medium
type: task
ordinal: 245000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Set up the OpenObserve observability platform decided in ADR-022 as running infrastructure with its operational contract documented and verified. This task does not instrument the Go server; log shipping and then metrics and traces are separate follow-up tasks.

Layout: the shared staging and production instance runs as its own Compose project on the deployment host, with a shared external network and a persistent data volume, separate from the timeful-staging and timeful-production app stacks. Development runs its own instance in the local development stack.

Constraints from ADR-022 and QR-015 through QR-019: open-source edition, single-node mode, local-disk storage, image pinned by digest to v1.0.1, which supersedes the originally recorded v0.92.2 release; loopback-only listener with no Caddy route and SSH-tunnel UI access; explicit memory and disk caps for the memory cache, query engine memory pool, memtable, and disk cache instead of the memory-consuming defaults; anonymous telemetry disabled; retention of 14 days for every stream; authenticated ingest with per-environment organizations, streams, and environment-scoped credentials.

The environment contract must be updated together: .env.development.example, .env.staging.example, .env.production.example, docs/environments.md, and DEPLOYMENT.md. Compose has no application-value fallbacks, so every interpolated variable must be declared in the selected env file. The isolated test stack must remain unaffected.

Governing requirements: QR-015, QR-016, QR-017 (verification lands with the log-shipping task), QR-018, QR-019. Decision record: ADR-022.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 OpenObserve runs in the documented Compose topology with the image pinned by digest to v1.0.1, single-node local-disk storage, and a loopback-only listener, and no Caddy route or public port exposes it
- [x] #2 Staging and production share one OpenObserve instance that runs as its own Compose project on the deployment host with a persistent data volume, and neither app stack runs a second observability instance
- [x] #3 Development runs its own OpenObserve instance in the local development stack with its own data volume, and development data is never sent to the shared instance
- [x] #4 Explicit limits are configured for the memory cache, the query engine memory pool, the memtable, and the disk cache, and anonymous telemetry is disabled (QR-015)
- [x] #5 Retention is configured to 14 days for every stream and is not left unlimited (QR-018)
- [x] #6 A connection attempt from a public interface to the OpenObserve port fails, an unauthenticated or incorrectly authenticated ingest request is rejected, and an ingest request with valid environment credentials succeeds (QR-016)
- [x] #7 Read and ingest attempts that use staging credentials against the production organization, or production credentials against the staging organization, are rejected, while each credential succeeds against its own environment (QR-019)
- [x] #8 .env.development.example, .env.staging.example, .env.production.example, docs/environments.md, and DEPLOYMENT.md document the OpenObserve variables, per-environment organization and credential provisioning, resource caps, retention, and SSH-tunnel UI access
- [x] #9 The isolated test stack is unaffected: compose.test.yaml, server-test, and server-route-test need no OpenObserve service or variables, and the documented development, staging, and production Compose commands still validate
- [x] #10 ADR-022 records the image pinned by digest to v1.0.1 and no longer treats the 1.x line as an unreviewed future upgrade
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
1. Revise ADR-022 to pin the digest-resolved v1.0.1 release instead of v0.92.2, replace the 1.x compatibility caveat, and align TASK-0244's description and acceptance criteria.
2. Add compose.observability.yaml for the shared staging and production instance: its own project, external timeful-edge network, persistent data volume, loopback-only 5080, explicit caps, telemetry disabled, and 14-day retention.
3. Add the development openobserve service to compose.development.yaml with its own data volume and the same caps, and declare the server-side OpenObserve variables in the development, staging, and production app overlays.
4. Add .env.observability.example and the OpenObserve variables to .env.development.example, .env.staging.example, and .env.production.example; ignore .env.observability.
5. Document the topology, variables, organization and credential provisioning, resource caps, retention, and SSH-tunnel UI access in docs/environments.md and DEPLOYMENT.md.
6. Verify with compose config for every documented stack, run the development instance, exercise QR-016 (public-interface refusal, unauthenticated rejection, valid ingest) and QR-019 (cross-organization rejection), inspect caps and retention, and confirm the isolated test stack still configures.
7. Run the root Markdown checks, record the evidence, and finalize the task.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Review findings (corrections made during verification)

- OpenObserve v1.0.1 rejects hyphens in organization names (letters, digits, spaces, and underscores only), so the environment organizations are `timeful_development`, `timeful_staging`, and `timeful_production` rather than the hyphenated names recorded during implementation.
- v1.0.1 generates a random organization identifier at creation and ignores an identifier supplied in the create request. Environment credentials (organization users and service accounts) authenticate only against that generated identifier, not the organization display name; only the root user accepts both. The server contract variable was therefore renamed to `OPENOBSERVE_ORGANIZATION_ID`, and the env examples leave it blank for the operator to fill after provisioning with a comment explaining the identifier requirement.
- The open-source edition accepts only `admin` and `service_account` roles when creating an organization user, so the provisioning documentation now creates a service account for ingestion. OpenObserve returns the service-account token only at creation; the service-account list masks it afterward. `OPENOBSERVE_INGEST_USERNAME` holds the service-account email and `OPENOBSERVE_INGEST_PASSWORD` holds the token.

## QR-016 exercise (running development instance `timeful-development-openobserve-1`)

- Loopback-only listener: `ss -tlnp` shows `127.0.0.1:5080` only, `docker port` shows `127.0.0.1:5080->5080/tcp`, and `curl http://10.242.1.84:5080/healthz` from the host LAN interface fails with connection refused (curl exit 7, HTTP 000).
- Unauthenticated `POST /api/<org>/qr016_stream/_json` returns 401; a wrong service-account token returns 401.
- A valid service-account credential ingests with HTTP 200 (`{"code":200,"status":[{"name":"qr016_stream","successful":1,"failed":0}]}`), and `select * from "qr016_stream"` returns the record (total 1).
- The shared `compose.observability.yaml` instance uses the same image pin, loopback publish, local-disk single-node settings, and caps.

## QR-019 exercise (two organizations on the development instance)

- Created `timeful_development` (identifier `3JQ1a8basz7mdePUqGq6P7jk0ot`) and `timeful_production` (identifier `3JQ1a8a5CLpxK6HMmW9hIMn30jA`), each with its own service account.
- Own-organization ingests succeed: development 200, production 200.
- Cross-organization ingests are rejected: development credential to production organization 401, production credential to development organization 401.
- Cross-organization reads are rejected: each credential's `_search` against the other organization returns 401, while its own-organization `_search` returns 200 with the ingested record.
- The development instance substitutes for the shared instance, which does not run locally; OpenObserve performs the same organization authorization in both cases.

## Caps, retention, and telemetry (development instance)

- Container environment: `ZO_MEMORY_CACHE_MAX_SIZE=128`, `ZO_MEMORY_CACHE_DATAFUSION_MAX_SIZE=256`, `ZO_MEM_TABLE_MAX_SIZE=128`, `ZO_DISK_CACHE_MAX_SIZE=1024`, `ZO_COMPACT_DATA_RETENTION_DAYS=14`, `ZO_TELEMETRY=false`, `ZO_LOCAL_MODE=true`, `ZO_LOCAL_MODE_STORAGE=disk`.
- Startup log confirms the effective caps: `Caches info: Disk max size 1.00 GB, MEM max size 128.00 MB, Datafusion pool size: 256.00 MB`.
- The ingested stream's settings show `data_retention: 0`, which inherits the global 14-day default from `ZO_COMPACT_DATA_RETENTION_DAYS` (OpenObserve default 3650 days, minimum 3), so no stream is unlimited at the default. The actual 14-day deletion cannot be observed within the session; the configuration plus the documented inherit semantics are the evidence.

## Compose validation

- Development, staging, production, observability, and isolated test `docker compose config` all exit 0.
- Staging and production render no `openobserve` service; the test stack renders no OpenObserve variable or service.
- A blank `OPENOBSERVE_ORGANIZATION_ID` fails interpolation with `required variable OPENOBSERVE_ORGANIZATION_ID is missing a value`, as intended.

## Root checks

- `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, `npm run test:markdown-format` (30 passed), and `npm run fmt:check` all pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Set up OpenObserve v1.0.1 (digest-pinned) as running infrastructure in the Compose topology: a shared `timeful-observability` project for staging and production on the external `timeful-edge` network with a persistent volume, loopback-only 5080, explicit memory and disk caps, 14-day retention, and disabled telemetry; plus a dedicated development instance. Documented the environment contract, per-environment provisioning, and SSH-tunnel access across the app env examples, docs/environments.md, and DEPLOYMENT.md, and revised ADR-022 to the v1.0.1 pin. Verification on the running development instance exercised QR-016 (loopback-only refusal from a public interface, unauthenticated and invalid-token rejection, valid service-account ingest) and QR-019 (two organizations, rejected cross-organization reads and ingests, successful own-organization reads and ingests). Every documented Compose stack configures, the isolated test stack is unaffected, and the root Markdown and formatting checks pass. Verification also corrected the contract: OpenObserve generates organization identifiers and environment credentials authenticate only against them, so the server variable is `OPENOBSERVE_ORGANIZATION_ID` and provisioning creates a service account whose token is shown only once.
<!-- SECTION:FINAL_SUMMARY:END -->
