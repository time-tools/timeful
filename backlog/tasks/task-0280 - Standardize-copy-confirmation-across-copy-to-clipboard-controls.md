---
id: TASK-0280
title: Standardize copy confirmation across copy-to-clipboard controls
status: Done
assignee:
  - opencode
created_date: '2026-09-19 21:01'
updated_date: '2026-09-19 21:23'
labels: []
dependencies: []
priority: medium
type: enhancement
ordinal: 280000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Standardize how every copy-to-clipboard control signals the copy result.

Current behavior is inconsistent: the event page shows a success snackbar even when `writeText` rejects, while the Manage access dialog and Continue on this device swap the button label to "Copied" and never revert.

Required standard: await the clipboard write, only then show an in-place confirmation (check icon + "Copied") that automatically reverts after two seconds; restart the timer on repeat activation; clear it on unmount; report failures without a success signal and without removing the copyable text; announce the outcome through a polite live region. Use the snackbar only when the trigger disappears (event-list overflow menu) or for errors with no inline error slot.

Scope: event page Copy link, Manage access Copy link, Continue on this device Copy code, event-list menu Copy link, respondents-list email copy. Delete the unused footer contract-address copy helper. Record the durable behavior as FR-130.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All copy-to-clipboard controls await the clipboard write before signaling success
- [x] #2 A successful copy swaps the activating control to a check icon and "Copied" label, then automatically reverts after two seconds; repeated activation restarts the timer and unmount clears it
- [x] #3 The Manage access dialog and Continue on this device page no longer stay stuck on "Copied"
- [x] #4 A failed copy shows no success state, preserves the copyable text, and reports an error inline in dialogs and via error snackbar elsewhere
- [x] #5 The event-list overflow menu confirms a copy via snackbar because the menu closes on activation
- [x] #6 Every copy outcome is announced through a polite live region
- [x] #7 The unused footer contract-address copy helper is removed
- [x] #8 FR-130 and its requirements index row are added
- [x] #9 Unit tests cover success, timed revert, repeat activation, failure, and announcements for the affected controls
- [x] #10 A Firefox e2e spec with a stubbed clipboard verifies the transient copy state on the event page and in the Manage access dialog
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add `frontend/src/composables/useCopyFeedback.ts` with `copied`, `announcement`, and `copy(text)`: success-only state, two-second auto-revert timer restarted on repeat activation, cleared on scope dispose.

2. Author FR-130 in `docs/requirements/functional/fr/FR-130.md` and add its row to `docs/requirements/README.md` after FR-129.

3. Apply the composable to the five live copy controls: event page Copy link (via `useEventEditing` exposing `linkCopied`/announcement), Manage access Copy link, Continue on this device Copy code, event-list menu Copy link (snackbar fallback after an awaited write), respondents-list email copy (in-place icon swap with per-row copied tracking).

4. Delete the unused footer contract-address copy helper and any now-unused imports.

5. Add `e2e/specs/timed-event-copy-feedback-firefox.spec.ts` with a clipboard stub installed through `addInitScript`; assert the transient "Copied" state and automatic revert on the event page and in the Manage access dialog without any clipboard permission grants.

6. Update and extend unit tests with fake timers for revert, repeat activation, failure, and announcements.

7. Run frontend lint/fmt/typecheck/build/unit, the focused Firefox e2e spec, root Markdown formatting, and `graphify update .`.

8. (Regression follow-up) Decouple the Manage access step highlight from the transient copy confirmation: track a durable `linkCopied` ref for `activeStep` and reset it with a new transfer link.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-19 21:08
---
Registered the new requirement as FR-130, not FR-128: FR-128 (Access Transfer step flow) and FR-129 (Access Transfer reload restoration) already exist. An earlier draft accidentally overwrote FR-128; it has been restored from HEAD. The FR-130 index row was added after FR-129.
---

created: 2026-09-19 21:23
---
Regression found while reviewing the staged change: `activeStep` in `EventAccessTransfer.vue` read the transient `copied` ref, so when the two-second copy confirmation auto-reverted, the Manage access highlight fell back from step 2 to step 1. Fixed by tracking a durable `linkCopied` ref for step progression (set only after a successful write, reset when a replacement link is created) and leaving `copied` purely to the button confirmation. Added a fake-timer unit test asserting step 2 stays current after the revert, plus step-2 `aria-current` assertions in the Firefox copy-feedback e2e spec.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Standardized copy confirmation across all copy-to-clipboard controls.

Implementation:
- Added `frontend/src/composables/useCopyFeedback.ts`: awaits the clipboard write, sets success-only `copied` state, exposes a polite `announcement`, auto-reverts after two seconds, restarts the timer on repeat activation, and clears the timer on scope dispose or `reset()`.
- Manage access (`EventAccessTransfer.vue`) and Continue on this device (`AccessTransfer.vue`) now use the composable, so "Copied" reverts instead of sticking; both keep their inline error and existing live region.
- Event page (`Event.vue` / `useEventEditing.ts`) copy link now awaits the write, swaps to a check icon plus "Copied" with a live region, and reports failure via error snackbar instead of always showing success.
- Event-list menu (`EventItem.vue`) awaits the write; success uses the existing snackbar because the menu closes, and failure now shows an error snackbar.
- Respondents list email copy uses an in-place check icon plus live region with per-row tracking; failures keep the error snackbar.
- Removed the unused footer contract-address copy helper and its imports.
- Added FR-130 and its index row.

Evidence:
- `frontend`: lint (0 errors), fmt:check, typecheck, build, test:unit -> 151 files / 1182 tests passed.
- `e2e`: lint, typecheck, `--project=firefox-desktop specs/timed-event-copy-feedback-firefox.spec.ts` -> 2 passed (stubbed clipboard; no permission grants).
- Root `npm run format:markdown` and `graphify update .` completed.
- Swagger regeneration (/5) and root scripts/prettier formatting (/7) were not applicable: no server or scripts/prettier changes.

Note: FR-130, not FR-128: FR-128 and FR-129 already existed. An earlier draft briefly overwrote FR-128; it was restored from HEAD before finalizing.

Regression follow-up (from the staged review): decoupled the Manage access step highlight from the transient copy confirmation. `activeStep` now reads a durable `linkCopied` ref that is set only after a successful write and reset when a replacement link is created, while the button still reverts from "Copied" after two seconds. Added `keeps step 2 current after the copied confirmation reverts` to `EventAccessTransfer.test.ts` and step-2 `aria-current` assertions to `timed-event-copy-feedback-firefox.spec.ts`.

Updated evidence: `frontend` lint/fmt:check/typecheck/build pass and `test:unit` -> 151 files / 1183 tests passed; focused Firefox e2e `specs/timed-event-copy-feedback-firefox.spec.ts` -> 2 passed; `graphify update .` completed.
<!-- SECTION:FINAL_SUMMARY:END -->
