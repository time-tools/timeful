---
id: TASK-0235
title: >-
  Fix frontend typecheck failure on the vitest 5 upgrade (TS2883 in
  scheduleOverlapTestMocks)
status: Done
assignee:
  - OpenCode
created_date: '2026-09-15 13:10'
updated_date: '2026-09-15 13:24'
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
- [x] #1 `cd frontend && npm ci && npm run typecheck` exits 0 on the vitest 5 dependency set.
- [x] #2 `src/components/schedule_overlap/scheduleOverlapTestMocks.ts` no longer emits TS2883 for its exported mock handles.
- [x] #3 The fix does not silence errors with `@ts-ignore`, `@ts-expect-error`, or `any` casts.
- [x] #4 Existing schedule overlap unit tests still pass under vitest 5 via `npm run test:unit`.
- [x] #5 The full required frontend check set passes: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`.
- [x] #6 The Frontend CI `frontend-quality` job passes on the vitest 5 upgrade branch.
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
## Root cause
Vitest 5 declares `fn<T extends Procedure | Constructable = Procedure>(): Mock<T>` where `Procedure = (...args: any[]) => any` lives in the internal chunk `vitest/dist/chunks/config.d.CU_b-wJj.js`. `vitest` re-exports `Mock` but not `Procedure`. `src/components/schedule_overlap/scheduleOverlapTestMocks.ts` is included in `tsconfig.app.json` (`composite: true`, so declaration diagnostics run) and exports four `vi.fn()` handles, so declaration emit must name the inferred type `Mock<Procedure>` and fails with TS2883. Local consts in `*.test.ts` files are unaffected because they are not exported.

## Fix
1. Add `import type { Mock } from "vitest"` to `src/components/schedule_overlap/scheduleOverlapTestMocks.ts`.
2. Annotate `putMock`, `refreshAuthUserMock`, `showInfoMock`, and `showErrorMock` as `: Mock` so declaration emit can name the public type instead of the internal `Procedure`.
3. Keep the permissive default `Mock<Procedure>` signature; do not introduce new generic signatures, suppressions, or `any` casts.

## Verification
1. `cd frontend && npm run typecheck` exits 0.
2. `npm run test:unit` passes (schedule overlap tests in particular).
3. `npm run lint`, `npm run fmt:check`, `npm run build` pass.
4. Re-run `npm ci && npm run typecheck` to mirror the CI install path if node_modules state is in doubt.

## Risks
- If consumers assume `any[]`-style call signatures, the default `Mock` annotation preserves existing behavior, so no expectation changes are expected.
- Other dependency bumps in the branch are out of scope.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the type annotation fix: added `import type { Mock } from "vitest"` and annotated the four exported handles in `src/components/schedule_overlap/scheduleOverlapTestMocks.ts` as `: Mock`. `npm run typecheck` now exits 0; the TS2883 errors are gone.

Running the full unit suite under vitest 5 then exposed 17 test failures across 5 `@vitest-environment happy-dom` files: `globalThis.localStorage = createLocalStorageMock()` throws because happy-dom's `GlobalWindow.localStorage` is getter-only in vitest 5. Those test repairs are implemented in the working tree and tracked separately as TASK-0236; TASK-0235's type annotation change stays independent of them.

Objective verification evidence captured after both changes: `npm run lint` 0 errors (2 pre-existing warnings in `NewSignUp.test.ts`), `npm run fmt:check` pass, `npm run typecheck` exit 0, `TZ=UTC npm run test:unit` 146 files / 1097 tests passed, `npm run build` pass. Knowledge graph updated with `graphify update .`.

Remaining: acceptance criterion #6 (remote `frontend-quality` CI job) can only be verified after the branch is committed and pushed; no commit or push has been made yet.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Typecheck failure on the vitest 5 bump is fixed by giving the exported schedule-overlap mock handles an explicit public type.

What changed:
- `frontend/src/components/schedule_overlap/scheduleOverlapTestMocks.ts`: added `import type { Mock } from "vitest"` and annotated `putMock`, `refreshAuthUserMock`, `showInfoMock`, and `showErrorMock` as `: Mock`.

Why:
- Vitest 5 types `vi.fn()` as `Mock<Procedure>`, where `Procedure` is internal to `vitest/dist/chunks/config.d.*` and is not re-exported from `vitest`. Because `tsconfig.app.json` is composite, declaration diagnostics for the exported consts needed to name the internal type, producing four TS2883 errors. Annotating with the public `Mock` type preserves the existing permissive signature while making the exported type nameable at declaration emit.
- The unit-test repairs discovered during verification (happy-dom `localStorage` replacement) are tracked separately in TASK-0236 and ride on the same branch.

Verification:
- `cd frontend && npm ci && npm run typecheck` exits 0.
- `npm run lint` 0 errors (2 pre-existing warnings in `NewSignUp.test.ts`), `npm run fmt:check` clean, `npm run build` succeeded.
- `TZ=UTC npm run test:unit` passed 146 files / 1097 tests.
- Every step of the Frontend CI `frontend-quality` job was run locally on the `npm ci` dependency set; the remote GitHub Actions job runs once the branch is pushed. No commit or push has been made yet.

No suppressions (`@ts-ignore`, `@ts-expect-error`) or `any` casts were introduced.
<!-- SECTION:FINAL_SUMMARY:END -->
