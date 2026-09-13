---
id: TASK-0113
title: >-
  Add e2e regression spec asserting no upgrade, paywall, or ad surfaces remain
  in the app
status: To Do
assignee: []
created_date: '2026-08-29 22:05'
updated_date: '2026-09-13 15:06'
labels:
  - e2e
  - cleanup
  - frontend
dependencies: []
references:
  - e2e/specs/landing-hero.spec.ts
  - frontend/src/router/index.ts
  - >-
    backlog/tasks/task-0112 -
    Address-TASK-0104-follow-up-observations-dead-Teams-upgrade-dialog-unused-slackbot-monetization-plumbing-e2e-no-upgrade-check-graphify-purge.md
type: chore
ordinal: 119000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Postponed from TASK-0112 (owner decision 2026-08-30): add an e2e regression spec asserting the app exposes no upgrade, paywall, or advertising surfaces, per the TASK-0104 removal.

Scope:
- Public pages asserted directly: landing (/) and sign-in (/sign-in) must render no upgrade, paywall, or ad text or components.
- Signed-in shell covered via Playwright API route interception (no real auth): intercept /api/auth/status and the user profile/events calls, then assert /home, /settings, and the user menu expose no upgrade, paywall, or ad surfaces.
- Keep monitoring analytics out of scope; this spec only asserts absence of monetization surfaces.

Constraints:
- Do not reintroduce any freemium, billing, or advertising plumbing.
- Keep repo-tracked Playwright specs under e2e/specs/ and run them via npm run test:e2e from e2e/ (isolated test stack; never the development API).
- When starting execution, read BACKLOG_WORKFLOW.md, mark the task In Progress, assign it, and keep plan/notes current.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A Playwright spec under e2e/specs/ asserts that public pages (landing and sign-in) expose no upgrade, paywall, or ad text or components.
- [ ] #2 The spec renders the signed-in shell by intercepting /api/auth/status and related user API calls via page.route and asserts that home, settings, and the user menu expose no upgrade, paywall, or ad surfaces.
- [ ] #3 The new spec passes via npm run test:e2e on the default desktop project from e2e/.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Research findings carried over from the 2026-08-30 TASK-0112 scoping session (verified against the worktree; refreshed 2026-09-13 after the MongoDB retirement and the root e2e package move): there is no authenticated e2e infrastructure today. Auth is OTP-based (server/routes/auth.go check-email/send/verify); OTP challenges live in PostgreSQL (server/postgres/otp.go) and are delivered through Listmonk emails with no test hook, so real signed-in flows cannot be scripted cheaply. The firefox-desktop Playwright project matches only timed-event-*firefox.spec.ts, so a new generic spec runs on the default desktop (chromium) project. Feasible approach: assert public pages (landing, sign-in) directly, and render the signed-in shell by intercepting /api/auth/status and user API calls via page.route (route stubbing precedent exists in e2e/specs/landing-hero.spec.ts). Router guard redirects /home and /settings to landing when /auth/status fails (frontend/src/router/index.ts:113), so the interception must cover it. The frontend get/post helpers hit same-origin /api paths, so interception patterns should use the /api/ prefix.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-13 15:06
---
2026-09-13: refreshed for the MongoDB retirement (TASK-0199) and the root e2e package move (TASK-0164). OTP storage is now PostgreSQL (server/postgres/otp.go) and the spec paths/commands now point at e2e/.
---
<!-- COMMENTS:END -->
