---
id: TASK-0242
title: >-
  Replace the vendored undo icon with the project icon set's backup-restore icon
  in the timezone selector
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-15 18:37'
updated_date: '2026-09-15 18:55'
labels: []
dependencies: []
priority: low
type: enhancement
ordinal: 234000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The timezone selector reset control currently renders a vendored Material Symbols `undo` SVG at `frontend/src/components/icons/UndoIcon.vue` (used twice in `frontend/src/components/schedule_overlap/TimezoneSelector.vue`). The desired glyph is the `settings_backup_restore` look, which the project's existing MDI icon set provides as `backup-restore` and imports through `~icons/mdi/...` (unplugin-icons), consistent with every other icon in the frontend. Replace the vendored icon with the project icon-set equivalent in both reset-control variants and delete the vendored component once it has no remaining usages. This is a visual-only change: sizing, placement, and reset behavior must not change.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The timezone selector reset control renders the backup-restore icon from the project's unplugin-icons/MDI set instead of the vendored UndoIcon component
- [x] #2 Both reset-control variants keep their current size, placement, and behavior: the non-compact text variant and the compact outlined 32px variant still emit the reset event on click without opening the timezone select menu
- [x] #3 frontend/src/components/icons/UndoIcon.vue is removed once no usages remain
- [x] #4 Existing unit and e2e coverage still passes; add or adjust regression coverage if a test asserts the reset control's icon
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Required frontend checks pass: npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Research (2026-09-15):
- frontend/src/components/schedule_overlap/TimezoneSelector.vue renders the vendored UndoIcon in two reset buttons: the non-compact `v-btn icon color="primary" variant="text"` and the compact `v-btn icon size="32" variant="outlined"` (class `timezone-select__reset-button--right`). The vendored component sizes its svg to 22x22 CSS px in its own scoped style.
- The project icon pipeline is unplugin-icons (`Icons({ compiler: "vue3", scale: 1 })` in frontend/vite.config.ts) with `@iconify-json/mdi`; `backup-restore` is the MDI equivalent of Material `settings_backup_restore` and exists in the installed set (verified in node_modules).
- MDI icons render an svg at 1em, so the previous 22px rendered size must be re-applied in TimezoneSelector.
- No unit or e2e test asserts the icon markup. frontend/src/components/schedule_overlap/TimezoneSelector.test.ts uses source-substring assertions plus reset-emit behavior; e2e/specs/event-toolbar-mobile-layout.spec.ts asserts compact reset geometry via `.timezone-select__reset-button--right`.

Plan:
1. Replace the `UndoIcon` import and both `<UndoIcon />` usages with `MdiBackupRestore` from `~icons/mdi/backup-restore`, tagging each icon with class `timezone-select__reset-icon`.
2. Add a scoped `.timezone-select__reset-icon { display: block; height: 22px; width: 22px; }` rule to preserve the previous rendered size (the class lands on the child svg root, so the parent scoped style applies without `:deep`).
3. Add a unit regression assertion in TimezoneSelector.test.ts that the reset actions use the project MDI icon, matching the file's existing source-assertion style.
4. Delete the now-unused frontend/src/components/icons/UndoIcon.vue after confirming no remaining imports.
5. Run required frontend checks (lint, fmt:check, typecheck, build, test:unit) and the focused chromium-mobile e2e test for the timezone reset control on the isolated test stack.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-15: Implemented the swap. TimezoneSelector.vue now imports `MdiBackupRestore` from `~icons/mdi/backup-restore` (after the existing import groups, matching the repo's `~icons` convention), both reset buttons render `<MdiBackupRestore class="timezone-select__reset-icon" />`, and a scoped `.timezone-select__reset-icon { display: block; height: 22px; width: 22px; }` rule preserves the previous rendered size. Added the source-level regression assertion in TimezoneSelector.test.ts (two icon usages, no UndoIcon reference) and deleted frontend/src/components/icons/UndoIcon.vue after `rg` confirmed no remaining imports outside task records.

Evidence so far: focused unit file 24/24 passed; full `npm run test:unit` 148 files / 1106 tests passed; `npm run lint` 0 errors (2 pre-existing NewSignUp.test.ts warnings); `npm run fmt:check` clean; `npm run typecheck` passed; `npm run build` passed. A temporary scratch test confirmed the class prop falls through to the icon's svg root (tagName svg, class present, viewBox 0 0 24 24), so the 22px rule applies; scratch file removed afterwards.

FR-109 ("timezone-reset control shall use a counter-clockwise arrow icon") stays satisfied: the MDI backup-restore glyph is the project icon set's counter-clockwise restore arrow. TASK-0178's completed note that UndoIcon "stays as is" is superseded by this task.

2026-09-15: Final verification after the last test edit: lint 0 errors (2 pre-existing NewSignUp.test.ts warnings), fmt:check clean, typecheck passed, test:unit 148 files / 1106 tests passed, root format:markdown:check clean. Focused chromium-mobile e2e run passed: 1 passed (42.4s), log at /tmp/opencode/task-0238-e2e-mobile.log. graphify update . completed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replaced the last vendored Material Symbols icon with the project's unplugin-icons MDI set, per request.

Why: the timezone reset control rendered a vendored Material Symbols `undo` SVG (frontend/src/components/icons/UndoIcon.vue). It now renders `MdiBackupRestore` from `~icons/mdi/backup-restore` (Material `settings_backup_restore`), matching every other frontend icon.

Changes:
- frontend/src/components/schedule_overlap/TimezoneSelector.vue: both reset buttons (non-compact text variant and compact 32px outlined variant) render `<MdiBackupRestore class="timezone-select__reset-icon" />`; scoped `.timezone-select__reset-icon { display: block; height: 22px; width: 22px; }` preserves the previous rendered size; the import follows the repo's `~icons` import convention.
- frontend/src/components/schedule_overlap/TimezoneSelector.test.ts: new source-level regression assertion (two icon usages, the import, the 22px CSS block, and no UndoIcon reference).
- Deleted frontend/src/components/icons/UndoIcon.vue after verifying no remaining references.

Evidence: frontend lint 0 errors (2 pre-existing NewSignUp.test.ts warnings), fmt:check clean, typecheck passed, build passed, test:unit 148 files / 1106 tests passed; focused e2e `specs/event-toolbar-mobile-layout.spec.ts` "mobile timezone control keeps its fixed width when the reset button appears" (chromium-mobile, isolated test stack) passed; root `format:markdown:check` clean. A temporary scratch test confirmed the class lands on the rendered svg root (tagName svg, viewBox 0 0 24 24) so the 22px rule applies; the scratch file was removed.

Risk/notes: the reset glyph intentionally changes from the plain undo arrow to the backup-restore (clock plus counter-clockwise circular arrow) glyph; FR-109's counter-clockwise-arrow requirement remains satisfied. TASK-0178's completed note that UndoIcon "stays as is" is superseded by this task. No e2e spec asserts the glyph itself; the unit source assertion guards the icon choice.
<!-- SECTION:FINAL_SUMMARY:END -->
