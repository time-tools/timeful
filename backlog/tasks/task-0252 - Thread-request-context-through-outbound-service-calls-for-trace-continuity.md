---
id: TASK-0252
title: Thread request context through outbound service calls for trace continuity
status: To Do
assignee: []
created_date: '2026-09-16 21:08'
labels: []
dependencies: []
references:
  - server/observability/transport.go
  - server/services/services.go
  - server/services/listmonk/listmonk.go
  - server/services/calendar/google_calendar.go
  - server/services/auth/auth.go
  - server/slackbot/commands/utils.go
documentation:
  - docs/environments.md
  - docs/design/architecture/adr/ADR-022.md
  - docs/requirements/quality/qr/QR-004.md
  - docs/requirements/quality/qr/QR-017.md
priority: medium
type: enhancement
ordinal: 253000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Outbound HTTP client spans recorded by `observability.Transport` (added in TASK-0246) currently start their own root traces, because the service functions that make outbound calls build requests without a request context (`http.NewRequest` defaults to `context.Background()`). Request, PostgreSQL, and log records already join one trace; this task closes the remaining continuity gap for outbound calls so an operator can follow a failed request into the Listmonk, OAuth or calendar, or Slack call it made.

Expected outcome: request-scoped outbound call sites accept a `context.Context`, switch to `http.NewRequestWithContext`, and pass the request context (or a documented equivalent) so their client spans nest under the request span.

Key decision to settle during execution: cancellation semantics. The recommended default is `context.WithoutCancel(requestContext)` (Go 1.21+) plus an explicit per-call timeout, which links the span to the request trace without letting a client disconnect abort an outbound call that must complete. Using the raw request context is appropriate only where abort-on-disconnect is desired; the choice and its rationale must be recorded in the implementation notes and reflected in docs/environments.md.

Scope notes: the transport itself needs no change; the work is context plumbing through `server/services/services.go`, `server/services/listmonk`, `server/services/calendar`, `server/services/auth`, and `server/slackbot/commands`, plus their route callers. The dormant Cloud Tasks functions in `server/services/gcloud` are included only if that path still exists when this task runs; TASK-0251 may remove it. QR-004 attribute rules are unchanged, so no URL, query, header, credential, or token may be recorded.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Outbound HTTP calls made while handling a request produce client spans in the same trace as the request span, proven by a unit test that drives a service call with a request span in context and asserts matching trace identifiers.
- [ ] #2 Cancellation semantics are explicit and documented: outbound calls detach from client cancellation with `context.WithoutCancel` plus an explicit timeout, or use the raw request context only where aborting the call on client disconnect is the intended behavior; existing success paths are unchanged.
- [ ] #3 Request-scoped outbound service functions accept `context.Context` as their first parameter and their callers pass the request context (or a documented detached equivalent); no request-scoped outbound call builds its request with `context.Background()`.
- [ ] #4 QR-004 still holds: outbound spans and every other exported signal continue to exclude URL paths and queries, headers, credentials, and tokens.
- [ ] #5 Unit tests cover the context plumbing and the chosen cancellation behavior, and docs/environments.md records the outbound trace-continuity behavior and cancellation semantics.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
