---
id: TASK-0323
title: Add a version-gated check for the GALA translation probe corpus
status: To Do
assignee: []
created_date: '2026-09-23 16:36'
labels: []
dependencies: []
references:
  - server/scripts/20260923_gala_translation_probes/run.sh
  - docs/gala-translation.md
  - docs/ci.md
  - .github/workflows/backend-ci.yml
priority: medium
type: task
ordinal: 321000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0322 landed the Go-to-GALA probe corpus at `server/scripts/20260923_gala_translation_probes/`, but nothing runs it automatically, so a GALA or Go toolchain bump can leave the version matrix in `docs/gala-translation.md` stale without a signal.

The runner already pins `gala` 0.81.0 and the Go 1.26 series and fails on expectation drift; `GALA_PROBES_ALLOW_ANY_VERSION=1` and `GALA_PROBES_ALLOW_ANY_GO=1` exist only for deliberate re-baselining.

Outcome: the corpus becomes a checked artifact, run by CI when the pinned GALA CLI can be provisioned there, or otherwise covered by a documented scheduled or manual check with an explicit trigger and command.

Approach:
1. Decide whether the pinned GALA 0.81.0 CLI can be installed or built on a GitHub-hosted runner without unpinned downloads; record the decision in `docs/ci.md`.
2. If CI can run it, add or extend a workflow following the action pinning rules in `docs/ci.md`, and keep the corpus out of the Backend CI Compose stack unless the toolchain fits there.
3. If CI cannot run it, document the manual or scheduled check in `docs/ci.md` and point at it from `docs/gala-translation.md` "When to revisit".
4. Verify the chosen path catches drift before finishing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `docs/ci.md` records whether the corpus runs in CI, and when it does not, where the documented manual or scheduled check lives.
- [ ] #2 An automated job, when added, installs the pinned GALA 0.81.0 compiler and a Go 1.26 toolchain, runs `server/scripts/20260923_gala_translation_probes/run.sh`, and fails on any probe failure.
- [ ] #3 The automated path does not set `GALA_PROBES_ALLOW_ANY_VERSION` or `GALA_PROBES_ALLOW_ANY_GO`, so a compiler or toolchain bump fails visibly instead of being silently accepted.
- [ ] #4 The chosen check is verified to catch drift: an intentionally broken expectation is observed failing before the change is reverted.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
