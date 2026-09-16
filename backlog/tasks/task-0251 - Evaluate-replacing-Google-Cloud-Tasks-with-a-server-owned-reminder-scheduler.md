---
id: TASK-0251
title: Evaluate replacing Google Cloud Tasks with a server-owned reminder scheduler
status: To Do
assignee: []
created_date: '2026-09-16 20:42'
labels: []
dependencies: []
references:
  - server/services/gcloud/tasks.go
  - server/models/event.go
  - server/main.go
priority: medium
type: spike
ordinal: 252000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Evaluate replacing Google Cloud Tasks with a reminder scheduler owned by the server, and record a decision plus, if adopted, the follow-up work needed.

Context: the Cloud Tasks reminder path is currently dormant in this repository. `CreateEmailTask` and `DeleteEmailTask` in server/services/gcloud/tasks.go have no production callers, `Event.TaskIds` in server/models/event.go is unused, and no requirement under docs/requirements/ mentions reminders. `gcloud.InitTasks` still initializes a `cloudtasks` client from `SERVICE_ACCOUNT_KEY_PATH` at server startup when that variable is set, and the scheduled tasks POST to Listmonk's transactional endpoint to send reminder emails at +0, +24 hours, and +3 days.

Preliminary pros of a server-owned scheduler:
- Removes the external Google Cloud dependency, its service-account credentials, quotas, and egress, and shrinks the environment contract.
- Aligns with the self-hosted direction of ADR-022 and keeps scheduling logic, retries, and observability in one codebase and one trace.
- Server-side traces and metrics can cover the reminder path directly.

Preliminary cons:
- The server then owns at-least-once semantics: crash recovery, duplicate-email prevention, retry and backoff, race-free cancellation, and safe claiming under multiple instances.
- Timezone, daylight-saving, and "now versus due" semantics must be defined explicitly at a domain boundary.
- More tests and operational surface than the managed scheduler, which provides dispatch and retry for free.

Scope: this is an evaluation spike. It does not implement the scheduler. If adoption is recommended, the outcome is a design (ADR + QRs) plus the implementation or requirement work needed; a requirement record may need to be authored first per docs/requirements/AGENTS.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The evaluation decides whether the server should own reminder scheduling instead of Google Cloud Tasks, with the decision rationale recorded against the current system state (call sites, `Event.TaskIds` usage, and whether any requirement covers reminders).
- [ ] #2 If adoption is recommended, the evaluation records a design covering durable scheduling state, at-least-once delivery and duplicate prevention, retry and backoff, cancellation when an event changes or is deleted, safe claiming under multiple server instances, and explicit timezone and due-time semantics for the +24 hour and +3 day reminders.
- [ ] #3 If adoption is recommended, the evaluation records the environment-contract reduction (for example `SERVICE_ACCOUNT_KEY_PATH` and `GOOGLE_CLOUD_TASKS_*`) and the follow-up implementation task(s) or requirement work needed before implementation starts.
- [ ] #4 If Cloud Tasks is retained, the evaluation records why the current managed path remains preferable.
- [ ] #5 The preliminary pros and cons recorded in the description are validated or corrected against the current codebase, and the findings are recorded in the task notes.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
