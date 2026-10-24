---
id: TASK-0354
title: >-
  Rewrite the server GALA twins' error and nil handling with the adopted runtime
  idioms
status: Done
assignee:
  - opencode
created_date: '2026-10-10 13:19'
updated_date: '2026-10-10 13:48'
labels:
  - gala
  - server
dependencies: []
references:
  - server/GALA.md
  - server/scripts/gala/verify.sh
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-loop/SKILL.md
ordinal: 358300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The committed twins still spell failure the Go way: `val x, err = f()` plus `if err != nil`, `var err error` plus reassignment, and bare `nil` checks, even though every one now imports `go.gala.fyi/stdlib`. The adopted runtime makes those sites expressible as values: a Go call returning `(T, error)` used as a value is already a `Try[T]`, `match` takes `Success`/`Failure` apart, `OnFailure`/`GetOrElse`/`OnSuccess` handle the ends, `std.FromError`/`Try(...)` handles a void error return, and `Option` with `go_interop.OptionFromMap` covers absence at boundaries internal to a `.gala` file. Rewrite those sites where the emitted Go keeps exactly the current behavior and the package's Go-facing API and wire format, and record the ones that must stay on the Go spelling.

Files with err/nil today: middleware/auth.gala, utils/request_utils.gala, routes/respondent_identity.gala, routes/users.gala, routes/guest_response_ownership.gala, services/services.gala, slackbot/commands/num_users.gala, slackbot/commands/active_users.gala, discord_bot/init.gala, discord_bot/commands/num_users.gala, discord_bot/commands/active_users.gala.

Guardrails: many names cross into handwritten Go — `populateSignUpResponsePayloadIdentity` is called by routes/event_routes.go, `CallApi` by services/*, and route tests call `canonicalGuestName`, `guestNameValidationErrorMessage`, and `shouldExposeGuestSignUpResponsePayload` — so those signatures do not change; only bodies and strictly internal helpers change. The runtime is a value layer, not a behavior change: `Try(...)`-style panic catching must not replace a propagating panic with a Failure where the current code propagates, and every panic path keeps its exact logger call.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every remaining `err` binding or `!= nil` check in the twins is either rewritten with the runtime spelling (a Go call used as a `Try` value, `match` on `Success`/`Failure`, `OnFailure`/`OnSuccess`, `GetOrElse`, `FromError`, `bind`) or kept with the reason recorded in the task notes and `server/GALA.md`.
- [x] #2 `nil` checks on Go-boundary pointers are rewritten with `Option` only where every caller is inside the `.gala` file or its package's generated code and the emitted Go still builds; pointers whose signature is read by handwritten Go keep the check and are recorded.
- [x] #3 The rewritten twins pass `server/scripts/gala/verify.sh` (byte-identical regeneration, gofmt-clean, only `go.gala.fyi/stdlib` runtime imports), `go build ./...`, and the canonical Compose backend test sequence from server/README.md.
- [x] #4 Observable behavior is unchanged: the same panic site and logger call on the same errors, the same HTTP statuses, and the same wire payloads; no test is relaxed and no generated file is hand-edited.
- [x] #5 `server/GALA.md` records the idiom pass and any finding change, and `.agents/skills/gala-from-go/references/constructs.md` records the confirmed spelling and check for each error/nil substitute this pass depends on (Try-as-a-value on a Go call, OnFailure/Get vs match, OptionFromMap).
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation plan (2026-10-10)

### Phase 0 - provenance
- Confirm `nix develop` gives GALA 0.87.1 at flake rev `cd2fdcb5` and the extraction marker matches `server/GALA_COMPILER`.

### Phase 1 - pin the runtime idioms on the compiler (scratch probes under /tmp/opencode, never committed)
- Pin `f()` used as a `Try` value for a `(T, error)` Go call (emits `std.GoTry`), `match` on `Success`/`Failure`, `OnFailure(...).Get()`, `FromError(voidCall)`, and `Try(voidCall)` (emits `std.Try[Void]{}.Apply`, which catches panics) against `go.gala.fyi/stdlib@v0.87.1`.
- Pin `go_interop.OptionFromMap`, including the lambda-annotation case for a map whose type comes from a handwritten sibling.
- Pin Go-signature resolution through import aliases: a differently named alias loses the signature; an alias equal to the package's own name resolves.

### Phase 2 - rewrite the twins (scratch transpile per file, then `verify.sh --write`)
- `middleware/auth.gala`: nested `match` over `accounts.Resolve`/`accounts.LoadSessionUser` with the same 401/500 JSON responses.
- `utils/request_utils.gala`, `services/services.gala`, both `num_users.gala` and both `active_users.gala` twins: `OnFailure((err) => logger.StdErr.Panicln(err)).Get()`; `strconv.Atoi` through `match`.
- `routes/respondent_identity.gala`: `go_interop.OptionFromMap` with `Some(liveUser) if liveUser != nil` / `case _`.
- `discord_bot/init.gala`: `bot.User`/`FromError(bot.Open())`, `commandMap` through `OptionFromMap(...).ForEach` with an annotated lambda, and keep `discordgo.New`'s publish-then-check binding.
- Keep pointer nil guards and blank error discards with reasons recorded.

### Phase 3 - verify and record
- `verify.sh --write`, then a second `verify.sh` for byte-identical regeneration; canonical Compose backend sequence; `npm run format:markdown`; refresh the codebase-memory index.
- Record the pass, kept sites, and the alias finding in `server/GALA.md`; record the confirmed spellings and checks in `.agents/skills/gala-from-go/references/constructs.md`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation notes (2026-10-10)

Provenance re-confirmed before editing: `nix develop` provides GALA 0.87.1 (`/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`), flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c`, extraction `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa`, matching `server/GALA_COMPILER`. Scratch probes ran from `server/` into `/tmp/opencode/gala-idioms/` (never committed); each rewrite was scratch-transpiled and the emitted Go read before the committed twin was regenerated.

### Idioms pinned on the pinned compiler

- A Go call returning `(T, error)` used as a value is a `Try[T]`: the transpiler emits `std.GoTry(call(...))`; `match` on `Success`/`Failure`, `OnFailure`, `OnSuccess`, and `GetOrElse` all compile against it. `call(...).OnFailure((err) => logger.StdErr.Panicln(err)).Get()` logs and panics with the same error in the same function, so the Go spelling's panic site and logger call are kept.
- `FromError(voidCall)` emits `std.FromError(...)` and only checks the error; `Try(voidCall)` emits `std.Try[Void]{}.Apply` over a closure that panics on the error, which would catch a panic the call itself raised. The pass used `FromError` so no propagating panic became a `Failure`.
- `go_interop.OptionFromMap(m, k)` emits `std.Option` and covers nil maps and missing keys; for a map whose type comes from a handwritten sibling (`commandMap = newCommandMap()`), the `ForEach` lambda needs an explicit parameter type or the transpiler leaks its own `T` into the generated Go.
- A Go call behind an import alias whose name differs from the package's own name loses its resolved signature: `pgstore "timeful/server/postgres"` emitted `.OnFailure` without `std.GoTry` (rejected by `go build`); re-spelling the import as `"timeful/server/postgres"` and calling `postgres.DefaultRepository()` emitted `std.GoTry`. The bot and command twins dropped the `pgstore` alias for this reason.

### Rewritten sites

- `middleware/auth.gala`: `accounts.Resolve` and `accounts.LoadSessionUser` are matched as `Try` values; failure branches emit the same 401 `UserDoesNotExist` and 500 `failed to load account integration data` JSON responses, aborts, and returns, and success sets the same `authUser`/`authAccount` context keys and calls `Next`.
- `utils/request_utils.gala`: `url.QueryUnescape(s).OnFailure((err) => logger.StdErr.Panicln(err)).Get()`.
- `services/services.gala`: `http.DefaultClient.Do(req).OnFailure((err) => logger.StdErr.Panicln(err)).Get()`; the `user`/`body` parameter checks and the `req` variable stay Go-spelled.
- `slackbot/commands/{num_users,active_users}.gala` and `discord_bot/commands/{num_users,active_users}.gala`: `postgres.DefaultRepository()` and `repository.CountAccounts`/`ListActiveUserDays` are `Try` values; `strconv.Atoi` arguments are matched as `Success(parsed) => days = parsed` / `Failure(_) => {<same message>; return}` and `var err error` is gone.
- `routes/respondent_identity.gala`: the `liveUsers` comma-ok lookup became `go_interop.OptionFromMap(liveUsers, lookupKey)` matched as `Some(liveUser) if liveUser != nil` with `case _` carrying the unchanged fallback, so a missing key and an explicit nil value both fall back as before.
- `discord_bot/init.gala`: `bot.User("@me")` through `OnFailure(...).Get()`; `bot.Open()` through `FromError(...).OnFailure(...)`; `commandMap` through `OptionFromMap(commandMap, commandName).ForEach((command commands.Command) => command.Execute(s, m, args))`.

### Kept Go-spelling sites and reasons (also recorded in server/GALA.md)

- Pointer nil guards: `routes/respondent_identity.gala` `cloneUser` (`user == nil`), `sanitizedResponseUser` (`sanitized == nil`), `populateSignUpResponsePayloadIdentity` (`response == nil`, `liveUser != nil`); `routes/users.gala` (`user == nil`); `routes/guest_response_ownership.gala` (`response == nil`); `services/services.gala` (`user != nil`, `body != nil`). The adopted runtime has no pointer-to-Option conversion (`go_interop.OptionFromMap` is its only absence-to-Option helper), so an Option rewrite would relocate the same nil test into a `Some`/`None` constructor without removing it. `populateSignUpResponsePayloadIdentity` is called by `routes/event_routes.go`, `CallApi` by `services/*`, and `canonicalGuestName`/`guestNameValidationErrorMessage`/`shouldExposeGuestSignUpResponsePayload` by `routes/guest_response_ownership_test.go`, so those signatures are fixed; `liveUser != nil` distinguishes an explicit nil map value from a missing key.
- `discord_bot/init.gala` `val botSession, err = discordgo.New(...)`: the original publishes the returned session to the package-level `bot` before the error check, and a `Try` carries no value on `Failure`, so no value spelling preserves publish-then-check. `discordgo.New` v0.27.1 always returns a non-nil session, and the panic site and logger call are unchanged.
- Blank error discards stay blank: `json.Marshal` and `http.NewRequest` in `services.gala`, `bot.GuildChannels` in `init.gala`; the original deliberately ignores those errors.

### Evidence

- `server/scripts/gala/verify.sh --write` then `server/scripts/gala/verify.sh`: `OK (22 twins)` on both runs - every twin regenerates byte-identically under `--stdlib-module go.gala.fyi/stdlib`, is gofmt-clean, imports only `go.gala.fyi/stdlib/...` (no `martianoff/gala`), and `go build ./...` is green. Nine twins were rewritten; the other thirteen regenerated unchanged.
- Canonical Compose backend sequence from `server/README.md`: `docker compose ... run --rm server-route-test` exits 0 with `go test ./... -count=1`; 15 packages `ok`, no FAIL. This includes the route tests that read `canonicalGuestName`, `guestNameValidationErrorMessage`, and `shouldExposeGuestSignUpResponsePayload`.
- `npm run format:markdown` (GALA.md already conformant), `npm run format:markdown:check`, and `npm run fmt:check` are clean.
- `codebase-memory-mcp cli index_repository --repo-path .`: `status: indexed`, 9007 nodes / 28409 edges.
- No swag annotation changed, so `swag init` and `npm run gen:api` were not run; no browser-e2e spec covers these server internals and the task's required test layer (canonical Compose backend sequence) is green.
- No test was changed or relaxed, and no generated file was hand-edited (all regenerated by `verify.sh`).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Rewrote the committed server GALA twins' Go-spelled error and nil handling with the adopted `go.gala.fyi/stdlib` runtime idioms on the pinned compiler (flake rev `cd2fdcb5`, GALA 0.87.1).

Rewritten: `middleware/auth.gala` (nested `match` on `accounts.Resolve`/`LoadSessionUser` with the same 401/500 responses), `utils/request_utils.gala`, `services/services.gala`, all four bot command twins (`OnFailure(...).Get()` and `match` on `strconv.Atoi`, with `postgres.DefaultRepository()` un-aliased so the transpiler resolves its `(T, error)` signature), `routes/respondent_identity.gala` (`go_interop.OptionFromMap` for the `liveUsers` lookup), and `discord_bot/init.gala` (`OnFailure(...).Get()` for `bot.User`, `FromError` for `bot.Open`, and `OptionFromMap(...).ForEach` for `commandMap`).

Kept on the Go spelling with reasons recorded in `server/GALA.md` and the task notes: the pointer nil guards in `respondent_identity`, `users`, `guest_response_ownership`, and `services` (the runtime has no pointer-to-Option conversion and the Go-facing signatures are fixed), `discord_bot/init.gala`'s `discordgo.New` binding (publishes the session before the error check; a `Try` carries no value on `Failure`), and the deliberately ignored blank errors. `FromError` was used for the void `bot.Open()` call so a propagating panic is never caught and converted.

Documentation: `server/GALA.md` gains a runtime idiom pass section with the kept-site reasons and the alias finding; `.agents/skills/gala-from-go/references/constructs.md` records the confirmed `Try`-as-a-value, `OnFailure`/`Get` vs `FromError`/`Try`, `OptionFromMap`, and alias-signature spellings with their checks.

Evidence: `verify.sh` `OK (22 twins)` with byte-identical regeneration, gofmt-clean, only `go.gala.fyi/stdlib` imports, and `go build ./...` green; canonical Compose backend sequence exit 0 (15 packages `ok`); Markdown and `fmt:check` clean; codebase-memory index refreshed (9007 nodes / 28409 edges). No swag annotation changed; no test relaxed; no generated file hand-edited. No commit made (not requested).
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
