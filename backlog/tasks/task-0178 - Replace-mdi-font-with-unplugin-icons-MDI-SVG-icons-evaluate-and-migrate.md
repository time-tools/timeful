---
id: TASK-0178
title: Replace @mdi/font with unplugin-icons MDI SVG icons (evaluate and migrate)
status: To Do
assignee: []
created_date: '2026-09-08 12:11'
updated_date: '2026-09-15 15:46'
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
- [ ] #1 Icon inventory is recorded in this task, covering every app call site including string-valued props and computed icon names, the Vuetify internal alias set that relies on the mdi class font, and the unit and e2e assertions that depend on rendered mdi-* classes
- [ ] #2 Bundle-size delta is measured from dist output and recorded: @mdi/font CSS and font asset bytes versus the unplugin-icons build output, for CSS, JS, and font assets
- [ ] #3 The migration is completed with unplugin-icons + @iconify-json/mdi: @mdi/font is removed from package.json and frontend/src/main.ts, Vuetify's mdi-svg alias set is registered in frontend/src/plugins/vuetify.ts, and every app icon becomes a build-time icon import; visual parity is verified on desktop and mobile
- [ ] #4 Class-based unit and e2e assertions are updated to implementation-agnostic selectors, and there are no missing-icon regressions: every former mdi-* usage maps to an imported icon or a registered alias and renders identically in size, color inheritance, and hover state
- [ ] #5 All required frontend checks pass: lint, fmt:check, typecheck, build, test:unit, plus the relevant e2e specs
<!-- AC:END -->

<!-- TODO consider alternatives -->

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
