---
id: TASK-0240
title: Reduce frontend production bundle size
status: To Do
assignee: []
created_date: '2026-09-15 19:42'
updated_date: '2026-09-15 20:07'
labels:
  - frontend
  - bundle-size
dependencies: []
priority: medium
type: chore
ordinal: 236000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Outcome

Cut the production frontend's initial payload and emitted asset weight through independently deliverable improvements, each with a recorded before/after measurement.

## Context

- The last recorded production build is 3.0 MB in `frontend/dist`, with an eager preload graph of roughly 1.1 MB raw / 319 KB gzip, measured from the built `index.html` modulepreload set.
- The largest eager contributors are the entry chunk (237 KB raw / 75 KB gzip), the Vue runtime (126 KB / 48 KB), Vuetify core/utility chunks, the Temporal chunk (60 KB / 21 KB), the full Vuetify component CSS plus app stylesheet (269 KB raw / 42 KB), and bundled fonts (23 files, ~270 KB).
- The landing demo image `frontend/src/assets/demo/event.webp` alone is 417,756 bytes.
- Each subtask carries its own scope, acceptance criteria, and measurement requirement; this parent records the consolidated outcome.

## Out of scope

- posthog-js changes; the lighter entrypoint was evaluated and skipped for now.
- Backend, API, or build-deployment changes.

## Constraints

- No user-visible behavior or appearance change is acceptable unless a subtask explicitly records an approved tradeoff.
- Keep the improvements independently deliverable and reviewable.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All included subtasks are completed with their required evidence recorded
- [ ] #2 Consolidated eager-payload and emitted-asset before/after numbers are recorded in the parent final summary
- [ ] #3 No subtask leaves the required frontend checks failing
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
