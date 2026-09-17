---
id: TASK-0236
title: >-
  Fix vitest 5 happy-dom localStorage replacement failures in frontend unit
  tests
status: Done
assignee:
  - OpenCode
created_date: '2026-09-15 13:22'
updated_date: '2026-09-15 13:24'
labels:
  - frontend
  - tests
dependencies: []
references:
  - frontend/src/composables/event/useEventLoader.test.ts
  - frontend/src/components/groups/InvitationDialog.test.ts
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - frontend/src/components/schedule_overlap/RespondentsList.test.ts
  - >-
    frontend/src/components/schedule_overlap/useScheduleOverlapPreferences.test.ts
priority: medium
type: bug
ordinal: 233000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Vitest 5 resolves `@vitest-environment happy-dom` test files to happy-dom's `GlobalWindow`, where `localStorage` is a getter-only property. Tests that replace storage with `globalThis.localStorage = createLocalStorageMock()` now throw `TypeError: Cannot set property localStorage of #<GlobalWindow> which has only a getter`.

On the vitest 5 upgrade branch (`dependabot/npm_and_yarn/frontend/vitest-5.0.0`), 17 tests across 5 files fail this way once the typecheck blocker tracked in TASK-0235 is fixed:
- `src/composables/event/useEventLoader.test.ts`
- `src/components/groups/InvitationDialog.test.ts`
- `src/components/event/EventAccessTransfer.test.ts`
- `src/components/schedule_overlap/RespondentsList.test.ts`
- `src/components/schedule_overlap/useScheduleOverlapPreferences.test.ts`

Reproduction: `cd frontend && npm ci && TZ=UTC npm run test:unit`.

Desired outcome: these tests replace `localStorage` through vitest's supported global-stubbing API, matching the convention already used by most other frontend tests, with no production-code changes.

Current working-tree state: fixes are already implemented as part of the same session that produced TASK-0235; the full unit suite passes locally. This task exists so the test repairs are tracked, reviewed, and committed separately from the TASK-0235 type annotation change.

Out of scope: production code, e2e package, unrelated vitest 5 behavior changes, and the TS2883 type annotation work in TASK-0235.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The five affected test files no longer assign `globalThis.localStorage` directly and use vitest's supported global-stubbing API instead.
- [x] #2 The temporary localStorage replacement in `RespondentsList.test.ts` is restored after the test so later tests observe the environment default.
- [x] #3 `TZ=UTC npm run test:unit` passes all 146 files and 1097 tests under vitest 5.
- [x] #4 The full required frontend check set passes: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`.
- [x] #5 No production source files are modified; changes stay within `*.test.ts` files.
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
Vitest 5 resolves `@vitest-environment happy-dom` files with happy-dom's `GlobalWindow`. Its `populateGlobal` installs a setter that forwards `globalThis.localStorage = value` to `win.localStorage`, but `GlobalWindow.localStorage` is a getter-only property, so the assignment throws `TypeError: Cannot set property localStorage of #<GlobalWindow> which has only a getter`. Files without the pragma run in the node environment, where Node's own `localStorage` has a setter, so their direct assignments still work.

## Applied fix (in working tree)
1. `src/composables/event/useEventLoader.test.ts`, `src/components/groups/InvitationDialog.test.ts`, `src/components/event/EventAccessTransfer.test.ts`: replace the `beforeEach` assignment with `vi.stubGlobal("localStorage", createLocalStorageMock())`, matching the convention used by most other frontend tests.
2. `src/components/schedule_overlap/useScheduleOverlapPreferences.test.ts`: same replacement plus add `vi` to the vitest import.
3. `src/components/schedule_overlap/RespondentsList.test.ts`: replace the manual `previousStorage` capture/restore with `vi.stubGlobal("localStorage", createLocalStorageMock({...}))` and `vi.unstubAllGlobals()` in the existing `finally` block.

`vi.stubGlobal` uses `Object.defineProperty(globalThis, name, { value, writable: true, configurable: true })`, which bypasses the getter-only property; `vi.unstubAllGlobals()` restores the captured original descriptor.

## Verification
1. Targeted run of the five files: 5 files, 54 tests passed.
2. `TZ=UTC npm run test:unit`: 146 files, 1097 tests passed.
3. `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build` all pass.

## Risks
- `vi.unstubAllGlobals()` clears every global stub in `RespondentsList.test.ts`; the file currently stubs only `localStorage`, so no other stub is affected.
- Remaining remote CI verification depends on committing and pushing the branch.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reproduced under vitest 5 with `TZ=UTC npm run test:unit`: 5 files failed, 17 tests failed, all with `TypeError: Cannot set property localStorage of #<GlobalWindow> which has only a getter` at the direct `globalThis.localStorage = createLocalStorageMock()` assignments.

Applied the fix in the working tree: four files now call `vi.stubGlobal("localStorage", createLocalStorageMock())`; `useScheduleOverlapPreferences.test.ts` gained the `vi` import; `RespondentsList.test.ts` stubs with the seeded mock and restores with `vi.unstubAllGlobals()` in its `finally` block.

Evidence: targeted run of the five files passed 5 files / 54 tests; full `TZ=UTC npm run test:unit` passed 146 files / 1097 tests; `npm run lint` 0 errors (2 pre-existing warnings in `NewSignUp.test.ts`); `npm run fmt:check`, `npm run typecheck`, and `npm run build` all pass.

The changes are uncommitted; committing them separately from the TASK-0235 type annotation change is pending user direction, as is remote CI verification.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Vitest 5's happy-dom environment no longer tolerates direct `localStorage` assignment in the five affected unit-test files.

What changed (test files only):
- `useEventLoader.test.ts`, `InvitationDialog.test.ts`, `EventAccessTransfer.test.ts`, `useScheduleOverlapPreferences.test.ts`: replaced `globalThis.localStorage = createLocalStorageMock()` with `vi.stubGlobal("localStorage", createLocalStorageMock())`; `useScheduleOverlapPreferences.test.ts` also now imports `vi`.
- `RespondentsList.test.ts`: uses `vi.stubGlobal` with the seeded mock for the delete-contract test and restores the environment default with `vi.unstubAllGlobals()` in its existing `finally` block, replacing the manual `previousStorage` capture/restore.

Why:
- Vitest 5 resolves `@vitest-environment happy-dom` to happy-dom's `GlobalWindow`, whose `localStorage` is getter-only, so assignment threw `TypeError: Cannot set property localStorage of #<GlobalWindow> which has only a getter` for 17 tests. `vi.stubGlobal` installs the value via `Object.defineProperty`, matching the convention already used by most frontend tests.

Verification:
- Targeted run of the five files: 5 files / 54 tests passed.
- `TZ=UTC npm run test:unit`: 146 files / 1097 tests passed under vitest 5.
- `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build` all pass.
- No production source files changed; scope is limited to `*.test.ts` files.

Changes are staged but not yet committed, separately from the TASK-0235 type annotation change.
<!-- SECTION:FINAL_SUMMARY:END -->
