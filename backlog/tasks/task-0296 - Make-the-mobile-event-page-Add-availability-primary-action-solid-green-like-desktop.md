---
id: TASK-0296
title: >-
  Make the mobile event-page Add availability primary action solid green like
  desktop
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-20 22:08'
updated_date: '2026-09-20 22:16'
labels:
  - frontend
  - mobile
  - event-page
  - styling
dependencies: []
references:
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
  - e2e/specs/event-mobile-editing-options.spec.ts
  - docs/requirements/functional/fr/FR-076.md
documentation:
  - docs/requirements/README.md
  - frontend/AGENTS.md
priority: medium
type: bug
ordinal: 296000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On a phone viewport, the event page bottom action bar renders its primary Add availability action as a white elevated button with green text and a green glow. The desktop event header renders the same Add availability action as a solid green button with white text and a subtle neutral shadow. The same availability action therefore reads as a secondary control on phones and visually contradicts the desktop primary treatment that the product expects on both viewports.

The user report shows the mobile no-response event page: the bottom bar contains the blue outlined Schedule action on the left and the primary "+ Add availability" action on the right with a white fill, green border, and green glow shadow.

Confirmed product expectation: the mobile Add availability primary action must look like the desktop Add availability primary action when there are no responses: solid green fill and white text, without the green shadow treatment.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a phone viewport, when the event page bottom action bar shows Add availability as its primary availability action, the button renders with the solid primary-action green fill (#00994c) and a white label and icon, matching the desktop Add availability button treatment.
- [x] #2 The mobile Add availability primary action no longer uses the white elevated treatment: no white background, no green label, and no green glow box-shadow.
- [x] #3 The mobile Edit availability state, the secondary add-guest availability action, the mobile Schedule action, and the desktop Add availability button are visually unchanged.
- [x] #4 Frontend unit coverage asserts the mobile Add availability primary action renders with the solid green fill and white text classes and without the white elevated-button treatment.
- [x] #5 Browser coverage on the mobile project asserts the rendered computed background color and box-shadow of the mobile Add availability primary action on a no-response timed event page.
- [x] #6 The relevant requirements or documentation records are updated if this styling is captured by a functional requirement.
- [x] #7 npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit pass in frontend/.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Reproduce in the browser with a new chromium-mobile Playwright test that opens a no-response timed event and asserts the computed background, text color, and box-shadow of #mobile-primary-availability-btn; run it to capture the current white fill and green glow.

2. In frontend/src/views/Event.vue, give the mobile primary availability button the same solid primary treatment as the desktop Add availability button: static tw:bg-green tw:text-white fill, a mobile-primary-availability-button--add variant for the Add availability text, and a --add rule with the desktop border and neutral shadow. Keep the --edit rules and disabled-outline rule intact.

3. In frontend/src/views/Event.test.ts, update the mobile add availability class assertions to the solid green/white expectations and add a style-rule test that the desktop and mobile add variants share the neutral shadow and contain no green glow.

4. Keep the computed-style e2e assertion as browser regression coverage for AC5.

5. Run the required frontend checks, e2e lint/typecheck, and the affected chromium-mobile and chromium-desktop specs, then run graphify update.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reproduced on chromium-mobile before the fix: the e2e assertion failed with computed backgroundColor rgb(255, 255, 255) for #mobile-primary-availability-btn. The button carried timeful-elevated-button tw:bg-white tw:text-green, so index.css applied the green glow box-shadow 0px 2px 8px #00994c80; the failure screenshot and run directory are /tmp/opencode/timeful-e2e-artifacts/2026-09-20T22-09-18.422158936Z-p2306064/. Desktop reference: #desktop-primary-availability-btn is tw:bg-green tw:text-white with border 1px solid var(--timeful-primary-action-bg) and box-shadow 0px 2px 6px rgba(0, 0, 0, 0.14).

Fix mirrors the desktop add treatment: the mobile primary availability button now carries the static tw:bg-green tw:text-white fill, mobilePrimaryAvailabilityButtonClass emits mobile-primary-availability-button--add for the Add availability text and keeps --edit, and the new --add rule adds the same primary-action border and neutral 0px 2px 6px rgba(0, 0, 0, 0.14) shadow with no #00994c80 glow.

Scope note: the same button is also the mobile sign-up primary (Edit slots). It now renders solid green without the white elevated treatment, which matches the desktop primary, and no unit or e2e coverage asserted the old white treatment for that state.

Requirements: FR-076 records the white elevated mobile action bar surface and states that availability colors remain reserved for availability actions; it does not specify the primary button fill, and it stays accurate, so no requirement change was needed.

Checks: frontend lint, fmt:check, typecheck, build, and test:unit pass (154 files, 1234 tests); e2e lint and typecheck pass; the new chromium-mobile computed-style test passes; the affected chromium-mobile and chromium-desktop specs pass with 8 passed and 8 expected skips; graphify update . completed. e2e fmt:check still flags only the pre-existing unmodified specs/timed-event-access-transfer-firefox.spec.ts.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The mobile event-page primary `Add availability` action in the bottom action bar now uses the same solid primary treatment as the desktop `Add availability` button instead of the white elevated button with green text and a green glow.

- frontend/src/views/Event.vue: the mobile primary availability button now carries the static `tw:bg-green tw:text-white` fill, `mobilePrimaryAvailabilityButtonClass` emits `mobile-primary-availability-button--add` for the Add availability text alongside the existing `--edit` variant, and the new `.mobile-primary-availability-button--add` rule mirrors the desktop add button with `border: 1px solid var(--timeful-primary-action-bg)` and `box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.14)`. The `timeful-elevated-button tw:bg-white tw:text-green` classes and the resulting `#00994c80` glow are gone. The Edit availability, secondary add-guest, and Schedule actions are untouched.
- frontend/src/views/Event.test.ts: the mobile add-availability test now asserts the `--add` variant with `tw:bg-green`/`tw:text-white` and the absence of `timeful-elevated-button`, `tw:bg-white`, and `tw:text-green`; a new style-rule test asserts the desktop and mobile add variants share the neutral shadow and contain no green glow.
- e2e/specs/event-mobile-editing-options.spec.ts: a new chromium-mobile test opens a no-response timed event and asserts the rendered computed `backgroundColor` `rgb(0, 153, 76)`, `color` `rgb(255, 255, 255)`, and `boxShadow` `rgba(0, 0, 0, 0.14) 0px 2px 6px 0px`.

## Reproduction

Before the fix, the new e2e test failed with `Received: "rgb(255, 255, 255)"` and the failure screenshot showed the white button with the green glow, matching the report. Artifact run: `/tmp/opencode/timeful-e2e-artifacts/2026-09-20T22-09-18.422158936Z-p2306064/`.

## Scope notes

- The same button is the mobile sign-up primary (`Edit slots`); it now renders solid green without the white elevated treatment, matching the desktop primary. No unit or e2e coverage asserted the old treatment for that state.
- FR-076 records the white elevated action-bar surface and that availability colors stay reserved for availability actions; it does not specify the primary button fill and remains accurate, so no requirement change was needed.

## Verification

- frontend: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` (154 files, 1234 tests) all pass.
- e2e: `npm run lint` and `npm run typecheck` pass; the focused chromium-mobile test passes; the affected `event-mobile-editing-options.spec.ts` and `event-page-no-responses-layout.spec.ts` runs pass with 8 passed and 8 expected skips. e2e `fmt:check` still flags only the pre-existing, unmodified `specs/timed-event-access-transfer-firefox.spec.ts`.
- Root `format:markdown:check` passes; `graphify update .` completed (pre-existing zero-node warnings only). Swagger, scripts/prettier, and contract documents are unaffected.
<!-- SECTION:FINAL_SUMMARY:END -->
