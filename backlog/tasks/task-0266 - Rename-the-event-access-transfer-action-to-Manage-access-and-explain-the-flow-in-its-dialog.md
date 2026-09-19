---
id: TASK-0266
title: >-
  Rename the event access transfer action to Manage access and explain the flow
  in its dialog
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 13:50'
updated_date: '2026-09-19 13:54'
labels:
  - frontend
dependencies: []
documentation:
  - docs/design/architecture/adr/ADR-010.md
priority: medium
type: enhancement
ordinal: 266000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The event header action `Continue on another device` only describes creating a transfer link, but the dialog it opens also lists, approves, cancels, and revokes access transfers. The mismatch makes users expect device or session management that the label does not name. The agreed direction is the user-facing label `Manage access` for both the header action and the dialog title, with the dialog's opening paragraph explaining what the action is about: using the event on another browser and revoking access granted earlier, the create-link and matching-code approval steps with the five-minute window, and that opening the link alone gives no access. `browser` is the canonical term in FR-081, FR-082, FR-083, and the glossary Access Transfer definition, so the new copy should use it and keep the target page's `Continue on this device` wording.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The event page header action and the dialog it opens are labeled `Manage access`
- [x] #2 The dialog's opening paragraph explains that the action uses the event on another browser or revokes access granted earlier, describes creating a transfer link and approving the matching code within five minutes, and states that opening the link alone gives no access
- [x] #3 The signed-in and anonymous outcome lines remain accurate, and the anonymous line no longer repeats the revoke instruction already covered by the opening paragraph
- [x] #4 Unit tests use and assert the new label, and at least one unit assertion covers the explanatory paragraph's key content
- [x] #5 The Firefox access-transfer E2E spec uses the new label and passes
- [x] #6 ADR-010 names the new user-facing action and its updated date reflects the change; requirements and glossary remain accurate without edits
- [x] #7 Frontend lint, formatting, typecheck, build, and unit suites pass
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
1. In `frontend/src/components/event/EventAccessTransfer.vue`, set the trigger button label and dialog title to `Manage access`; replace the opening paragraph with: "Use this event on another browser, or revoke access you granted earlier. Create a transfer link, open it on the other browser, and approve the matching code shown there within five minutes. Opening the link alone gives no access."; drop the now-duplicated "You can revoke granted access here." from the anonymous outcome line.
2. Update `frontend/src/components/event/EventAccessTransfer.test.ts`: replace the nine `Continue on another device` selectors and add one assertion that the dialog text contains the explanatory content.
3. Update `frontend/src/views/Event.test.ts`: rename the header-action test titles and update the expected label arrays to `Manage access`.
4. Update `e2e/specs/timed-event-access-transfer-firefox.spec.ts` button locators.
5. Update `docs/design/architecture/adr/ADR-010.md` line 42 to name the `Manage access` action and bump `updated_date` to 2026-09-19; the target page keeps `Continue on this device`.
6. Run frontend lint, fmt:check, typecheck, build, and test:unit; run the Firefox `timed-event-access-transfer-firefox.spec.ts` spec; run `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Renamed the label and rewrote the dialog intro per the approved naming discussion: `Manage access` avoids promising device or session management while covering the dialog's create, approve, cancel, and revoke actions.

The anonymous outcome line's trailing "You can revoke granted access here." was dropped because the new opening paragraph already names revocation, avoiding duplication.

Reviewers should treat `Continue on this device` on the target page (`frontend/src/views/AccessTransfer.vue`) as intentionally unchanged: it is the action performed in the browser that opens the transfer link.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Renamed the event header action and its dialog from `Continue on another device` to `Manage access`, and made the dialog's opening paragraph explain what the action does.

## Change
- `frontend/src/components/event/EventAccessTransfer.vue`: trigger label and dialog title are `Manage access`; the opening paragraph explains using the event on another browser or revoking access granted earlier, creating a transfer link, approving the matching code within five minutes, and that opening the link alone gives no access; the anonymous outcome line dropped its now-duplicated revoke sentence. No style or behavior changes.
- `frontend/src/components/event/EventAccessTransfer.test.ts`: nine label selectors updated plus a new assertion covering the explanatory paragraph.
- `frontend/src/views/Event.test.ts`: header-action order and green-treatment tests updated to the new label.
- `e2e/specs/timed-event-access-transfer-firefox.spec.ts`: two button locators updated.
- `docs/design/architecture/adr/ADR-010.md`: the decision names the `Manage access` action and `updated_date` is 2026-09-19. Requirements and glossary already used `browser` and needed no edits. The target page keeps `Continue on this device`.

## Tests
- Focused suites: 77 passed. Full frontend suite: 149 files / 1124 tests passed.
- Firefox `timed-event-access-transfer-firefox.spec.ts`: 8/8 passed with `E2E_FRONTEND=bundled`.
- Frontend lint (0 errors, 2 pre-existing warnings), fmt:check, typecheck, and build passed; e2e lint and typecheck passed; `npm run format:markdown` and `graphify update .` run.

No API, contract, or behavior changes; swagger and `scripts/` were untouched.
<!-- SECTION:FINAL_SUMMARY:END -->
