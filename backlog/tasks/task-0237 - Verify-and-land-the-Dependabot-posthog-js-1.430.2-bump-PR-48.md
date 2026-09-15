---
id: TASK-0237
title: 'Verify and land the Dependabot posthog-js 1.430.2 bump (PR #48)'
status: Done
assignee:
  - opencode
created_date: '2026-09-15 14:40'
updated_date: '2026-09-15 15:19'
labels:
  - frontend
dependencies: []
references:
  - 'https://github.com/time-tools/timeful/pull/48'
modified_files:
  - frontend/package.json
  - frontend/package-lock.json
  - frontend/src/plugins/posthog.ts
  - frontend/src/plugins/posthog.test.ts
  - frontend/src/plugins/posthog.contract.test.ts
priority: medium
type: chore
ordinal: 232000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Dependabot PR #48 (https://github.com/time-tools/timeful/pull/48) bumps posthog-js from 1.422.5 to 1.430.2 in frontend, updating frontend/package.json and frontend/package-lock.json. Transitive updates in the lockfile include @posthog/browser-common, @posthog/core, @posthog/types, and web-vitals 5.x to 6.x.

The frontend's only PostHog integration is frontend/src/plugins/posthog.ts, a lazy wrapper around the posthog-js default export that uses init, capture, identify, reset, and get_distinct_id; roughly 30 call sites consume the wrapper and there is no dedicated test for the plugin.

Outcome: the upgrade is verified safe against the used client surface and required frontend checks, then merged, or closed with a recorded evidence-backed reason to reject or defer it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 frontend/package.json and frontend/package-lock.json consistently resolve posthog-js 1.430.2 and the lockfile passes npm ci on the PR branch
- [x] #2 Required frontend checks pass on the PR branch: lint, fmt:check, typecheck, build, test:unit
- [x] #3 The 1.422.5 to 1.430.2 upstream changes are reviewed for effects on the used client surface (init, capture, identify, reset, get_distinct_id) and any required adaptation is implemented and tested
- [x] #4 Runtime evidence confirms the upgraded package loads through the plugin's dynamic import and exposes the used methods
- [x] #5 PR #48 CI is green and the PR is either merged or closed with a recorded rationale
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
Research findings (2026-09-15, branch dependabot/npm_and_yarn/frontend/posthog-js-1.429.5 at 8d211116):
- PR #48 CI: frontend-quality fails on typecheck; browser-e2e (chromium) fails because its production-assets setup runs npm run build, which fails on the same errors. firefox-desktop and firefox-touch pass.
- Root cause: @posthog/types moved from 1.407.1 to 1.411.1; identify(new_distinct_id?: string, ...) became identify(new_distinct_id: string, ...) (verified against the 1.407.1 tarball). The internal User type inherits optional `_id?: string` from generated transport types, so the five `posthog.identify(user._id, {...})` call sites fail TS2345.
- The only PostHog integration is frontend/src/plugins/posthog.ts (lazy dynamic import, init/capture/identify/reset/get_distinct_id). No plugin tests exist; unit tests mock the plugin. vitest default env is node; happy-dom is available.
- Real posthog-js 1.430.2 imports cleanly under node and exposes all five used methods (scratch check).

Plan:
1. Adapt the plugin boundary: widen the wrapper's identify first parameter to `string | undefined` and no-op on a missing distinct id, forwarding valid string ids unchanged. Keep the adaptation at the plugin boundary instead of adding five call-site guards.
2. Add focused plugin coverage in frontend/src/plugins/: wrapper tests with a mocked posthog-js (init options, capture/identify/reset forwarding, missing-id no-op) and a contract test importing the real package that asserts the used methods exist (runtime evidence for AC #4).
3. Run npm ci (done) and required checks: lint, fmt:check, typecheck, build, test:unit.
4. Commit to the Dependabot branch, push, and confirm frontend-quality and browser-e2e (chromium) go green on PR #48.
5. Report results, update the task with evidence, and ask the user before merging.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-15: Root cause of both PR #48 CI failures found: @posthog/types 1.407.1 -> 1.411.1 changed identify(new_distinct_id?: string, ...) to identify(new_distinct_id: string, ...); User._id is optional, so five identify call sites failed typecheck, and browser-e2e (chromium) failed only because its production-assets setup runs npm run build.

Changes on the branch worktree:
- frontend/src/plugins/posthog.ts: identify now takes `string | undefined` and no-ops when the distinct id is missing; valid ids are forwarded unchanged with their properties. Adaptation stays at the plugin boundary instead of adding five call-site guards.
- frontend/src/plugins/posthog.test.ts: new happy-dom wrapper suite (mocked posthog-js) covering lazy init options, capture/identify forwarding, missing-id no-op, reset clearing the stored fallback id, and fallback/client distinct id behavior.
- frontend/src/plugins/posthog.contract.test.ts: new node-env test importing the real posthog-js 1.430.2 and asserting init/capture/identify/reset/get_distinct_id exist (runtime evidence).

Evidence:
- npm ci: clean install from the bumped lockfile.
- npm ls: posthog-js@1.430.2, @posthog/types@1.411.1, @posthog/core@1.53.2, @posthog/browser-common@0.8.3, web-vitals@6.2.1.
- npm run lint: 0 errors (2 pre-existing vue/one-component-per-file warnings in NewSignUp.test.ts).
- npm run fmt:check: all files formatted.
- npm run typecheck: passes (previously failed with 5 TS2345 errors).
- npm run build: passes; posthog chunk emitted (dist/assets/posthog-^).
- npm run test:unit: 148 files, 1105 tests passed (includes 8 new plugin tests).

Remaining: commit and push to the Dependabot branch, confirm the PR CI is green, then merge or close with rationale. Awaiting user confirmation before pushing.

2026-09-15: Pushed fix commit 552f29a1 to dependabot/npm_and_yarn/frontend/posthog-js-1.429.5. PR #48 CI is fully green: frontend-quality pass 1m19s, browser-e2e chromium pass 5m31s, firefox-desktop pass 6m23s, firefox-touch pass 4m22s, markdown-quality pass 24s. Repo merge settings: merge commits only (squash and rebase disabled), branch auto-delete on merge.

2026-09-15: PR #48 merged by the user at 14:59:39Z as merge commit b8547cb5; the Dependabot branch was auto-deleted. Local main fast-forwarded to b8547cb5 (posthog-js 1.430.2 and the plugin adaptation are on main) and graphify update . completed.

2026-09-15: Task was created as TASK-0235 but renumbered to TASK-0237 by `backlog doctor --fix --yes` because a concurrent main-branch session had already created TASK-0235 (vitest 5 typecheck fix, Done) before this task's file was merged. The six reference lines doctor reported are prose mentions in TASK-0235/TASK-0236 that refer to the vitest task and need no change.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged Dependabot PR #48 (merge commit b8547cb5), bumping frontend posthog-js from 1.422.5 to 1.430.2 with transitive @posthog/browser-common 0.8.3, @posthog/core 1.53.2, @posthog/types 1.411.1, and web-vitals 6.2.1.

Why code changed: @posthog/types 1.411.1 made identify's distinct id required (was optional in 1.407.1). Five call sites pass the optional User._id, which broke typecheck and the production build (and with it the chromium E2E production-assets setup). frontend/src/plugins/posthog.ts now accepts `string | undefined` for identify and skips the call when the distinct id is missing, keeping the compatibility adaptation at the plugin boundary rather than adding view-level guards.

Tests: added frontend/src/plugins/posthog.test.ts (wrapper suite with a mocked client: lazy init options, capture/identify forwarding, missing-id no-op, reset clearing the stored fallback id, fallback/client distinct id behavior) and frontend/src/plugins/posthog.contract.test.ts (imports real posthog-js and asserts init/capture/identify/reset/get_distinct_id exist). Local: npm ci, lint, fmt:check, typecheck, build, and test:unit (148 files, 1105 tests) all pass. CI: frontend-quality, browser-e2e (chromium, firefox-desktop, firefox-touch), and markdown-quality all pass on the merged head.

Risk: none expected for analytics; identify calls with a present id are unchanged. Follow-up observations (pre-existing, out of scope): loadPosthog calls import.meta.env.VITE_POSTHOG_API_KEY.trim() before the blank check, so a genuinely unset variable throws instead of disabling analytics; and frontend/Dockerfile declares only VITE_POSTHOG_API_KEY though compose passes VITE_POSTHOG_API_HOST.

Note: this task was created as TASK-0235 and renumbered to TASK-0237 by `backlog doctor` after a concurrent-session ID collision with the vitest 5 typecheck task (TASK-0235).
<!-- SECTION:FINAL_SUMMARY:END -->
