---
id: TASK-0177
title: Retire or replace unmaintained frontend dependencies
status: To Do
assignee: []
created_date: '2026-09-08 12:11'
updated_date: '2026-09-13 15:06'
labels:
  - frontend
dependencies:
  - TASK-0171
  - TASK-0173
references:
  - frontend/package.json
  - frontend/src/components/home/Dashboard.vue
  - frontend/src/components/landing/GithubStarButton.vue
priority: medium
type: chore
ordinal: 183000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A 2026-09-08 dependency audit found unmaintained or awkward frontend packages remaining after the Tailwind v4 and Vuetify v4 migrations. For each, make and record a keep/replace/drop decision, then act on it. The goal is a smaller, maintained dependency set with no appearance or behavior change.

Audit findings:
- is-ua-webview ^1.1.2: tiny, unmaintained user-agent-webview detection helper.
- github-buttons ^2.33.0: 2.x line, used by frontend/src/components/landing/GithubStarButton.vue.
- vuedraggable ^4.1.0 (vue.draggable.next, stale since ~2022): powers Dashboard drag-and-drop in frontend/src/components/home/Dashboard.vue.

Resolved: bson-objectid ^2.0.4 was dropped by TASK-0177.01 (both call sites removed; no import or lockfile entry remains), so it is no longer part of this audit.

Replace with maintained alternatives, drop where trivially inlinable, or keep with recorded rationale. Each decision must state what the replacement is and why it is maintained (or why keeping is justified).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every remaining listed package (is-ua-webview, github-buttons, vuedraggable) has a recorded drop/replace/keep decision with rationale
- [ ] #2 Dropped or replaced packages have zero remaining imports and lockfile entries
- [ ] #3 Affected components keep unit-test coverage and the required frontend checks pass: lint, fmt:check, typecheck, build, test:unit
- [ ] #4 No visible behavior or appearance change in Dashboard drag-drop, GithubStarButton, or other affected call sites
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-12 19:53
---
bson-objectid decision recorded: drop. TASK-0177.01 removed both call sites (Dashboard ObjectID-timestamp sorting was dead; signup blocks now use non-canonical local keys) and deleted `bson-objectid` from `frontend/package.json` and `frontend/package-lock.json`, so no import or lockfile entry remains. Rationale and PostgreSQL-backed reachability evidence are in TASK-0177.01 comment #2 and its final summary. The other three audit findings (is-ua-webview, github-buttons, vuedraggable) are still open under this task.
---

author: opencode
created: 2026-09-13 15:06
---
2026-09-13: refreshed after TASK-0177.01 completed the bson-objectid decision (dropped; no imports or lockfile entries remain). Moved it out of the open audit findings and removed it from AC #1; the frontend also no longer sits in a Mongo era.
---
<!-- COMMENTS:END -->
