---
id: TASK-0238
title: Add the OpenObserve observability ADR and quality requirements
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-15 21:04'
updated_date: '2026-09-15 21:08'
labels: []
dependencies: []
references:
  - 'https://github.com/openobserve/openobserve'
  - 'https://openobserve.ai/docs/architecture/'
  - >-
    https://openobserve.ai/docs/administration/configuration/environment-variables/
documentation:
  - docs/design/architecture/adr/TEMPLATE.md
  - docs/design/architecture/README.md
  - docs/requirements/quality/README.md
  - docs/requirements/quality/qr/TEMPLATE.md
  - docs/requirements/README.md
  - docs/AGENTS.md
priority: medium
type: docs
ordinal: 234000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add the architecture decision record and the quality requirements for self-hosted observability.

Decision to record: Timeful adopts OpenObserve (open-source edition, single-node, local-disk storage) as its observability platform for logs, metrics, and traces. Server logs ship first. Staging and production share one instance on the deployment host, isolated by organization/stream and credentials; development runs its own loopback instance in the default Compose stack so local agents and developers can inspect logs and traces while debugging. The Go server emits OpenTelemetry OTLP records directly to OpenObserve, with no sidecar log shipper. The UI and ingest endpoint stay on loopback, and operators reach the UI through an SSH tunnel. Observability data is retained 14 days. The image is pinned initially to v0.92.2 pending a 1.x review. Resource use is explicitly capped because OpenObserve defaults consume all available RAM, and anonymous telemetry is disabled. OpenObserve's AGPL-3.0 license matches the repository license, and PostHog remains the product analytics tool.

Existing requirements: the decision enables QR-010 (diagnose failed requests) and constrains QR-004 (exclude secrets from diagnostics). Four new quality requirements follow from it: bounded observability resource use, private endpoints with authenticated ingest, serving unaffected by ingestion failure, and deletion after the retention window.

Deliverables: docs/design/architecture/adr/ADR-022.md, docs/requirements/quality/qr/QR-015.md through QR-018.md, the ADR row in docs/design/README.md, the QR rows in docs/requirements/README.md, and the ADR addresses metadata pointing at QR-010 and the four new QRs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ADR-022 exists with the next free ADR id, follows docs/design/architecture/adr/TEMPLATE.md, carries status accepted and components backend and infrastructure, and records context, considered options (OpenObserve, SigNoz, VictoriaMetrics/VictoriaLogs with Grafana, and retaining file logs only), decision outcome, and consequences for the agreed decision
- [x] #2 ADR-022 fixes single-node local-disk deployment, one shared staging and production instance with per-environment organization/stream separation, a separate development instance in the default Compose stack, OTLP from the Go server, loopback-only endpoints, 14-day retention, the v0.92.2 initial pin, explicit resource caps, disabled anonymous telemetry, and the PostHog product-analytics boundary
- [x] #3 ADR-022 frontmatter declares addresses.enables QR-010 and addresses.constrains QR-004, QR-015, QR-016, QR-017, and QR-018, and declares no untracked quality attribute for resource utilization
- [x] #4 QR-015 through QR-018 exist, each declaring exactly one valid ISO/IEC 25010 classification pair and all six quality-scenario subsections, and each stays self-contained without citing the ADR
- [x] #5 docs/design/README.md lists ADR-022 in the active records table, and docs/requirements/README.md lists QR-015 through QR-018 in the Quality Requirements table with stable relative links
- [x] #6 npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format, and npm run fmt:check pass from the repository root
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
1. Author `docs/design/architecture/adr/ADR-022.md` from TEMPLATE.md. Frontmatter: id ADR-022, title "Use OpenObserve as the self-hosted observability platform", components backend and infrastructure, status accepted, dates 2026-09-15, addresses.enables QR-010, addresses.constrains QR-004 plus QR-015 through QR-018. Sections: Context (4 GB single deployment host, staging and production behind one Caddy edge, file-only logging via server/logger and server_logs with no metrics or traces, QR-010 diagnosis need, local agent debugging need, PostHog boundary); Considered Options (retain file logs, OpenObserve, SigNoz with ClickHouse, VictoriaMetrics/VictoriaLogs with Grafana, managed cloud service); Decision Outcome (single-node open-source local-disk mode, v0.92.2 initial digest pin, one shared staging/production instance with per-environment organizations, separate development instance in the default stack, OTLP direct from the Go server, loopback-only endpoints, 14-day retention, explicit resource caps, disabled anonymous telemetry, AGPL-3.0 compatibility); Consequences.
2. Author QR-015 (performance efficiency / resource utilization), QR-016 (security / resistance), QR-017 (reliability / fault tolerance, backend and infrastructure), and QR-018 (security / confidentiality) from quality/qr/TEMPLATE.md with all six scenario subsections, self-contained and without ADR citations.
3. Add the ADR-022 row to the active records table in `docs/design/README.md` and QR-015 through QR-018 rows to the Quality Requirements table in `docs/requirements/README.md`.
4. Run root Markdown formatting and checks: `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, `npm run test:markdown-format`, `npm run fmt:check`; then `graphify update .`.
5. Finalize the task with acceptance-criteria evidence and DoD state; unit and e2e tests are exempt because this is a documentation-only change.

Risks: preserve one-sentence-per-line authoring; keep ADR frontmatter consistent with docs/design/architecture/README.md; keep QR records independent of the ADR text.

Step 1 executed as written except the ADR dates: created_date and updated_date use the local calendar date 2026-09-16 (the Backlog record was created 2026-09-15 UTC).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Authored ADR-022 and QR-015 through QR-018 from the repository templates. Verified QR classifications against the ISO/IEC 25010 table, verified all six scenario subsections in each QR, and confirmed no QR cites the ADR. Because scripts/markdown.mjs resolves files through git ls-files and the new files were untracked, checked them directly with prettier (already clean) before running the repository Markdown checks; all passed. Ran graphify update . to refresh the knowledge graph.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Recorded the OpenObserve observability decision as ADR-022 and added four observability quality requirements.

- Added `docs/design/architecture/adr/ADR-022.md`: accepted decision adopting the OpenObserve open-source edition in single-node local-disk mode for logs, metrics, and traces. It fixes one shared staging and production instance with per-environment organizations and credentials, a separate loopback development instance in the default Compose stack, direct OpenTelemetry OTLP from the Go server with no sidecar shipper, loopback-only endpoints with SSH-tunnel UI access, 14-day retention, the digest-pinned v0.92.2 initial release with a separate review before 1.x, explicit memory and disk caps because the defaults consume all available RAM, disabled anonymous telemetry, AGPL-3.0 license alignment, and the PostHog product-analytics boundary. Frontmatter enables QR-010 and constrains QR-004 plus the four new QRs.
- Added `docs/requirements/quality/qr/QR-015.md` through `QR-018.md`: bounded observability resource use (performance efficiency / resource utilization), private endpoints with authenticated ingest (security / resistance), serving unaffected by observability ingestion failure (reliability / fault tolerance), and deletion after the retention window (security / confidentiality). Each record is self-contained with all six quality-scenario subsections and cites no ADR.
- Added the ADR-022 row to the active records table in `docs/design/README.md` and the QR-015 through QR-018 rows to the Quality Requirements table in `docs/requirements/README.md`.

Evidence: `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, `npm run test:markdown-format` (30 passed), and `npm run fmt:check` all pass from the repository root; `graphify update .` rebuilt the knowledge graph. Unit and e2e tests are exempt for this documentation-only change.

Follow-ups (not created): implementation tasks for the Compose service, environment variables, and deployment docs; the structured-logging rework that fulfills QR-010; and metrics and traces instrumentation.
<!-- SECTION:FINAL_SUMMARY:END -->
