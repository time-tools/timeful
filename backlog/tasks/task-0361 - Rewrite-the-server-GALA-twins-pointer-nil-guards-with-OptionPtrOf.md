---
id: TASK-0361
title: >-
  Reduce raw pointer usage in the server GALA twins with Ptr and
  OptionPtr
status: In Progress
assignee:
  - '@dsh'
created_date: '2026-10-21 20:21'
updated_date: '2026-10-26 12:28'
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
The committed server GALA twins still spell most non-Go-facing values the Go way: raw nullable `*T` pointers and `== nil` guards on pointers that GALA itself never lets be nil.
GALA 0.93.0 provides the pointer-safety pair this pass needs: `Ptr[T]` is a compile-time non-null pointer that erases to Go's `*T`, `OptionPtr[T]` is its pointer-sized absent-or-present counterpart, `OptionPtrOf(p)` converts a nullable `*T` from Go once at the boundary, and `RequirePtr(p)` does the same when nil is a bug.
The pass must reduce the raw `*T` surface without changing a Go-facing signature, a wire shape, an HTTP status, or an observable behavior.

Work items:
- Convert raw `*T` to `Ptr[T]` where the pointer is non-null and does not cross the Go boundary, and to `OptionPtr[T]`/`OptionPtrOf` where nil means absence and the value stays inside GALA code.
- Keep `*T` exactly where it is the Go boundary spelling: a Go call's nullable return, a Go-facing exported signature whose nil behavior must not change, a value that must be assigned nil or a Go pointer later, or an address target.
- Re-derive TASK-0361's original pointer nil guards on the current pinned compiler and rewrite them with `OptionPtrOf` where the rewrite is behavior-preserving.
- Record every kept raw-pointer site and its reason, rather than silently leaving it.

Guardrails: the pinned compiler is GALA 0.93.0 at flake rev 8ba02b2b.
`populateSignUpResponsePayloadIdentity` is called by `routes/event_routes.go`, `CallApi` by `services/*`, and route tests call `canonicalGuestName`, `guestNameValidationErrorMessage`, and `shouldExposeGuestSignUpResponsePayload`, so those signatures do not change.
No generated file is hand-edited; every twin is regenerated with `verify.sh --write` and checked byte-identically.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every raw `*T` in the twins is either converted to `Ptr[T]` where the pointer is non-null and internal to GALA, converted to `OptionPtr[T]`/`OptionPtrOf` where nil means absence, or kept as `*T` at a Go boundary with its reason recorded.
- [ ] #2 Every pointer nil guard listed in `server/GALA.md`'s runtime idiom pass is re-derived on rev `8ba02b2b` and either rewritten with `std.OptionPtrOf` or recorded with the reason it stays Go-spelled.
- [ ] #3 Every rewritten member keeps its Go-facing signature, wire shape, HTTP status, and observable behavior; no test is relaxed and no generated file is hand-edited.
- [ ] #4 The generated Go imports only `go.gala.fyi/stdlib`, passes `gofmt` and `go build ./...`/`go vet ./...`, and `server/scripts/gala/verify.sh` reports OK with byte-identical regeneration.
- [ ] #5 `server/GALA.md` records the pass, the before/after raw-pointer counts, the kept sites with reasons, and the construct facts on the pinned rev.
- [ ] #6 `.agents/skills/gala-from-go/references/constructs.md` records the verified `Ptr`/`OptionPtr` spellings and checks this pass depends on.
- [ ] #7 The canonical Compose backend test sequence from `server/README.md` passes.
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation plan (2026-10-25)

### Phase 0 - provenance and inventory
- Confirm `gala version` 0.93.0, flake rev `8ba02b2b`, and the `server/GALA_COMPILER` extraction marker.
- Inventory every raw `*T` in the committed sources and classify each as convert or keep-with-reason.

### Phase 1 - pointer nil guards with OptionPtrOf
- Re-derive the TASK-0361 sites (`cloneUser`, `sanitizedResponseUser`, `populateSignUpResponsePayloadIdentity`, `routes/users.gala`, `routes/guest_response_ownership.gala`, `services/services.gala`).
- Rewrite each guard as `std.OptionPtrOf(p) match { case Some(v) => ...; case None() => ... }` where the Go-facing signature stays `*T` and behavior stays identical.
- Record every guard that stays Go-spelled and why.

### Phase 2 - `Ptr[T]` and `OptionPtr[T]`
- Convert non-null internal pointer parameters, returns, receivers, and locals to `Ptr[T]`; add `RequirePtr` only where a nullable Go `*T` enters GALA and nil was already a bug.
- Use `OptionPtr[T]` for absent-or-present pointers that stay inside GALA code; do not put one in a Go-facing signature or a struct field a Go caller or wire decoder reads.
- Keep raw `*T` at Go boundaries, nil-assigned variables, address targets, and any site where the conversion would add a nil panic that the original did not have.
- Trial-transpile and diff every changed file before regenerating the committed twins.

### Phase 3 - verify and record
- `verify.sh --write`, then `verify.sh` again for byte-identical regeneration, `gofmt -l`, `go build ./...`, `go vet ./...`, and the canonical Compose backend sequence.
- Update `server/GALA.md` with the pass, pointer counts, kept sites, and construct facts; update the gala-from-go constructs reference with the verified spellings and checks.
- Refresh the codebase-memory index; no commit unless requested.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Progress (2026-10-25)

Provenance re-confirmed before editing: `gala version` 0.93.0, flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`, extraction `6368c09...`, matching `server/GALA_COMPILER`.

Pointer pass: raw pointer declaration lines 162 -> 95 (223 -> 134 tokens). Converted all 31 `*gin.Context`, 6 `*gin.RouterGroup`, 8 mock-provider `*http.Request`, and 13 Discord handler pointer parameters to `Ptr[T]`; converted 13 non-null receivers (credentialCodec, tracingTransport, the four calendar providers, visitor) and 6 infallible results (NewRepository, NewRepositoryFromTx, NewReadinessMonitor, TruePtr, FalsePtr, calendarAccountToStorage); `cloneUser`/`sanitizedResponseUser` now take and return `Ptr[models.User]`.

`OptionPtrOf` now guards eleven sites; `RequirePtr` marks the three nullable-to-Ptr boundaries (`services.gala`'s two request results and `group.gala`'s `visitor.controls`). Kept raw pointers are the nullable Go results, Go struct fields/dereferences, address targets, package-level nil-initialized pointers, Go interface/callback boundaries, and Go-declared route helper parameters; the ledger section groups them with reasons.

Docs: `server/GALA.md` gained the pass section and registry style update; `gala-from-go/references/constructs.md` gained the `Ptr` row, the expanded `OptionPtrOf` row, the builtin-qualification trap, and the pointer-valued local-`val` note.

A full `verify.sh --check` run completed on 2026-10-25 with `OK (63 twins)`, 63/63 transpiled, `gofmt` clean, the orphan scan clean, and `go build ./...` green; the first attempt was cancelled after concurrent probe transpiles made it look stalled, and the clean re-run finished normally.

## Revert (2026-10-26)

The local `var` -> `val` sweep was reverted: all 354 keyword flips across 37 twins are undone, so the local `var` bindings match HEAD again and only the pointer changes remain.
A full `verify.sh --check` run after the revert reports `OK (63 twins)`: 63/63 byte-identical, `gofmt` clean, the orphan scan clean, and `go build ./...` green.
The first post-revert run caught one stale twin, `services/auth/auth.go`, whose only difference was the +2 `//line` mapping from the sweep's added blank and comment lines; restoring the twin to HEAD made it regenerate byte-identically.
<!-- SECTION:NOTES:END -->
