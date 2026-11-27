---
id: TASK-0373
title: Align the scheduled time row with desktop event header actions
status: Done
assignee:
  - Codex
created_date: '2026-11-27 06:20'
updated_date: '2026-11-27 06:31'
labels:
  - frontend
dependencies: []
documentation:
  - docs/requirements/functional/fr/FR-139.md
modified_files:
  - frontend/src/views/Event.vue
  - e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts
  - docs/requirements/functional/fr/FR-139.md
type: bug
ordinal: 312000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On a scheduled Timed Event Page on desktop, the saved Event Occurrence Span appears as a short line beneath the title. This leaves Edit event, Copy link, and Manage access partway between the options row and Reschedule event, and shifts the description below the metadata actions by the same amount.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a desktop Timed Event Page, a saved Event Occurrence Span that fits on one line is vertically centered in its own row below the title, at the same height as a standard desktop header action row.
- [x] #2 When the title and span each fit on one line and the header action groups each fit in a standard row, the span aligns with the options row and the Edit event, Copy link, and Manage access row aligns with Reschedule event.
- [x] #3 With no saved span, the desktop header keeps the unscheduled layout without reserving a span row. The phone layout retains its existing order and spacing.
- [x] #4 A browser regression check verifies the scheduled desktop row height and alignment. Record the failure before the fix and the pass after the fix in the task notes.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add a focused Firefox layout regression beside the saved-span browser coverage. Measure the unscheduled desktop alignment, saved desktop row height and alignment, and compact phone layout; run it unchanged against the current UI and capture the geometry failure.
2. Give timed scheduled spans the existing desktop details row height and spacing in Event.vue while retaining the independent controls column and compact phone and dates-only layouts.
3. Rerun the same Firefox regression, then the surrounding scheduled-header coverage and required frontend checks. Refresh the codebase index and record the verification evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Fail before: the isolated Firefox test `desktop scheduled time occupies a full header action row` passed its unscheduled and phone steps, then failed the desktop height assertion at 20px versus the 40px More options row. Playwright run `2026-11-27T06-25-54.054037109Z-p96514`.

Pass after: reran the identical isolated Firefox regression. The scheduled row matched the 40px More options row, the date/time was vertically centered, all three metadata buttons centered alongside Reschedule event, and the unscheduled desktop and compact phone assertions passed.

Frontend lint, format, typecheck, production build, and all 1,298 unit tests passed. Frontend lint reports the existing two one-component-per-file warnings in NewSignUp.test.ts. E2E typecheck and targeted spec lint passed. The codebase index refresh completed.

The adjacent Firefox scheduling spec and Chromium desktop event-header spec passed 8 tests. The Chromium desktop project skipped the mobile-only availability-group test by design. Markdown format check and git diff --check passed. Swagger and root scripts/Prettier checks were inapplicable. FR-139 records the changed desktop behavior.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The saved time on a desktop Timed Event Page now uses the standard 40px header row and aligns with More options. Edit event, Copy link, and Manage access align with Reschedule event. The saved time keeps its compact styling on phones, and the unscheduled desktop layout has no reserved row.

The Firefox regression reproduced the 20px versus 40px mismatch before the change and passed unchanged afterward. Related Firefox and Chromium desktop checks passed, as did frontend lint, formatting, typecheck, build, and all 1,298 unit tests.
<!-- SECTION:FINAL_SUMMARY:END -->
