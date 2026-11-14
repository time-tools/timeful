---
id: TASK-0351
title: Fix floating Event name label clipped in the New event editor
status: Done
assignee:
  - '@claude'
created_date: '2026-10-08 18:49'
updated_date: '2026-10-08 18:58'
labels:
  - bug
  - frontend
dependencies: []
modified_files:
  - frontend/src/components/NewEvent.vue
  - e2e/specs/new-event-name-label-layout.spec.ts
priority: low
ordinal: 309000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The floating label of the first outlined field ("Event name (required)") in the New event editor has its top cut off directly under the dialog title.
The form body `v-card-text` is the scroll container (`overflow: auto`) and has only 4px top padding, while an outlined Vuetify field draws its floating label centred on the outline's top border, so part of the label sits above the field box and gets clipped.
`NewSignUp.vue` and `NewGroup.vue` share the card-text classes but their first field is a label-less `solo` field, so they are not affected.

Constraints: fix with layout (padding) only, no Vuetify internal overrides; leave the shared `EditorDialogHeader.vue` spacing unchanged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The floating Event name label in the New event editor is fully inside its scroll container (not clipped)
- [x] #2 A browser E2E regression check fails before the fix and passes after it, with evidence recorded
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add Chromium E2E spec `e2e/specs/new-event-name-label-layout.spec.ts`: open the New event dialog from `/`, compare the floating label's top with its `.v-card-text` scroll container's top.
2. Run it in isolation and record the failure.
3. Change NewEvent.vue card-text `tw:py-1` to `tw:pt-3 tw:pb-1`.
4. Rerun the identical check, then run the frontend required checks.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Bug Fix Protocol evidence (command: `VITE_ENABLE_COOKIE_CONSENT=false npm run test:e2e -- --project=chromium-desktop -g "keeps the floating Event name label"` from `e2e/` inside the nix dev shell):
- Fail before: `expect(received).toBeGreaterThanOrEqual(0)` received `-5`; the floating label top sat 5px above the `.v-card-text` scroll container and was clipped.
- Pass after: same command, 1 passed. Also passes on chromium-mobile; 6/6 with `--repeat-each=3` across both Chromium projects on a warm dev server.

Notes:
- Two repeat runs right after `npm run build` timed out on the `Create event` click with a blank page (cold Vite dev server, ~15s load). The existing `styling-migration` specs and warm reruns were stable, so this is environmental, not the spec.
- The local `.env.test` lacks `VITE_ENABLE_COOKIE_CONSENT` (present in `.env.test.example`), so it was passed on the command line.
- Frontend checks: lint, fmt:check, typecheck, and build pass. test:unit has 1 unrelated failure in `src/utils/timezoneDateRules.test.ts` ("uses the rendered week when deriving weekly schedule offsets": expected -120, got -60), which does not touch NewEvent.vue.
- Visually confirmed at desktop (1440) and mobile (390) widths from the recorded video frames.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The New event form body (`v-card-text`, the scroll container) had only 4px top padding, so `overflow: auto` clipped the outlined "Event name (required)" floating label, which sits across the field's top border. Changed `tw:py-1` to `tw:pt-3 tw:pb-1` in `NewEvent.vue`. Added the Chromium E2E regression `e2e/specs/new-event-name-label-layout.spec.ts`, which asserts the floating label stays inside its scroll container; it failed before the fix (-5px) and passes after. The sign-up and group editors share the layout but use label-less solo fields, so they are unaffected.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
