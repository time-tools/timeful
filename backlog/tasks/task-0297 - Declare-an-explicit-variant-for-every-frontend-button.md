---
id: TASK-0297
title: Declare an explicit variant for every frontend button
status: To Do
assignee: []
created_date: '2026-09-20 22:39'
labels:
  - frontend
  - styling
  - buttons
dependencies: []
references:
  - frontend/src/views/Event.vue
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/BottomFab.vue
  - frontend/src/index.css
  - e2e/specs/event-mobile-editing-options.spec.ts
documentation:
  - frontend/AGENTS.md
  - docs/requirements/functional/fr/FR-112.md
  - docs/requirements/migration/backlog-fr-inventory-candidates/CAND-201.md
priority: medium
type: enhancement
ordinal: 297000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Vuetify renders `v-btn` as `variant="elevated"` by default, so most frontend buttons never declared their elevation.
An audit of `frontend/src` found 192 `<v-btn>` elements, 97 without an explicit variant: 70 inherit Vuetify's default elevated shadow, 26 rely on the `timeful-elevated-button` custom shadow utility, and 1 uses `timeful-flat-button`.
Two buttons in Event.vue bind `variant` dynamically with an `undefined` fallback (lines 227 and 236), which silently re-enables the default, and Event.vue:762 still uses the legacy boolean `flat` prop.
Some buttons additionally stack custom subtle shadows or glows, such as the `--add` rules on the Add availability buttons and the inline green glow on the home `+ Create new`.
Whether a button is flat or elevated is therefore accidental, and changing a button's context can silently add or keep a shadow.
The outcome of this task is that every `<v-btn>` declares its variant explicitly and no button receives a shadow implicitly.

The heuristic discussed for the first pass is only a starting point: filled or colored buttons flat; neutral surface buttons keep a shadow when they need separation from a same-color surface, for example `Create new transfer link`, `Copy link`, and `Cancel transfer` in the Manage access dialog; floating controls elevated, for example `BottomFab`, the Friends CTA, and `CreateSpeedDial`.
The heuristic classification must be reevaluated per button against the button's role, surface, state, and viewport during execution, and every divergence from the heuristic must be recorded so reviewers can see the judgment applied.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every `<v-btn>` under `frontend/src` declares an explicit `variant`; none falls back to Vuetify's default `elevated`, including dynamic bindings whose fallback is `undefined`.
- [ ] #2 Legacy boolean variant props on `<v-btn>` are replaced with the explicit `variant` equivalent.
- [ ] #3 Every button's variant is reevaluated individually against the heuristic, and the final assignment plus each divergence from the heuristic is recorded for review.
- [ ] #4 The event-page Add availability primary action is explicitly flat with no box-shadow on desktop and mobile, its `--add` shadow rules are removed, and Edit availability stays explicitly flat with no shadow.
- [ ] #5 Buttons that keep an elevation shadow do so through an explicit elevated variant or an explicitly named elevation class, and no flat button regains a shadow via custom CSS, utilities, or inline styles.
- [ ] #6 Button colors, labels, sizes, disabled and loading states, and interactions stay unchanged; only elevation treatment changes.
- [ ] #7 A source-level unit test enforces that every `<v-btn>` declares a variant and that no legacy boolean variant props remain, and browser coverage asserts computed `box-shadow` for representative flat and elevated buttons on mobile and desktop.
- [ ] #8 npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit pass in frontend/, and the affected e2e specs pass.
- [ ] #9 If the reevaluation settles a durable styling rule, it is recorded in frontend/AGENTS.md.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `graphify update .`
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
