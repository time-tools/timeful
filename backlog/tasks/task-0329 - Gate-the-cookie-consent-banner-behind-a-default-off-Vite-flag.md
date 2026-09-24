---
id: TASK-0329
title: Gate the cookie consent banner behind a default-off Vite flag
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-24 20:57'
updated_date: '2026-09-24 21:51'
labels:
  - frontend
  - e2e
  - environment
dependencies: []
references:
  - backlog/tasks/task-0328.01 - Close-TASK-0328-E2E-optimization-review-gaps.md
  - docs/requirements/functional/fr/FR-141.md
documentation:
  - docs/environments.md
  - frontend/AGENTS.md
  - e2e/AGENTS.md
  - docs/requirements/README.md
  - e2e/inspect/AGENTS.md
priority: medium
type: feature
ordinal: 330000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The frontend still contains the cookie-consent banner and persistence flow, but it is not currently mounted and the browser E2E helpers unconditionally probe for consent UI. Add a build-time switch so deployments choose whether the banner is active, with safe default behavior when the variable is absent or blank.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `VITE_ENABLE_COOKIE_CONSENT` is declared and documented as a frontend build-time variable; only an explicit normalized `true` enables the banner, while unset, blank, false, and other values keep it disabled.
- [x] #2 When the flag is enabled, the existing cookie-consent banner is available in the app shell and preserves its current consent interactions; when disabled, the banner is not mounted.
- [x] #3 Browser E2E consent-dismissal helpers use the resolved flag and skip consent probing when the banner is disabled, while enabled-mode handling still supports delayed consent actions.
- [x] #4 Frontend and E2E regression coverage verifies disabled-by-default behavior, enabled behavior, and helper flag alignment.
- [x] #5 Environment examples, Docker/Compose build plumbing, and the environment contract reflect the new variable without changing unrelated consent persistence or cookie-settings behavior.
- [x] #6 Docker Compose requires `VITE_ENABLE_COOKIE_CONSENT` to be declared with a nonblank value, while direct frontend builds keep the default-off parser behavior for absent or blank input.
- [x] #7 Inspection consent helpers use an explicit `FRONTEND_TOOLING_MODE`, handle legacy and current consent actions deterministically, and do not silently leave ambiguous UI mounted.
- [x] #8 The app-shell gate has behavioral frontend coverage, scoped enabled-mode contract coverage remains reproducible, and FR-141 records the durable behavior.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 Changed Markdown files are formatted with npm run format:markdown
- [x] #4 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #5 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #6 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
- [x] #7 Changed-surface E2E contract passes in both disabled and enabled modes; the full matrix is not part of this task.
- [x] #8 No Swagger/API changes were made; Swagger regeneration is not applicable.
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Keep `VITE_ENABLE_COOKIE_CONSENT` required and nonblank in Docker Compose while retaining the frontend parser's default-off behavior for direct builds that omit or blank the variable.
2. Add an explicit environment-mode input for inspection consent helpers, make duplicate consent actions fail or be handled deterministically, and add focused tests for disabled, enabled, delayed, and ambiguous cases.
3. Replace the source-only App gate check with behavioral frontend coverage, add a canonical functional requirement plus requirements index row, and verify enabled behavior with the scoped Firefox contract without changing the default CI matrix.
4. Run focused regression checks first, then required frontend/E2E checks, Compose checks, Markdown formatting, and the code-graph refresh; record fail/pass evidence and reconcile the task Definition of Done and final summary.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation progress (2026-09-24): Added the default-off `VITE_ENABLE_COOKIE_CONSENT` parser and build-time constant, declared the Vite env key, mounted the existing banner in `App.vue` only when enabled, and threaded the value through Docker/Compose plus development/test/staging/production environment examples and local ignored env files. Updated both E2E consent-dismissal helpers to resolve the same root-env flag before probing.

Regression evidence: focused Firefox helper contract passed 5/5 in default-disabled mode and 5/5 with `VITE_ENABLE_COOKIE_CONSENT=true`; the real app-shell assertion saw no banner when disabled and saw the banner when enabled. The enabled/disabled synthetic helper paths also passed.

Verification completed: frontend lint (0 errors, 2 pre-existing warnings), fmt:check, typecheck, production build, and unit tests (159 files / 1298 tests) pass; E2E lint and typecheck pass; root fmt:check, Markdown lint, Compose config, and `git diff --check` pass. E2E fmt:check still reports only the two pre-existing files `specs/schedule-overlap-mobile-touch-firefox.spec.ts` and `specs/timed-event-scheduling-response-selection-firefox.spec.ts`. Codebase-memory index refreshed.

Final verification: the focused Firefox contract was rerun after the app-shell assertion was added and passed 5/5 in default-disabled mode and 5/5 with `VITE_ENABLE_COOKIE_CONSENT=true`; the code graph was refreshed again afterward. No Swagger/API changes were made, so the Swagger-specific DoD item is not applicable; root `npm run fmt:check` and the relevant E2E checks are recorded. The full E2E matrix was not rerun; the changed-surface contract is green, while package-wide fmt remains blocked only by the two documented pre-existing files.

Follow-up refinement after initial finalization: both E2E helpers now return immediately when the flag is disabled, and the inspection helper recognizes the Vue banner's `Accept all` label as well as the legacy `Agree` label. Typecheck, lint, focused Firefox runs in both modes, and the code-graph refresh were rerun successfully.

Review follow-up approved 2026-09-25: allow absent/blank Compose values, use an explicit inspection environment mode, and add a canonical functional requirement. The implementation must also address the weak App source assertion, ambiguous consent actions, enabled-mode CI coverage, and inaccurate conditional DoD claims.

Regression fail-before evidence (2026-09-25): the expanded Firefox contract failed 2/8 because blank `VITE_ENABLE_COOKIE_CONSENT` was rejected by Compose interpolation and the inspection helper left both `Agree` and `Accept all` buttons mounted; the explicit-mode check initially failed to import the missing resolver. Fail artifacts are under `/tmp/opencode/timeful-e2e-artifacts/review-task-0329-fail-before/` and `/tmp/opencode/timeful-e2e-artifacts/review-task-0329-mode-fail-before/`. After the fixes, the contract passed 8/8 in default-disabled mode and 8/8 with `VITE_ENABLE_COOKIE_CONSENT=true`.

User decision (2026-09-25): do not enable cookie consent in CI or alter the default CI matrix. Enabled behavior remains covered by the scoped Firefox contract run with `VITE_ENABLE_COOKIE_CONSENT=true`.

The earlier review note mentioning enabled-mode CI coverage is superseded by the user decision: CI remains unchanged, and enabled behavior is verified only by the scoped local Firefox contract.

User decision (2026-09-25, superseding the earlier blank-Compose decision): `VITE_ENABLE_COOKIE_CONSENT` remains required and nonblank in Docker Compose. The default-off parser behavior applies to direct frontend builds; the Compose blank-failure regression expectation was removed.

After the user's Compose decision, the final scoped Firefox contract passed 7/7 in default-disabled mode and 7/7 with `VITE_ENABLE_COOKIE_CONSENT=true`. A blank Compose value was verified to fail with the required-variable error at `/tmp/opencode/task-0329-compose-blank.out`; the default CI workflow remains unchanged.

Final verification: frontend build and full unit suite passed (159 files / 1,298 tests); an initial full-suite run timed out only in `Landing.minimal.test.ts`, which passed in isolation and on the immediate full rerun. Frontend lint has 0 errors and 2 pre-existing warnings in `NewSignUp.test.ts`; E2E lint/typecheck and changed-file formatting pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Completed the default-off `VITE_ENABLE_COOKIE_CONSENT` feature and its review follow-ups.

- The existing `CookieConsent` component is mounted only when the normalized flag is `true`.
- Direct frontend builds default the flag off when it is absent or blank; Docker Compose requires a declared nonblank value.
- Environment examples, Docker/Compose plumbing, frontend typing, and `docs/environments.md` describe the contract.
- Inspection consent helpers now use the explicit `FRONTEND_TOOLING_MODE` variable and handle both `Agree` and `Accept all` actions without silently leaving them mounted.
- App-shell coverage now exercises the gate behavior instead of checking source text only.
- Added `FR-141` and linked it from the requirements index.
- The default CI matrix remains unchanged; enabled behavior is covered by the scoped Firefox contract rather than CI.

## Verification

- Frontend build passed.
- Frontend unit suite passed: 159 files and 1,298 tests.
- Frontend typecheck and formatting passed; lint passed with 0 errors and 2 pre-existing warnings.
- E2E typecheck, lint, and changed-file formatting passed.
- Scoped Firefox contract passed 7/7 with the default disabled mode and 7/7 with `VITE_ENABLE_COOKIE_CONSENT=true`.
- Root Markdown formatting/lint, root formatting, Compose config, and `git diff --check` passed.
- Blank Compose input was verified to fail with the required-variable error, as required.
- Codebase-memory index was refreshed.
- No Swagger/API changes were made, so Swagger regeneration is not applicable.

## Scope notes

The full E2E matrix was not rerun. Package-wide E2E formatting still reports the two pre-existing files `specs/schedule-overlap-mobile-touch-firefox.spec.ts` and `specs/timed-event-scheduling-response-selection-firefox.spec.ts`.
<!-- SECTION:FINAL_SUMMARY:END -->
