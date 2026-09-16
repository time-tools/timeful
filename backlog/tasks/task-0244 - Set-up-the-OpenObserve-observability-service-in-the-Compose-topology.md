---
id: TASK-0244
title: Set up the OpenObserve observability service in the Compose topology
status: To Do
assignee: []
created_date: '2026-09-16 12:35'
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

Constraints from ADR-022 and QR-015 through QR-019: open-source edition, single-node mode, local-disk storage, image pinned by digest to v0.92.2; loopback-only listener with no Caddy route and SSH-tunnel UI access; explicit memory and disk caps for the memory cache, query engine memory pool, memtable, and disk cache instead of the memory-consuming defaults; anonymous telemetry disabled; retention of 14 days for every stream; authenticated ingest with per-environment organizations, streams, and environment-scoped credentials.

The environment contract must be updated together: .env.development.example, .env.staging.example, .env.production.example, docs/environments.md, and DEPLOYMENT.md. Compose has no application-value fallbacks, so every interpolated variable must be declared in the selected env file. The isolated test stack must remain unaffected.

Governing requirements: QR-015, QR-016, QR-017 (verification lands with the log-shipping task), QR-018, QR-019. Decision record: ADR-022.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 OpenObserve runs in the documented Compose topology with the image pinned by digest to v0.92.2, single-node local-disk storage, and a loopback-only listener, and no Caddy route or public port exposes it
- [ ] #2 Staging and production share one OpenObserve instance that runs as its own Compose project on the deployment host with a persistent data volume, and neither app stack runs a second observability instance
- [ ] #3 Development runs its own OpenObserve instance in the local development stack with its own data volume, and development data is never sent to the shared instance
- [ ] #4 Explicit limits are configured for the memory cache, the query engine memory pool, the memtable, and the disk cache, and anonymous telemetry is disabled (QR-015)
- [ ] #5 Retention is configured to 14 days for every stream and is not left unlimited (QR-018)
- [ ] #6 A connection attempt from a public interface to the OpenObserve port fails, an unauthenticated or incorrectly authenticated ingest request is rejected, and an ingest request with valid environment credentials succeeds (QR-016)
- [ ] #7 Read and ingest attempts that use staging credentials against the production organization, or production credentials against the staging organization, are rejected, while each credential succeeds against its own environment (QR-019)
- [ ] #8 .env.development.example, .env.staging.example, .env.production.example, docs/environments.md, and DEPLOYMENT.md document the OpenObserve variables, per-environment organization and credential provisioning, resource caps, retention, and SSH-tunnel UI access
- [ ] #9 The isolated test stack is unaffected: compose.test.yaml, server-test, and server-route-test need no OpenObserve service or variables, and the documented development, staging, and production Compose commands still validate
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
