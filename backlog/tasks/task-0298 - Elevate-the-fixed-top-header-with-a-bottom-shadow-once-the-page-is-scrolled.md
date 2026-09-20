---
id: TASK-0298
title: Elevate the fixed top header with a bottom shadow once the page is scrolled
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-20 22:45'
updated_date: '2026-09-20 23:00'
labels: []
dependencies: []
references:
  - frontend/src/App.vue
  - frontend/src/index.css
  - frontend/src/App.test.ts
  - e2e/specs/landing-hero.spec.ts
modified_files:
  - frontend/src/App.vue
  - frontend/src/index.css
  - frontend/src/App.test.ts
  - e2e/specs/landing-hero.spec.ts
priority: medium
type: enhancement
ordinal: 298000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The fixed top header is a flat white bar, so once page content scrolls underneath it the header blends into the content passing behind it and the layering is no longer readable (for example the event title scrolling under the header). The header should read as elevated only while content is actually behind it: flat at the top of the page, visibly separated from the content below once the page is scrolled.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On routes that show the header, the header has no bottom shadow while the page is scrolled to the top.
- [x] #2 Scrolling the page down gives the header a visible bottom shadow that separates it from content scrolling underneath.
- [x] #3 Returning to the top of the page removes the bottom shadow again.
- [x] #4 The shadow state applies to both desktop and phone header layouts and does not cover or block header controls.
- [x] #5 Unit coverage asserts the header enters the elevated state on scroll and returns to the flat state at the top.
- [x] #6 Browser e2e coverage asserts the header has no shadow at the top of a scrollable page and gains one after scrolling.
- [x] #7 The scrolled header shadow is limited to the centered header content column rather than spanning the full viewport width.
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
1. Track window scroll state in App.vue: turn the existing no-op `handleScroll` lifecycle hook into a `window.scrollY > 0` tracker that feeds a scrolled flag, and bind a header elevation class to the fixed `[data-testid="app-header"]` container.
2. Define the header elevation surface in `frontend/src/index.css` as an owned class backed by a `--timeful-*` shadow token, consistent with existing elevation tokens, instead of component-local raw palette values or a bare Tailwind shadow.
3. Seed the initial state on mount (a restored scroll position may exist before the first scroll event) and keep the listener add/remove lifecycle intact.
4. Add unit regression coverage in `frontend/src/App.test.ts` asserting the header is flat at the top, elevated after a scroll event, and flat again after returning to the top.
5. Add e2e regression coverage in `e2e/specs/landing-hero.spec.ts` asserting no shadow at the top and a shadow after scrolling the landing page, then returning flat at the top.
6. Run required frontend checks (`lint`, `fmt:check`, `typecheck`, `build`, `test:unit`) and the focused Firefox desktop e2e test; run `graphify update .`.

Refinement after review: the elevation class is bound to the centered `max-w-5xl` header content container (`[data-testid="app-header-content"]`) rather than the full-width fixed bar, so the shadow width is limited to the header content column (full width on phone, capped at 1024px and centered on wider viewports). Unit and e2e coverage assert the shadow on that content container.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented scroll-driven header elevation: `App.vue` tracks `window.scrollY > 0` in the existing `handleScroll` listener (seeded once after listener registration), and `index.css` defines `.timeful-elevated-header` with the new `--timeful-header-shadow` token. The class is bound to the centered `max-w-5xl` header content container (`[data-testid="app-header-content"]`) so the shadow width is limited to the content column.

Unit coverage added in `App.test.ts` (`elevates the fixed header on scroll and flattens it again at the top`); scoped run passed 10/10 and the full suite passed 154 files / 1235 tests.

Browser coverage added in `e2e/specs/landing-hero.spec.ts`; chromium-desktop and chromium-mobile runs passed (4 passed, 2 project skips). Firefox projects restrict `testMatch` to their own specs, so this spec is Chromium-only by config.

E2E diagnosis: the landing page is not scrollable on first paint because the hero image/lazy content grows the layout after load. The test sets a 700px viewport and waits for `scrollHeight > innerHeight` with `expect.poll` before scrolling; this replaced an earlier `window.scrollTo(0, 400)` that silently no-oped.

Visual check after scrolling at 1440x700 saved to `/tmp/opencode/task-0298-scrolled-header.png`; the shadow paints only under the centered content column.

Checks run: frontend lint (0 errors; 2 pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, build, full test:unit; e2e package lint and typecheck; root `format:markdown:check`; `graphify update .`.

Pre-existing and unrelated: the e2e package `npm run fmt:check` flags unmodified `specs/timed-event-access-transfer-firefox.spec.ts`.

Temporary diagnostic/screenshot specs used during this task were moved out of `e2e/specs/` to `/tmp/opencode/` and are not part of the change.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The fixed top header now reads as elevated while the page is scrolled: it stays flat at the top of the page and shows a bottom shadow once content is behind it. Per review feedback, the shadow is limited to the centered `max-w-5xl` header content column (full width on phone, centered and capped at 1024px on wider viewports) rather than spanning the full-viewport white bar.

## Changes

- `frontend/src/App.vue`: the existing window `scroll` listener hook now drives an `isScrolled` flag from `window.scrollY > 0`, seeded once after listener registration so a restored scroll position is reflected before the first scroll event. The `timeful-elevated-header` class binds to the centered header content container, which gained `data-testid="app-header-content"` for stable targeting.
- `frontend/src/index.css`: added the `--timeful-header-shadow` semantic token (`rgba(0, 0, 0, 0.18)`) and the owned `.timeful-elevated-header` class (`box-shadow: 0 2px 8px`).
- `frontend/src/App.test.ts`: added coverage that the header content is flat at the top, elevated after a `scroll` event with non-zero `scrollY`, and flat again after returning to the top.
- `e2e/specs/landing-hero.spec.ts`: added a browser check that the header content has no shadow at the top, gains the elevated class and a painted shadow after scrolling, and returns to flat at the top. The test also asserts the shadowed container is centered and at most 1024px wide.

## Verification

- `npm run lint` (0 errors; 2 pre-existing warnings in `NewSignUp.test.ts`)
- `npm run fmt:check`
- `npm run typecheck`
- `npm run build`
- `npm run test:unit` — 154 files, 1235 tests passed
- `npm run test:e2e -- --project=chromium-desktop --project=chromium-mobile specs/landing-hero.spec.ts` — 4 passed, 2 project-specific skips, 0 failed
- `e2e` package `npm run lint` and `npm run typecheck` passed
- `npm run format:markdown:check` passed; `graphify update .` ran
- Visual check after scrolling at 1440x700: `/tmp/opencode/task-0298-scrolled-header.png`

## Notes

- `landing-hero.spec.ts` is Chromium-only by project config (the Firefox projects restrict `testMatch` to their own specs), so the browser check ran in `chromium-desktop` and `chromium-mobile`.
- The landing page grows after first paint as hero content loads, so the e2e check waits for `scrollHeight > innerHeight` before scrolling; a 700px viewport keeps a comfortable scroll range for both project widths.
- Pre-existing, unrelated: the `e2e` package `npm run fmt:check` flags unmodified `specs/timed-event-access-transfer-firefox.spec.ts`. No Markdown, Swagger, `scripts/`, `prettier/`, or contract documents changed.
<!-- SECTION:FINAL_SUMMARY:END -->
