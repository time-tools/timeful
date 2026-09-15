---
id: TASK-0239
title: >-
  Prebundle vuetify/components/VAlert so cold-cache dev-server loads never
  re-optimize dependencies mid-session
status: Done
assignee:
  - opencode
created_date: '2026-09-15 19:13'
updated_date: '2026-09-15 19:31'
labels: []
dependencies: []
references:
  - frontend/vite.config.ts
  - frontend/src/views/Event.vue
  - .github/workflows/e2e-ci.yml
  - >-
    https://github.com/time-tools/timeful/actions/runs/35010848350/job/104522124067
  - e2e/AGENTS.md
priority: medium
type: bug
ordinal: 235000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Vite dev server's `optimizeDeps.include` list in `frontend/vite.config.ts` exists to prebundle every Vuetify component auto-imported in `frontend/src`, so the dep optimizer never discovers a component lazily during the first page load. `<v-alert>` is auto-imported by `vite-plugin-vuetify` in SFCs such as `frontend/src/views/Event.vue`, but `vuetify/components/VAlert` is absent from that list.

On a cold dep cache, the first transform of an SFC using `<v-alert>` makes Vite discover the unbundled dependency, re-run the optimizer, and swap the `.vite/deps` generation. All hashed chunk filenames change, so the browser's still-pending requests for the previous generation fail with a burst of `The file does not exist at ... which is in the optimize deps directory` pre-transform errors, and Vite then full-reloads the page. The E2E suites still pass because the reload happens during the first page load, but this is the mid-session reload class the include list was added to prevent; a reload later in a journey would discard in-memory app state.

Evidence: E2E CI run https://github.com/time-tools/timeful/actions/runs/35010848350 (firefox-touch job 104522124067: 73 errors; chromium job 104522124077: 50 errors) shows the burst within about 100 ms of the first test page load, while the bundled firefox-desktop job shows none. A local cold-cache dev-server probe logs `vite:deps new dependencies found: vuetify/components/VAlert`, then `optimized dependencies changed. reloading`, then the same missing hashed chunks (`defineComponent-Ja3a75sb.js`, `VList-BQaYi8EC.js`, `deepEqual-B0dF8smM.js`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 With a cold Vite dependency cache, the dev server prebundles every Vuetify component auto-imported from `<v-*>` tags in `frontend/src`, so first-transforming an SFC that uses `<v-alert>` (for example `frontend/src/views/Event.vue`) produces no `new dependencies found`, no `optimized dependencies changed. reloading`, and no `The file does not exist at ... which is in the optimize deps directory` errors.
- [x] #2 Starting the dev-server E2E suites (chromium, firefox-touch) against a cold dep cache no longer emits the stale-optimized-chunk pre-transform error burst right after the first page load.
- [x] #3 The required frontend checks pass: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit`.
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
1. Add `vuetify/components/VAlert` to `optimizeDeps.include` in `frontend/vite.config.ts`, keeping the list aligned with the components auto-imported in `frontend/src` (it is the only used component missing from the list).
2. Verify against a cold dep cache with an isolated probe: a temporary Vite config that mirrors the real plugins and include list but points `cacheDir` at a temp directory, then start the server and request `frontend/src/views/Event.vue`. Confirm the log shows `dependencies optimized` and no `new dependencies found`, no `optimized dependencies changed. reloading`, and no `The file does not exist at ... which is in the optimize deps directory` errors.
3. Re-derive the used-component set from Vuetify's `importMap.json` and the `<v-*>` tags in `frontend/src` to confirm no other auto-imported component is missing from the include list.
4. Run the required frontend checks: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`.
5. Record objective evidence for each acceptance criterion and finalize the task.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the one-line fix: added `"vuetify/components/VAlert"` to `optimizeDeps.include` (no other changes; the used-component audit found no other missing entry).

Cold-cache probe: started a temporary Vite dev server with `cacheDir` at `frontend/node_modules/.vite-probe` and `DEBUG=vite:deps`. Startup logged `dependencies optimized`; after requesting `/src/views/Event.vue` (which auto-imports VAlert) the log had no `new dependencies found`, no `optimized dependencies changed. reloading`, and no `The file does not exist at ... optimize deps directory`. Probe config and cache were removed afterward.

Pre-fix probe (same method) reproduced the bug and matched CI exactly: `new dependencies found: vuetify/components/VAlert`, `optimized dependencies changed. reloading`, and missing `defineComponent-Ja3a75sb.js` chunks.

Local E2E: `--project=firefox-touch --workers=1` -> 7 passed with zero dep-optimizer error lines; `--project=chromium-desktop --project=chromium-mobile --workers=2` -> 48 passed, 20 skipped, zero dep-optimizer error lines.

Local `E2E_FRONTEND=bundled --project=firefox-desktop --workers=2` was aborted by the user after two `timed-event-access-transfer-firefox` recordings hit their 40s budget (43.5s/50.5s) while other access-transfer journeys passed. Bundled mode runs `vite build` + preview, which does not use `optimizeDeps`, so this change cannot affect it; the firefox-desktop job was green in CI run 35010848350.

Checks run from `frontend/`: lint (0 errors, 2 pre-existing vue/one-component-per-file warnings in `src/components/NewSignUp.test.ts`), fmt:check, typecheck, build, test:unit (148 files, 1106 tests).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Changed `frontend/vite.config.ts` so `optimizeDeps.include` also prebundles `vuetify/components/VAlert`, closing the last gap between the Vuetify components auto-imported in `frontend/src` and the components prebundled at dev-server startup.

Why: `<v-alert>` (used by `Event.vue` and the access-transfer views) was discovered lazily on the first transform of those SFCs against a cold dep cache. Vite then re-ran the optimizer, swapped the `.vite/deps` generation, deleted the hashed chunks the browser was still requesting, and emitted the `The file does not exist at ... which is in the optimize deps directory` pre-transform burst followed by a full page reload. The E2E jobs still passed because the reload happened during the first page load, but this is the mid-session reload class the include list was added to prevent.

Verification:
- Cold-cache isolated probe (same plugins and include list, temp `cacheDir`): log shows `vite:deps dependencies optimized` only; requesting `src/views/Event.vue` produced no `new dependencies found`, no `optimized dependencies changed. reloading`, and no missing-chunk errors.
- Re-derived the used-component set from Vuetify's `importMap.json` and all `<v-*>` tags in `frontend/src` (42 tags, 31 include entries): no remaining gaps.
- E2E firefox-touch: 7 passed. Chromium desktop + mobile: 48 passed, 20 skipped. Both local logs contain zero re-optimization or stale-chunk lines.
- Frontend checks: lint (0 errors; 2 pre-existing warnings in `NewSignUp.test.ts`), fmt:check, typecheck, build, and unit tests (148 files / 1106 tests) all pass.
- The firefox-desktop bundled run was aborted locally after two heavy recorded access-transfer journeys exceeded their 40s budget; that project builds and previews production assets, which never run the dep optimizer, and the same job was green in CI run 35010848350 for this branch.

Risk/follow-up: the include list stays hand-maintained, so a Vuetify component added to a template without a matching `optimizeDeps.include` entry will reintroduce the reload. Scope was intentionally kept to the one missing entry.
<!-- SECTION:FINAL_SUMMARY:END -->
