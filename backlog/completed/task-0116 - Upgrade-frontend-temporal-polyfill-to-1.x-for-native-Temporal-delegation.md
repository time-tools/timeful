---
id: TASK-0116
title: Upgrade frontend temporal-polyfill to 1.x for native Temporal delegation
status: Done
assignee:
  - opencode
created_date: '2026-08-30 14:48'
updated_date: '2026-08-30 15:53'
labels:
  - frontend
  - dependencies
  - temporal
dependencies: []
references:
  - >-
    https://github.com/fullcalendar/temporal-polyfill/releases/tag/temporal-polyfill%401.0.1
  - >-
    https://github.com/fullcalendar/temporal-polyfill/blob/main/polyfill/README.md
  - 'https://github.com/tc39/proposal-temporal'
  - 'https://www.npmjs.com/package/temporal-polyfill'
documentation:
  - docs/design/architecture/adr/ADR-004.md
  - docs/design/architecture/adr/ADR-002.md
  - docs/design/architecture/adr/ADR-006.md
  - frontend/AGENTS.md
  - AGENTS.md
modified_files:
  - frontend/package.json
  - frontend/package-lock.json
  - docs/design/architecture/adr/ADR-004.md
  - frontend/AGENTS.md
priority: medium
type: chore
ordinal: 123000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Outcome

Upgrade the frontend Temporal runtime from `temporal-polyfill@^0.3.2` to the latest 1.x release (1.0.4 as of Aug 2026; re-verify latest 1.x at execution time). This moves the app to current Stage 4 Temporal spec conformance and makes the runtime automatically use engine-native `Temporal` where it exists, while the package's JS polyfill keeps browsers without native Temporal (notably Safari/WebKit as of Aug 2026) working. User value: spec-accurate time behavior, bugfixes in exactly this app's risk areas (future-offset `ZonedDateTime.from`, `Duration.round`/`total`), and no extra cost on modern browsers.

## Researched background (Aug 2026; re-verify facts at execution)

- Current state: `temporal-polyfill@^0.3.2` in `frontend/package.json`, imported as a ponyfill (`import { Temporal } from "temporal-polyfill"` / `import type { Temporal }`) in roughly 100 files across `frontend/src/` and `frontend/e2e/`. Value-semantics helpers are centralized in `frontend/src/utils/temporalPrimitives.ts` (ADR-006).
- `temporal-polyfill@1.0.x` (repo: fullcalendar/temporal-polyfill) keeps the same ponyfill entrypoint and, new in 1.0, delegates to native Temporal when available; 0.3 never did. Min+gzip 19.5 kB (23.4 kB for `/full/`).
- Native Temporal is Stage 4 and shipped in Firefox 139 (2025-05-27), Chrome 144 (2026-01-13), and Node 26 (2026-05-05). Safari/JavaScriptCore had NOT shipped it as of Aug 2026. Local Node is v26.7.0, so after the upgrade vitest unit tests and Playwright Firefox e2e run on native Temporal; only real WebKit browsers exercise the JS implementation in production.

## Confirmed decisions and constraints

- Package choice is decided: stay on `temporal-polyfill`. `@js-temporal/polyfill@0.5.1` was evaluated and rejected (52.1 kB min+gzip vs 19.5 kB, spec conformance dated Mar 2025, no native delegation). Do not switch packages in this task.
- Import sites must not change. The ponyfill entrypoint `import { Temporal } from "temporal-polyfill"` remains a supported 1.x entrypoint. Do not adopt `temporal-polyfill/global` (global types/`lib` changes are out of scope), the `/fns` tree-shakeable API, or a central re-export module in this task.
- 1.0 breaking changes and why they are acceptable here: ESM-only (project is Vite/ESM); BigInt required (Chrome 67+, Firefox 68+, Safari 14+, Node 16+; browser target is modern evergreen only — no legacy support matrix to honor); non-ISO calendar systems moved to `/full/` entrypoints (no calendar usage exists in `frontend/src/`, confirmed by search); `global.min.js` removed from the NPM package (not used here).
- TypeScript: `frontend/tsconfig.app.json` uses target ES2022 with `lib: ["ES2022", "DOM", "DOM.Iterable"]`, which has no built-in Temporal types; package-provided types keep working via the existing `import type` sites. No tsconfig changes are expected; only adjust if typecheck proves otherwise, without changing import sites.
- Known 1.0 behavior deltas vs 0.3 land in this repo's known Temporal risk areas and may legitimately change test outcomes: future-offset `Temporal.ZonedDateTime.from` calculations (e.g. Europe/Berlin 2044, Brazil wall times near offset transitions) and `Duration.round`/`total`/`from` fixes. These are spec-alignment fixes, not regressions; per ADR-004, runtime verification is mandatory because many Temporal regressions are runtime-only.
- Documentation updates belong in this task, not a follow-up.

## Explicitly out of scope (deferred)

- WebKit e2e smoke coverage of the JS-polyfill path (consider as a separate follow-up if polyfill-path confidence is wanted).
- Dropping the dependency entirely once Safari ships native Temporal.
- Adoption of the `temporal-polyfill/fns` tree-shakeable API or bundle-size work beyond the upgrade itself.
- Server-side changes; this task is frontend-only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 frontend/package.json declares temporal-polyfill at the latest 1.x release available at execution time (^1.0.4 or newer 1.x), and frontend/package-lock.json resolves that 1.x version
- [x] #2 Frontend and e2e import sites are unchanged: all continue to use the temporal-polyfill ponyfill entrypoint (import { Temporal } from "temporal-polyfill" and import type), with no migration to temporal-polyfill/global, /full/, /implementation, or /fns entrypoints
- [x] #3 Frontend required checks pass from frontend/: npm run lint, npm run typecheck, npm run build, and npm run test:unit
- [x] #4 npm run test:e2e -- --project=firefox-desktop passes from frontend/ with Playwright owning the isolated test stack (mongo-test, postgres-test, server-test on 3003, Vite on 4174), never the development API on 3002
- [x] #5 Any runtime behavior difference vs temporal-polyfill 0.3 observed in unit or e2e tests is either (a) explained by a documented 1.0 conformance or bugfix change from the release notes and accepted with a note in the task, or (b) fixed with a regression test added before completion, following ADR-004's test rules
- [x] #6 ADR-004 is updated to record the temporal-polyfill 1.x runtime model: one Temporal runtime model with native engine implementation where available (Firefox 139+, Chrome 144+, Node 26+) and the JS polyfill implementation as fallback (notably Safari/WebKit until it ships native Temporal), including the updated_date and the rationale that @js-temporal/polyfill was evaluated and rejected (52.1 kB vs 19.5 kB min+gzip, older spec, no native delegation)
- [x] #7 frontend/AGENTS.md remains accurate for the upgraded dependency (the "Temporal via temporal-polyfill" line); update its wording if it would mislead a future agent about the 1.x runtime model
- [x] #8 The task's final summary records the resolved temporal-polyfill version, the check results, and any accepted behavior changes vs 0.3
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
## Implementation Plan

Research confirmed (Aug 30, 2026):
- Latest 1.x on npm is 1.0.4 (dist-tag latest). Lockfile resolves temporal-polyfill 0.3.2.
- 136 files under frontend/src and frontend/e2e import the bare ponyfill entrypoint; zero subpath imports (/global, /full, /fns, /implementation, /shim). No Temporal calendar-system usage (withCalendar / Temporal calendarId) exists; earlier calendar hits are external calendar-account features, not Temporal calendars. Default entrypoint stays valid in 1.x.
- Node v26.7.0 locally, so vitest unit tests run on native Temporal after the upgrade.
- Release notes re-verified: 1.0.1 (native delegation, June 2026 spec, BigInt required, ESM-only, /full for non-ISO calendars, Duration.round/total/from + ZonedDateTime.from fixes), 1.0.2 (spec 2026-07-27, Duration.from invalid-string fix), 1.0.3 (require(esm) exports fix), 1.0.4 (/full native-delegation fix). Type-level risks: temporal-spec 1.0 renamed options interfaces and restricted Duration.total unit args; expect possible typecheck fallout at call sites.

Steps:
1. Upgrade dependency in frontend/: npm install temporal-polyfill@^1.0.4 (updates package.json and package-lock.json only; no import-site changes).
2. Verify AC #2 mechanically: no subpath temporal-polyfill imports; no source diffs outside package files.
3. Run required checks from frontend/: npm run lint, npm run typecheck, npm run build, npm run test:unit. Fix any type-level fallout within scope (no import-site migration; no tsconfig changes unless typecheck proves otherwise).
4. If unit/e2e runtime deltas vs 0.3 appear: map each to a documented 1.0 conformance/bugfix change and note it in the task (AC #5a), or fix with a regression test per ADR-004 (AC #5b).
5. Run browser E2E: npm run test:e2e -- --project=firefox-desktop from frontend/ (Playwright owns the isolated stack on 3003/4174).
6. Update docs/design/architecture/adr/ADR-004.md: record the 1.x runtime model (one Temporal runtime model; native engine Temporal where available: Firefox 139+, Chrome 144+, Node 26+; JS polyfill fallback notably Safari/WebKit until it ships native Temporal), bump updated_date, and add the @js-temporal/polyfill rejection rationale (52.1 kB vs 19.5 kB min+gzip, older spec conformance, no native delegation). Follow docs authoring rules (one sentence per source line).
7. Review frontend/AGENTS.md "Temporal via temporal-polyfill" line; update wording only if it would mislead about the 1.x runtime model.
8. Format changed Markdown files (npm run format:markdown per DoD).
9. Finalize: verify each AC with evidence, record final summary (resolved version, check results, accepted behavior deltas vs 0.3), mark Done.

Risks/checkpoints:
- Typecheck is the most likely failure point (temporal-spec 1.0 type changes, e.g. Duration.total unit restriction). Handle per-task only; no import-site migration.
- Runtime deltas concentrated in known risk areas: future-offset ZonedDateTime.from (Europe/Berlin 2044, Brazil transitions) and Duration.round/total/from. Vitest runs on native Temporal (Node 26), so unit failures here are likely spec-alignment fixes, not regressions.
- e2e depends on Docker test stack availability; if unavailable, report the blocked check per workflow.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Upgrade applied: temporal-polyfill ^0.3.2 -> ^1.0.4 (npm dist-tag latest verified at execution time). Lockfile resolves 1.0.4 with new transitive deps temporal-spec@1.0.1 and temporal-utils@1.0.2. No source or import-site changes needed for the upgrade itself (AC #1, #2 verified mechanically: 136 files import the bare ponyfill entrypoint; zero /global /full /fns /implementation /shim subpaths; no Temporal calendar-system usage).

Required checks after upgrade: lint PASS, typecheck PASS, build PASS, test:unit initially 3 failures in 952 tests (eventDateRules 'uses the viewed week as the reference date for weekly events', scheduleOverlap.regressions 'exports weekly DOW schedules from the displayed week instead of the seed week', eventEditorSchedule 'builds weekly canonical recurrence output from current-week anchors').

Root cause of the 3 failures (A/B verified by stashing the upgrade and re-running on 0.3.2: all 32 tests in the 3 files pass on 0.3.2): temporal-polyfill 1.0 delegates to native Temporal when available, and Node 26 has native Temporal. Native Temporal.Now reads the engine clock directly, so vitest's Date-based fake timers (vi.useFakeTimers + vi.setSystemTime) no longer control the runtime clock that Temporal.Now.plainDateISO/zonedDateTimeISO report. Under 0.3 the JS implementation consulted the patched Date clock, so the old fake-timer approach worked. This is the documented 1.0 feature 'Uses native Temporal when available (prior versions never did)', not an app-code behavior change.

Fix per ADR-004 test rules (fake clocks control the runtime clock, not Date): added src/test/fakeTemporalNow.ts exporting setFakeTemporalNow(instant|string) and restoreFakeTemporalNow(). It spies Temporal.Now's clock methods (instant, plainDateISO, plainDateTimeISO, plainTimeISO, zonedDateTimeISO) to derive values from a pinned Instant via instant.toZonedDateTimeISO(timeZone ?? Temporal.Now.timeZoneId()); timeZoneId itself is left real. Updated the 3 failing tests plus the adjacent eventEditorSchedule weeklyAnchorInstant test to pin Temporal.Now instead of vi.setSystemTime, with afterEach(restoreFakeTemporalNow) cleanup.

Post-fix checks re-run: test:unit PASS (137 files, 952 tests), lint PASS, typecheck PASS, build PASS.

Browser e2e environment fix (user-approved decision): @playwright/test was bumped to 1.62.1 by dependabot earlier today (commit 7bcd6f42), but the repo flake pins nixpkgs playwright-driver 1.61.1, whose browsers ship firefox-1532; Playwright 1.62.1 requires firefox-1538 and no nixpkgs rev currently provides a matching patched build (nixos-unstable and master still at driver 1.61.1). The downloaded firefox-1538 is not nix-patched and cannot launch on this NixOS host (missing libs). Realigned @playwright/test back to ^1.61.0 in frontend/package.json and package-lock.json to match the flake-provided browsers; e2e then passes: 22 passed, 1 skipped (by design), 0 failed, with Playwright owning the isolated test stack (mongo-test, postgres-test, server-test on 3003, Vite on 4174). The Firefox 1532 build is far newer than 139, so e2e exercised the native Temporal delegation path.

Follow-up to surface, not created: dependabot will likely re-propose the @playwright/test 1.62 bump; the durable fix is updating the flake nixpkgs pin to a rev whose playwright-driver provides a matching browser build first.

Post-completion fix (2026-10-08, added by hand with user approval because completed tasks are not editable through Backlog MCP): the migration missed timezoneDateRules.test.ts 'uses the rendered week when deriving weekly schedule offsets', which still pinned the clock with vi.useFakeTimers + vi.setSystemTime(2026-03-18). Under native Temporal it used the real date, so it passed until about 2026-10-04 and then failed (rendered week +3 crosses the Europe/Vienna DST end on 2026-10-25): 'expected -60 to be -120'. Switched it to setFakeTemporalNow with afterEach(restoreFakeTemporalNow); no production code changed. Fail-before: the old test fails with that exact assertion when Temporal.Now is shimmed to ignore fake timers (simulating native Temporal; this host's Node 25.9.0 lacks native Temporal, so the unshimmed run passes). Pass-after: the same shimmed run passes 5/5; full test:unit PASS (159 files, 1298 tests); lint, oxfmt and vue-tsc clean.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Upgraded the frontend Temporal runtime from `temporal-polyfill@^0.3.2` to `^1.0.4` (latest 1.x on npm at execution time, dist-tag `latest` re-verified). `frontend/package-lock.json` resolves 1.0.4, pulling in `temporal-spec@1.0.1` and `temporal-utils@1.0.2`. The runtime now delegates to engine-native `Temporal` where available (Firefox 139+, Chrome 144+, Node 26+) and falls back to the package JS polyfill elsewhere (notably Safari/WebKit).

### Scope kept

- No import-site changes: 136 files under `frontend/src/` and `frontend/e2e/` keep the bare ponyfill entrypoint; zero `/global`, `/full`, `/fns`, `/implementation`, `/shim` imports.
- No Temporal calendar-system usage exists, so the default entrypoint is sufficient.
- No tsconfig changes needed.

### Check results

- `npm run lint` PASS
- `npm run typecheck` PASS (under both the temporary 1.62.1 and final 1.61.0 Playwright types)
- `npm run build` PASS
- `npm run test:unit` PASS: 137 files, 952 tests
- `npm run test:e2e -- --project=firefox-desktop` PASS: 22 passed, 1 skipped (by design), 0 failed, with Playwright owning the isolated stack (mongo-test, postgres-test, server-test on 3003, Vite on 4174)

### Accepted behavior changes vs 0.3

Exactly one runtime delta was observed, and it was both explained and fixed (AC #5a + #5b):

- 1.0 native delegation means vitest's `Date`-based fake timers (`vi.useFakeTimers` + `vi.setSystemTime`) no longer control `Temporal.Now`; native `Temporal.Now` reads the engine clock. A/B verified: the same 3 clock-dependent tests pass on 0.3.2 and fail on 1.0.4.
- Fixed per ADR-004 test rules with the new shared helper `src/test/fakeTemporalNow.ts` (`setFakeTemporalNow` / `restoreFakeTemporalNow`), which pins `Temporal.Now`'s clock methods to a fixed `Instant`. Four tests now pin `Temporal.Now` instead of `Date`-based fake timers.
- No app-level behavior changes were observed: all other 949 unit tests and the full e2e suite passed unchanged, including the known risk areas (future-offset `ZonedDateTime.from`, `Duration.round`/`total`).

### Related dependency realignment (user-approved)

`@playwright/test` was realigned from the same-day dependabot bump `^1.62.1` back to `^1.61.0` (commit 7bcd6f42 had broken the local e2e environment): the repo flake pins nixpkgs `playwright-driver` 1.61.1 (browsers with firefox-1532), Playwright 1.62.1 requires firefox-1538, and no nixpkgs rev currently ships a matching patched build. Dependabot will likely re-propose the bump; update the flake's nixpkgs pin first.

### Docs

- ADR-004: new "Temporal runtime implementation" subsection (1.x runtime model, native delegation browsers, JS polyfill fallback, `@js-temporal/polyfill` rejection rationale: 52.1 kB vs 19.5 kB min+gzip, March 2025 conformance, no native delegation), `updated_date` bumped to 2026-08-30, and a fake-clock rule added to the test rules.
- `frontend/AGENTS.md`: the "Temporal via temporal-polyfill" line now notes the ponyfill entrypoint delegates to native engine `Temporal` where available.
- Changed Markdown formatted via `npm run format:markdown`.
<!-- SECTION:FINAL_SUMMARY:END -->
