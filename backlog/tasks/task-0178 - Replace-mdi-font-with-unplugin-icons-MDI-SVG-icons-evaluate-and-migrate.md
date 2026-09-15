---
id: TASK-0178
title: Replace @mdi/font with unplugin-icons MDI SVG icons (evaluate and migrate)
status: Done
assignee:
  - opencode
created_date: '2026-09-08 12:11'
updated_date: '2026-09-15 16:39'
labels:
  - frontend
  - vuetify
dependencies:
  - TASK-0173
references:
  - frontend/package.json
  - frontend/src/main.ts
  - frontend/src/plugins/vuetify.ts
  - frontend/src/components/icons/UndoIcon.vue
  - e2e/specs/schedule-overlap-pager-button-sizes.spec.ts
  - e2e/inspect/src/scenarios/event-respondents-panel-guest-edit.ts
  - 'https://vuetifyjs.com/en/features/icon-fonts/'
  - 'https://github.com/antfu/unplugin-icons'
  - 'https://icon-sets.iconify.design/mdi/'
priority: low
type: chore
ordinal: 184000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The frontend ships @mdi/font ^7.4.47 (imported at frontend/src/main.ts), which delivers the full MDI icon font CSS plus woff2, woff, ttf, and eot files to the bundle.
Vuetify 4 documents MDI SVG delivery as the production-recommended shape, so replace @mdi/font with build-time SVG icons using unplugin-icons and the @iconify-json/mdi collection, and register Vuetify's mdi-svg alias set in frontend/src/plugins/vuetify.ts so framework-internal icons render as SVG too.

The MDI package family (@mdi/font, @mdi/js, @mdi/svg) has had no upstream release since 7.4.47 (2023-12-27), so its glyph data is frozen.
That is accepted here to keep glyph-level visual parity, while the delivery tooling (unplugin-icons) stays maintained.
Switching to a maintained icon family (for example Material Symbols or Lucide) was considered and rejected for scope: it changes rendered glyphs and would require per-icon visual review.
The repo already vendors one Material Symbols icon for the timezone reset (frontend/src/components/icons/UndoIcon.vue, FR-109), which stays as is.

Known complexity: string-valued icon props (append-inner-icon, false-icon, on-icon, and similar) and computed icon names in EventItem.vue, RespondentsList.vue, and Dashboard.vue cannot remain mdi-* strings, and unit and e2e assertions that depend on rendered .mdi classes must be updated to implementation-agnostic selectors.
Any rendered icon must remain visually identical in size, color inheritance, and hover state.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Icon inventory is recorded in this task, covering every app call site including string-valued props and computed icon names, the Vuetify internal alias set that relies on the mdi class font, and the unit and e2e assertions that depend on rendered mdi-* classes
- [x] #2 Bundle-size delta is measured from dist output and recorded: @mdi/font CSS and font asset bytes versus the unplugin-icons build output, for CSS, JS, and font assets
- [x] #3 The migration is completed with unplugin-icons + @iconify-json/mdi: @mdi/font is removed from package.json and frontend/src/main.ts, Vuetify's mdi-svg alias set is registered in frontend/src/plugins/vuetify.ts, and every app icon becomes a build-time icon import; visual parity is verified on desktop and mobile
- [x] #4 Class-based unit and e2e assertions are updated to implementation-agnostic selectors, and there are no missing-icon regressions: every former mdi-* usage maps to an imported icon or a registered alias and renders identically in size, color inheritance, and hover state
- [x] #5 All required frontend checks pass: lint, fmt:check, typecheck, build, test:unit, plus the relevant e2e specs
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
## Plan (2026-09-15)

Scope confirmed by the retargeting review and existing acceptance criteria; no new decisions introduced.

1. Baseline metrics: run `npm run build` on the current tree and record dist CSS/JS/font bytes plus `.mdi-*:before` rule count (sanity-check review comment #2).
2. Tooling: add devDependencies `unplugin-icons` and `@iconify-json/mdi`; register `Icons({ compiler: "vue3", scale: 1 })` in `frontend/vite.config.ts` and `frontend/vitest.config.mjs`; add `unplugin-icons/types/vue` to `frontend/env.d.ts`.
3. Vuetify internals: in `frontend/src/plugins/vuetify.ts` register `{ aliases, mdi }` from `vuetify/iconsets/mdi-svg` with `defaultSet: "mdi"`; remove the `@mdi/font` CSS import from `frontend/src/main.ts` and the dependency from `frontend/package.json`.
4. Migrate all app call sites (48 files, 129 occurrences): explicit `import MdiX from "~icons/mdi/<kebab-name>"`; replace `<v-icon>mdi-x</v-icon>` with child components; replace string icon props (`append-inner-icon`, `false-icon`, `prepend-icon`) with component bindings; computed name ternaries in EventItem.vue, RespondentsList.vue, and Dashboard.vue return icon components; add `aria-label`s to the time-grid pager buttons so e2e can use role-based locators.
5. Tests: update the 8 unit test files that assert `mdi-*` text/classes to assert imported icon component identity, bound prop identity (the imported component), or semantic app classes/ARIA; update `schedule-overlap-pager-button-sizes.spec.ts` to role-based locators and the `event-respondents-panel-guest-edit` inspect scenario to `.respondent-name-line`/`.respondent-edit-status` + `aria-disabled`.
6. Verify: `lint`, `fmt:check`, `typecheck`, `build`, `test:unit`; targeted e2e (`schedule-overlap-pager-button-sizes`, respondents/timed-event specs) and visual spot-checks on desktop and mobile; record post-change dist metrics and the delta in this task.
7. Finalize per the task finalization guide with acceptance-criteria evidence.

Risks/notes: set unplugin-icons `scale: 1` so SVG is 1em and matches the font glyph box; verify `fill: currentColor` inherits Vuetify text color and hover styles; check unit-test shallow stubs can match imported icon components (fallback: assert semantic app classes/ARIA).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Migration implemented: unplugin-icons + @iconify-json/mdi registered in vite.config.ts and vitest.config.mjs; vuetify.ts now uses the mdi-svg alias set; @mdi/font removed from main.ts and package.json; all 48 app files migrated to explicit ~icons/mdi imports (slot children for <v-icon>, component bindings for icon props, computed ternaries replaced by computed icon returns/functions). Unit tests updated to component/state assertions; NewEvent component-import guard regex extended to accept ~icons imports. E2E pager spec now uses aria-label role locators; time-grid pager buttons got aria-labels; inspect scenario uses .respondent-name-line/.respondent-edit-status. Unit suite: 148 files / 1105 tests pass. Typecheck passes.

AC#1 icon inventory (recorded 2026-09-15). Scope: 48 distinct MDI icons, 108 import lines, 127 non-test usage sites across 47 app files, plus icon references in 7 unit test files for identity assertions. Per-file inventory (icons imported from `~icons/mdi/...`):
- src/App.vue: github, menu
- src/components/AlertText.vue: alert-circle
- src/components/AuthUserMenu.vue: cog, logout, message
- src/components/AutoSnackbar.vue: close
- src/components/calendar_permission_dialogs/CreateAccount.vue: check-circle
- src/components/calendar_permission_dialogs/MarkAvailabilityDialog.vue: calendar-sync
- src/components/CreateSpeedDial.vue: account-group, calendar-cursor, close, plus
- src/components/EditorDialogHeader.vue: close, information-outline
- src/components/event/EmailInput.vue: account
- src/components/EventItem.vue: account-group, account-multiple, calendar, calendar-month, calendar-range, check, chevron-right, dots-vertical
- src/components/EventType.vue: chevron-down, folder-plus
- src/components/ExpandableSection.vue: chevron-down
- src/components/FAQ.vue: plus
- src/components/FeatureNotReadyDialog.vue: close
- src/components/Footer.vue: calendar-blank, github, reddit
- src/components/FriendItem.vue: chevron-right
- src/components/general/UserChip.vue: account, close
- src/components/GuestDialog.vue: alert-circle, close
- src/components/home/Dashboard.vue: dots-horizontal, folder-plus, menu-down, menu-right, plus
- src/components/NewDialog.vue: close
- src/components/NewEvent.vue: alert-circle, checkbox-blank-off-outline, information-outline
- src/components/NewGroup.vue: chevron-down
- src/components/NewSignUp.vue: checkbox-blank-off-outline
- src/components/OverflowGradient.vue: chevron-down
- src/components/schedule_overlap/ConfirmDetailsDialog.vue: account, close, map-marker, text
- src/components/schedule_overlap/EditingAvailabilityAs.vue: pencil
- src/components/schedule_overlap/EventOptions.vue: tune-vertical
- src/components/schedule_overlap/GCalWeekSelector.vue: chevron-left, chevron-right
- src/components/schedule_overlap/RespondentsList.vue: account, check, content-copy, delete, dots-vertical, lock, pencil
- src/components/schedule_overlap/ScheduleOverlapDaysOnlyGrid.vue: chevron-left, chevron-right, close, information-outline
- src/components/schedule_overlap/ScheduleOverlapMobileOverlay.vue: calendar, close, information-outline
- src/components/schedule_overlap/ScheduleOverlapSidebar.vue: calendar, chevron-right, close
- src/components/schedule_overlap/ScheduleOverlapTimeGrid.vue: chevron-down, chevron-left, chevron-right, close, information-outline
- src/components/settings/CalendarAccount.vue: alert-circle, chevron-right, close
- src/components/settings/CalendarAccounts.vue: chevron-down
- src/components/settings/CalendarTypeSelector.vue: calendar-sync
- src/components/SignInDialog.vue: alert-circle, arrow-left
- src/components/sign_up_form/SignUpBlock.vue: account, check, pencil, undo
- src/components/sign_up_form/SignUpForSlotDialog.vue: close
- src/components/TimefulImportDialog.vue: close
- src/components/UpvoteRedditSnackbar.vue: arrow-up-bold, close
- src/components/UserAvatarContent.vue: apple, microsoft-outlook
- src/components/When2meetImportDialog.vue: close
- src/views/Event.vue: calendar-check, calendar-today, content-copy, pencil, plus, refresh, trash-can-outline
- src/views/Friends.vue: account-plus
- src/views/Home.vue: plus
- src/views/SignIn.vue: alert-circle, arrow-left

AC#1 inventory, string props and computed names. String-valued icon props (now component bindings): GuestDialog.vue:31 append-inner-icon MdiAlertCircle; NewEvent.vue:47 append-inner-icon MdiAlertCircle and :238, :353, :407 false-icon MdiCheckboxBlankOffOutline; NewSignUp.vue:183 false-icon MdiCheckboxBlankOffOutline; ConfirmDetailsDialog.vue:130 prepend-icon MdiMapMarker and :137 prepend-icon MdiText; ScheduleOverlapSidebar.vue:142 prepend-icon MdiCalendar; ScheduleOverlapMobileOverlay.vue:58 prepend-icon MdiCalendar. Computed icon names (now return components): EventItem.vue eventIcon() returns MdiAccountGroup/MdiCalendarRange/MdiCalendarMonth/MdiCalendar; Dashboard.vue folderToggleIcon() returns MdiMenuDown/MdiMenuRight; RespondentsList.vue renders MdiPencil/MdiLock via v-if/v-else. Vuetify internal alias set (previously served by the mdi class font, now by vuetify/iconsets/mdi-svg with defaultSet "mdi"): collapse, complete, cancel, close, delete, clear, success, info, warning, error, prev, next, checkboxOn, checkboxOff, checkboxIndeterminate, delimiter, sortAsc, sortDesc, expand, menu, subgroup, dropdown, radioOn, radioOff, edit, ratingEmpty, ratingFull, ratingHalf, loading, first, last, unfold, file, plus, minus, calendar, treeviewCollapse, treeviewExpand, tableGroupExpand, tableGroupCollapse, eyeDropper, upload, color, command, ctrl, space, shift, alt, enter, arrowup, arrowdown, arrowleft, arrowright, backspace, play, pause, fullscreen, fullscreenExit, volumeHigh, volumeMedium, volumeLow, volumeOff, search. Class-based fallout updated: 23 assertions in EventItem.test.ts, GuestDialog.test.ts, NewEvent.test.ts, RespondentsList.test.ts, ScheduleOverlapMobileOverlay.test.ts, ScheduleOverlapSidebar.test.ts, and Event.test.ts (now imported-icon identity, bound prop identity, or app classes/ARIA); schedule-overlap-pager-button-sizes.spec.ts uses role locators after pager aria-labels were added; inspect scenario event-respondents-panel-guest-edit.ts uses .respondent-name-line/.respondent-edit-status plus aria-disabled.

AC#2 bundle-size delta (raw dist bytes, from `npm run build` in frontend). Baseline (HEAD with @mdi/font ^7.4.47): CSS 732,671 B total, index CSS 592,854 B, 7,462 .mdi-*:before rules; JS 1,592,582 B across 41 files; fonts 3,856,604 B across 26 files (MDI woff2 403,216 B plus woff/ttf/eot variants); dist total 6,794,494 B. After migration: CSS 409,276 B total (−323,395 B, −44.1%), index CSS 268,751 B, 0 .mdi-*:before rules; JS 1,618,391 B across 46 files (+25,809 B, +1.6%, including per-icon chunks); fonts 249,864 B across 22 files (−3,606,740 B, −93.5%; only Chivo Mono and DM Sans woff/woff2 remain); dist total 2,890,409 B (−3,904,085 B, −57.5%). Method: baseline measured from a clean `git archive HEAD` copy built against the same locked node_modules (@mdi/font temporarily reinstalled with --no-save, then pruned); migrated numbers from a fresh working-tree build.

AC#3–#5 verification (2026-09-15). Frontend required checks: npm run lint (0 errors; 2 pre-existing NewSignUp.test.ts one-component-per-file warnings), npm run fmt:check pass, npm run typecheck pass, npm run build pass, npm run test:unit 148 files / 1105 tests pass. E2E on the isolated test stack: chromium pager/styling/mobile specs 13 passed with 5 expected skips; firefox weekly/group/signup/dashboard-folders 12 passed; full suite all projects 112 passed, 21 skipped, 2 failed in timed-event-access-transfer-firefox (dev-server frontend at 2 workers did not show the 'Edit Reload source response' button within the 5s expect and 40s test budgets). Both failures pass in isolation (8/8) and under CI's E2E_FRONTEND=bundled mode (8/8); the failing aria-label predates this migration, so they are the documented dev-server load-budget flake rather than an icon regression. Visual parity: extracted video frames (E2E_VIDEO=on) of chromium desktop and mobile event pages show migrated SVG icons rendering with no missing glyphs, expected size, and color inheritance; production styling checks in chromium-production-desktop and chromium-production-mobile pass. E2E package lint/typecheck/fmt:check pass; root format:markdown:check passes. Plan deviations: 7 unit test files needed updates (the plan estimated 8); NewEvent.test.ts also needed its component-import guard regex widened to accept ~icons imports; pager aria-labels were added in ScheduleOverlapTimeGrid only (the spec exercises that grid).
<!-- SECTION:NOTES:END -->

<!-- TODO consider alternatives -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-15 15:46
---
Review before execution (2026-09-15). User decisions: deliver MDI glyphs via unplugin-icons + @iconify-json/mdi instead of @mdi/js; this task owns the affected test updates; the inventory, measurements, and decision live in this task record.
---

author: opencode
created: 2026-09-15 15:46
---
Verified claims: @mdi/font ^7.4.47 is imported at frontend/src/main.ts:8, and Vuetify defaults to the mdi class set (frontend/node_modules/vuetify/lib/icons.js:17), so the font also serves framework-internal aliases (frontend/node_modules/vuetify/lib/iconsets/mdi.js). Dist cost: MDI rules are about 573 KB of the 592,854 B frontend/dist/assets/index-*.css (7,462 .mdi-*:before rules), plus woff2 403,216 B and the woff, ttf, and eot variants (~3.6 MB of font files in dist).
---

author: opencode
created: 2026-09-15 15:46
---
Staleness: @mdi/font, @mdi/js, and @mdi/svg all froze at 7.4.47 on 2023-12-27; none is npm-deprecated and the GitHub repos are not archived, so "unmaintained since December 2023" is the accurate framing. Keeping MDI glyphs via unplugin-icons preserves visual parity while the tooling stays maintained; a family switch was rejected for scope.
---

author: opencode
created: 2026-09-15 15:46
---
Inventory: 47 distinct mdi-* icons and 129 non-test occurrences; computed-name spots are simple ternaries in EventItem.vue:24-29, RespondentsList.vue:244-246, and Dashboard.vue:29-31; :icon="isPhone" in Event.vue:210,219 is a boolean VBtn prop, not an icon name. Fallout not previously in scope: 23 assertions in 7 unit test files (for example Event.test.ts:2166,2770 and GuestDialog.test.ts:217) and 2 e2e locations (schedule-overlap-pager-button-sizes.spec.ts:11,53-54 and the inspect scenario event-respondents-panel-guest-edit.ts:65-66).
---

author: opencode
created: 2026-09-15 15:46
---
Dependency TASK-0173 is Done; no active task overlaps; FR-109 keeps its vendored UndoIcon.vue and is unaffected by the icon delivery change.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Replace @mdi/font class-font delivery with build-time MDI SVG components from unplugin-icons + @iconify-json/mdi, and serve Vuetify's internal aliases through vuetify/iconsets/mdi-svg.

## What changed

- Tooling: added unplugin-icons and @iconify-json/mdi devDependencies; registered Icons({ compiler: "vue3", scale: 1 }) in frontend/vite.config.ts and frontend/vitest.config.mjs; added unplugin-icons/types/vue to frontend/env.d.ts.
- Vuetify: frontend/src/plugins/vuetify.ts registers { aliases, mdi } from vuetify/iconsets/mdi-svg with defaultSet "mdi"; removed the @mdi/font CSS import from frontend/src/main.ts and the dependency from frontend/package.json.
- App icons: 47 app files (48 distinct icons, 127 usage sites) now import explicit ~icons/mdi/<name> components; string props (append-inner-icon, false-icon, prepend-icon) bind components; computed icon ternaries in EventItem.vue, RespondentsList.vue, and Dashboard.vue return components; ScheduleOverlapTimeGrid pager buttons gained aria-labels.
- Tests: 23 class-based assertions in 7 unit test files now assert imported-icon identity, bound prop identity, or app classes/ARIA; e2e/specs/schedule-overlap-pager-button-sizes.spec.ts uses role locators; the event-respondents-panel-guest-edit inspect scenario uses .respondent-name-line/.respondent-edit-status + aria-disabled.

## Why

@mdi/font shipped ~573 KB of icon CSS (7,462 rules) plus ~3.6 MB of woff/woff2/ttf/eot into dist. MDI SVG is Vuetify's production-recommended delivery, and MDI glyph data is frozen at 7.4.47 regardless of delivery, so this keeps glyph-level visual parity while the tooling stays maintained.

## Measurements (dist, raw bytes)

CSS 732,671 → 409,276 B (−44.1%); JS 1,592,582 → 1,618,391 B (+1.6%, per-icon chunks); fonts 3,856,604 → 249,864 B (−93.5%); dist total 6,794,494 → 2,890,409 B (−57.5%). MDI .mdi-*:before rules: 7,462 → 0.

## Verification

Frontend lint, fmt:check, typecheck, build, and test:unit (148 files / 1105 tests) pass. E2E on the isolated stack: chromium pager/styling/mobile specs 13 passed, firefox weekly/group/signup/dashboard-folders 12 passed, full suite 112 passed / 21 skipped / 2 failed where timed-event-access-transfer-firefox exceeded its dev-server budgets at 2 workers; those 8 tests pass in isolation and under CI's E2E_FRONTEND=bundled mode. Desktop and mobile video frames reviewed for visual parity; production styling checks pass.

## Risks / follow-ups

None blocking. The dev-server access-transfer budget failures are pre-existing environment behavior (CI uses bundled mode for that job).
<!-- SECTION:FINAL_SUMMARY:END -->
