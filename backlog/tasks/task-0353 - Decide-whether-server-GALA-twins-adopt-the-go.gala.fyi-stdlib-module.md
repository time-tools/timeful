---
id: TASK-0353
title: Decide whether server GALA twins adopt the go.gala.fyi/stdlib module
status: To Do
assignee: []
created_date: '2026-10-10 10:16'
updated_date: '2026-10-10 10:17'
labels: []
dependencies: []
references:
  - server/GALA.md
  - server/GALA_COMPILER
  - server/scripts/gala/verify.sh
  - .agents/skills/gala-from-go/SKILL.md
  - 'https://github.com/martianoff/gala/issues/740'
  - 'https://github.com/martianoff/gala/pull/743'
  - >-
    backlog/tasks/task-0349 -
    Advance-the-GALA-translation-loop-cursor-and-land-its-runtime-free-twins.md
documentation:
  - server/GALA.md
  - server/README.md
priority: medium
ordinal: 356005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Context

Every committed twin under `server/` is runtime-free by the contract in `server/GALA.md`: its generated Go names no `martianoff/gala/...` or `go.gala.fyi/stdlib/...` package, `server/go.mod` carries no GALA runtime require, and `server/scripts/gala/verify.sh` rejects both import paths. That contract is why twins bind raw Go values with `var`, keep map and slice literals, multi-value results, and `Hash`/`Compare` in handwritten `.go` siblings, and avoid `val`, `Try`, GALA structs, and sealed types.

The pinned compiler (GALA 0.87.1 post-release HEAD, flake rev `cd2fdcb5`) adds `gala transpile --stdlib-module go.gala.fyi/stdlib` (upstream PR #743, closing this repository's issue #740). A scratch transpile on 2026-10-10 rewrote an emitted `martianoff/gala/std` import to `go.gala.fyi/stdlib/std`, and the published module is a plain Go module at v0.87.1, so the earlier objection that only `gala export` could rewrite imports and that a `replace` was refused by Go no longer applies to `gala transpile` output.

The decision was surfaced by TASK-0349's iteration 10 compiler bump and is deliberately deferred there: the runtime-free contract is unchanged until this task decides otherwise.

## Decision required

Keep the runtime-free contract, or adopt the published module for server twins. If adopting, the decision must also fix the scope (all existing twins, new twins only, or named candidates such as the still-blocked `models/` group or `postgres/`) and the migration and documentation work it triggers.

Weigh at least: the server build and test paths (`server/Dockerfile`, Compose stack, `verify.sh` regeneration, offline Go caches), the GALA-runtime dependency policy TASK-0347 established when it removed the vendored stdlib, and the runtime-free guidance in the `gala-from-go` and `gala-loop` skills.

This task decides and, if it changes course, creates the follow-up tasks; it does not migrate twins itself.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A reproducible scratch check on the pinned compiler (flake rev cd2fdcb5, `gala transpile --stdlib-module go.gala.fyi/stdlib`) shows generated Go building and running in a plain Go module that requires the published stdlib; the task notes record the module version, its transitive requirements, and the exact commands.
- [ ] #2 The keep-or-adopt decision and, if adopting, its scope are recorded in `server/GALA.md`, with the runtime-free contract wording either confirmed with the new rationale or replaced by the adopted dependency policy.
- [ ] #3 If the decision is keep: `server/GALA.md` states why the capability stays unused, and no runtime, verification, or skill behavior changes.
- [ ] #4 If the decision is adopt: follow-up Backlog tasks exist for the go.mod/verify.sh contract change, the twin migration scope, and the gala-from-go and gala-loop updates; no twin is migrated in this task.
- [ ] #5 Changed Markdown files are formatted with `npm run format:markdown`.
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
