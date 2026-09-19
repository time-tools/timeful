---
id: TASK-0275
title: End the target transfer page when a waiting link is cancelled or expired
status: To Do
assignee: []
created_date: '2026-09-19 18:14'
labels:
  - frontend
dependencies: []
references:
  - frontend/src/views/AccessTransfer.vue
  - frontend/src/views/AccessTransfer.test.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
documentation:
  - docs/design/architecture/adr/ADR-010.md
priority: medium
type: enhancement
ordinal: 275000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0274 added a two-second poll of the `open` action on the target `Continue on this device` page so it highlights step 3 as soon as the source approves. That poll deliberately treats every failure as transient. When the source cancels the transfer or the five-minute window expires while the target is still waiting, `open` starts returning 403/404, the poll keeps ignoring it, and the page keeps showing "Waiting for approval in the other browser" with an active Continue after approval control. The target user only discovers the transfer is unusable when they click Continue and an error replaces the waiting state. The waiting page should instead end the wait with a clear unavailable message and stop offering Continue, while genuine transient failures still keep waiting.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 When a waiting target page receives a definitive 403 or 404 from a status poll, it stops polling and replaces the waiting message with a clear unavailable message that names the link as expired, cancelled, or unavailable and points to asking the source browser for a new link.
- [ ] #2 Once the transfer is known unavailable, the Continue after approval control is no longer offered and step 3 no longer claims approval can still arrive.
- [ ] #3 A transient poll failure such as a 500 or network error keeps the page waiting silently with no error alert, and polling continues.
- [ ] #4 The existing initial-open failure behavior and the pending-to-approved highlight behavior remain unchanged.
- [ ] #5 Unit tests cover cancellation and expiry detection during polling, the transient-versus-definitive distinction, polling stopping after detection, and removal of the Continue control.
- [ ] #6 Firefox access-transfer e2e coverage for cancelled and expired links exercises the target waiting page ending without a Continue click.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `graphify update .`
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
