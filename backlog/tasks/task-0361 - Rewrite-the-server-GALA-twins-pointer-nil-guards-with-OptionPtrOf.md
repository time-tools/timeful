---
id: TASK-0361
title: Rewrite the server GALA twins' pointer nil guards with OptionPtrOf
status: To Do
assignee: []
created_date: '2026-10-21 20:21'
labels:
  - gala
  - tooling
dependencies: []
references:
  - server/GALA.md
  - server/routes/respondent_identity.gala
  - server/routes/users.gala
  - server/routes/guest_response_ownership.gala
  - server/services/services.gala
  - server/discord_bot/init.gala
  - 'https://github.com/martianoff/gala/releases/tag/0.92.0'
priority: medium
type: chore
ordinal: 365300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0354's runtime idiom pass kept every Go-style pointer nil guard unchanged because the adopted runtime then had no pointer-to-Option conversion: `go_interop.OptionFromMap` was the only absence-to-Option helper, so rewriting a guard would only have moved the same `== nil` test into a `Some`/`None` constructor.

GALA 0.92.0 adds `std.OptionPtrOf[T](p *T) OptionPtr[T]`, which is empty when `p` is nil, plus the sealed `OptionPtr` views that match as `Some(v)`/`None()`. A scratch transpile on the pinned rev `4bd2d4e8` (2026-10-21) confirmed `std.OptionPtrOf(user) match { case Some(_) => ...; case None() => ... }` emits `std.OptionPtrOf(user).MatchView()` and builds, so the recorded rationale no longer holds.

Re-derive each pointer nil guard the idiom pass left in place and rewrite it with `OptionPtrOf` where the Go-facing signature, the wire shape, and every observable behavior stay unchanged. Keep a guard handwritten where the conversion would change the surface.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every pointer nil guard listed in server/GALA.md's runtime idiom pass is re-derived on rev 4bd2d4e8 and either rewritten with std.OptionPtrOf or recorded with the reason it stays Go-spelled.
- [ ] #2 Every rewritten member keeps its Go-facing signature, its wire shape, its HTTP status, and its observable behavior; no test is relaxed.
- [ ] #3 The generated Go imports only go.gala.fyi/stdlib, passes gofmt and go build ./..., and server/scripts/gala/verify.sh reports OK.
- [ ] #4 server/GALA.md's runtime idiom section records the pass, the remaining Go-spelled sites, and the new reason on the pinned rev.
- [ ] #5 The canonical Compose backend test sequence passes.
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
