---
id: TASK-0235
title: >-
  Fix frontend typecheck failure on the vitest 5 upgrade (TS2883 in
  scheduleOverlapTestMocks)
status: To Do
assignee: []
created_date: '2026-09-15 13:10'
labels:
  - frontend
  - ci
dependencies: []
references:
  - frontend/src/components/schedule_overlap/scheduleOverlapTestMocks.ts
  - .github/workflows/frontend-ci.yml
  - frontend/package.json
priority: medium
type: bug
ordinal: 232000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Frontend CI fails at `npm run typecheck` on Dependabot branch `dependabot/npm_and_yarn/frontend/vitest-5.0.0` (commit dba118ea), which bumps vitest from 4.1.11 to 5.0.0 in `frontend/`.

The failing job is `frontend-quality` in `.github/workflows/frontend-ci.yml`, which runs `npm ci` before `npm run typecheck`.

Reproduction (verified locally):
`cd frontend && npm ci && npm run typecheck` exits 2 with four TS2883 errors in `src/components/schedule_overlap/scheduleOverlapTestMocks.ts` at lines 6-9. The inferred types of `putMock`, `refreshAuthUserMock`, `showInfoMock`, and `showErrorMock` cannot be named without a reference to an internal `Procedure` type from `vitest/dist/chunks/config.d.*`. A local `node_modules` that still has vitest 4 installed masks the failure; `npm ci` is required to reproduce.

Desired outcome: the vitest 5 dependency set typechecks, the mocked helpers and their unit-test consumers stay type-safe, and frontend CI is green on the upgrade branch.

Out of scope: unrelated dependency upgrades, refactors of the schedule overlap tests beyond what the vitest 5 types require, and changes to backend or e2e packages.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `cd frontend && npm ci && npm run typecheck` exits 0 on the vitest 5 dependency set.
- [ ] #2 `src/components/schedule_overlap/scheduleOverlapTestMocks.ts` no longer emits TS2883 for its exported mock handles.
- [ ] #3 The fix does not silence errors with `@ts-ignore`, `@ts-expect-error`, or `any` casts.
- [ ] #4 Existing schedule overlap unit tests still pass under vitest 5 via `npm run test:unit`.
- [ ] #5 The full required frontend check set passes: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`.
- [ ] #6 The Frontend CI `frontend-quality` job passes on the vitest 5 upgrade branch.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
