---
id: TASK-0369
title: >-
  File the GALA block-form struct-tag documentation defect and correct the
  translation ledger
status: Done
assignee:
  - deemp
created_date: '2026-10-27 14:21'
updated_date: '2026-10-27 14:25'
labels: []
dependencies: []
references:
  - 'https://github.com/martianoff/gala/issues/528'
  - >-
    https://github.com/martianoff/gala/blob/8ba02b2b18eca7ca3bc410f5e467b112366bdf6c/docs/GALA.MD
  - >-
    https://github.com/martianoff/gala/blob/8ba02b2b18eca7ca3bc410f5e467b112366bdf6c/internal/parser/grammar/gala.g4
  - >-
    https://github.com/martianoff/gala/blob/8ba02b2b18eca7ca3bc410f5e467b112366bdf6c/internal/transpiler/transformer/declarations.go
  - server/GALA.md
  - .agents/skills/gala-from-go/references/gaps.md
  - .agents/skills/gala-from-go/references/constructs.md
  - 'https://github.com/martianoff/gala/issues/885'
modified_files:
  - server/GALA.md
  - server/scripts/gala/probes/models-struct-tag/main.gala
  - server/scripts/gala/probes/models-struct-tag/notes.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
priority: medium
type: docs
ordinal: 373300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The GALA translation loop recorded struct tags as a wall for `models/location.go`, `models/event.go`, and `models/calendar.go`, based on a probe that only tested the shorthand form with a backtick tag. That probe tested the wrong spelling: on the pinned compiler (`GALA version 0.93.0`, flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`), a block-form declaration accepts a Go struct tag as an interpreted string and copies it verbatim into the generated Go.

Verified evidence:

- `internal/parser/grammar/gala.g4:73` is `structField: (VAL | VAR)? identifier type (STRING)?;` — only the block form has a tag slot.
- `transformStructField` in `internal/transpiler/transformer/declarations.go` copies `ctx.STRING()` into `field.Tag` with the comment “Copied verbatim into the generated Go”, and validates it as a struct field tag.
- `structs_test.go:99` and `string_escapes_test.go:285` cover the capability.
- A `models.Location` shape with `var` fields and tags transpiled to the exact Go struct; `json.Marshal` matched the handwritten struct byte for byte and `reflect.StructTag.Get("json")` returned `country_code`.
- The shorthand form rejects both quoted and backtick tags, and the block form rejects a backtick tag; only an interpreted string with escaped quotes works.
- The docs show no tag spelling, and #528 says tags are Go syntax GALA does not have and that a third-party reflection-tag type stays handwritten. That is a documentation contradiction, not a language gap.

This task files one upstream documentation defect and corrects the local record. It does not ask for an opt-out from synthesized `Copy`/`Equal`/`Unapply`/`StructMeta_X` and does not file the `models.Set` case: the opt-out is a policy-dependent feature request against a design #528 calls intended, and `Set` is blocked by `struct{}` and by the channel-blocked `GetUsersCalendarEvents` signature.

## Report to file

Title: Document the struct tag a block-form declaration accepts

```markdown
### A Go struct tag on a block-form struct

`type Location struct { var CountryCode string "json:\"country_code\"" }` is the Go-shaped spelling of a DTO, and its tag defines the wire shape. Block-form declarations accept that tag and copy it verbatim into the generated Go. The reference never shows the spelling, and #528 says struct tags are Go syntax GALA does not have.

**Classification:** defect (documentation), not a language gap.

**What works**
- `gala.g4:73`: `structField: (VAL | VAR)? identifier type (STRING)?;`
- `transformStructField` (declarations.go) copies the `STRING` into `field.Tag` (“Copied verbatim into the generated Go”) and validates it as a struct field tag.
- Covered by `structs_test.go:99` and `string_escapes_test.go:285`.

**Minimal repro**
```gala
package models

type Location struct {
    var CountryCode string "json:\"country_code\""
}
```
`gala transpile --stdlib-module go.gala.fyi/stdlib -i main.gala -o main.go` emits:
```go
type Location struct {
    CountryCode string "json:\"country_code\""
}
```
`json.Marshal(Location{CountryCode: "US"})` is `{"country_code":"US"}` and `reflect.StructTag.Get("json")` is `country_code`.

**The spellings that do not work**
- Shorthand: `struct Location(var CountryCode string "json:\"country_code\"")` → `extraneous input ... expecting ')'`.
- Block form with a raw-string tag: `var CountryCode string `json:"country_code"`` → `extraneous input ... expecting {'}', 'val', 'var', 'default', IDENTIFIER}`.

The only working spelling is an interpreted string with escaped quotes, and every field must be `var` to stay a plain Go field.

**Docs gap**
GALA.MD §4 “Block Struct Declaration” shows no tag; the same section says the block form “mirrors a Go declaration field for field” and is the right choice “for types that exist to meet Go.” GO_INTEROP.MD does not mention it either. #528's answer says tags are not GALA syntax and a reflection-tag type stays handwritten, which contradicts this path.

**Ask**
Document the working spelling in §4 and the Go-interop guide, with the emitted Go and the two rejected spellings, and reconcile #528's answer. If the capability is not meant to be public, reject the tag with a diagnostic instead of accepting it silently.

**Compiler**
GALA version 0.93.0; flake rev `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`; Go 1.26.7.

**Searched**
#528, #688, tracker searches for “struct tag” and “block-form tag”; docs §4 and GO_INTEROP; `gala explain --list` has no tag entry.
```

## Local corrections in scope

- Rewrite the committed probe at `server/scripts/gala/probes/models-struct-tag/` so it records the working block-form spelling, the two rejected spellings, and the synthesized members the block form emits.
- Correct the struct-tag row in `.agents/skills/gala-from-go/references/constructs.md` and the `answered`/“what not to file” treatment in `.agents/skills/gala-from-go/references/gaps.md`.
- Correct `server/GALA.md`: the findings row, the prose below it, cursor entries 1 and 4, and the report index after the issue is filed.
- Supersede TASK-0350's “no upstream action for tags” decision; this task is the new record.

## Out of scope

- Filing the synthesized-member opt-out request.
- Filing the `models.Set` / `map[T]struct{}` case.
- Relaxing the no-API-widening rule or running the five-file models batch.
- Any change to generated twins or runtime code.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The committed `models-struct-tag` probe records the working block-form spelling, both rejected spellings with their exact parse errors, and the emitted Go plus JSON evidence
- [x] #2 The `constructs.md` struct-tag row is a `workaround` that names the block-form interpreted-string spelling and the shorthand/raw-string rejections, and no longer calls tags an `answered` construct
- [x] #3 `gaps.md` no longer lists struct tags among answered “what not to file” constructs, and records the block-form capability plus the synthesized-member boundary gap
- [x] #4 `server/GALA.md` distinguishes the tag capability from the synthesized-member policy blocker in its findings, prose, and cursor, including the `models/calendar.go` const-versus-struct distinction
- [x] #5 The documentation defect is filed upstream with the `gaps.md` report template and its URL is recorded in `server/GALA.md`'s report index and the task notes
- [x] #6 TASK-0350's “no upstream action for tags” decision is explicitly superseded in the task notes and the ledger
- [x] #7 Changed Markdown passes `npm run format:markdown:check`, `npm run test:markdown-rules`, and `npm run test:markdown-format`
<!-- AC:END -->

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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Rewrite `server/scripts/gala/probes/models-struct-tag/main.gala` to the block-form `models.Location` shape, and `notes.md` to record the working spelling, the emitted tag plus synthesized members, both rejected spellings with their exact parse errors, and the scratch JSON/build evidence on 0.93.0 rev `8ba02b2b`.
2. Correct the `struct tag` row in `.agents/skills/gala-from-go/references/constructs.md` to a `workaround` naming the block-form interpreted-string tag and the shorthand/raw-string rejections, keeping `Codec[T]` as the substitute only when the wire shape may move.
3. Correct `.agents/skills/gala-from-go/references/gaps.md`: remove struct tags from the `answered`/what-not-to-file list, add the block-form capability to “Constructs That Were Never Gaps”, and record the synthesized-member boundary gap and its conditional opt-out report.
4. Search the upstream tracker once more, then file the documentation issue with the `gaps.md` template and record its URL.
5. Correct `server/GALA.md`: replace the tag finding with a tag row and a synthesized-member row, rewrite the prose below the findings table, rewrite cursor entries 1 and 4, update the #528 report-index row, and add the new issue to the report index.
6. Add a comment to TASK-0350 recording that TASK-0369 supersedes its “no upstream action for tags” decision, and record the same supersession in TASK-0369's notes.
7. Run `npm run format:markdown` and the Markdown checks, verify every acceptance criterion against the resulting files, then finalize the task.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Filed #885 and corrected the tag record (2026-10-27)

**Filed:** https://github.com/martianoff/gala/issues/885 — `Document the struct tag a block-form declaration accepts`. The body follows the `gaps.md` template: classification, what works, what I tried, exact diagnostics, minimal repro, caller need, docs gap, compiler, environment, and searched. It cites the grammar rule, `transformStructField`, the two tag tests, and the #528/§4 contradiction.

**Probe:** `server/scripts/gala/probes/models-struct-tag/main.gala` is now the block-form `models.Location` shape (seven `var` fields with interpreted-string tags). `notes.md` records the working spelling, the emitted tag, the JSON comparison, both rejected spellings with their exact parse errors, the synthesized members, and the history. Transpiling from the probe directory with `--stdlib-module go.gala.fyi/stdlib` emitted the exact tags plus `Copy`, `Equal`, `Unapply`, and `StructMeta_Location`; the scratch JSON test passed.

**Skill references:** `constructs.md`'s `struct tag` row changed from `answered`/`Codec` to `workaround`/block-form escaped tag, with the shorthand and raw-string rejections in the check. `gaps.md` removes struct tags from the answered list, adds the block-form “never a gap” entry, and adds `Boundary Gaps Without A Report` for the synthesized-member boundary and its conditional opt-out.

**Ledger:** `server/GALA.md` now has separate tag-workaround and synthesized-member boundary rows; the #528 report-index row is corrected; #885 is added to the report index; the prose below the findings table says `set.go` is `struct{}`, `location.go`/`event.go` are synthesized members (tags transpile), and `event.go`'s marshal methods need anonymous embedded structs; cursor 1 was rewritten and re-verified 2026-10-27; cursor 4 says the consts stay for `swag` and the structs are blocked by synthesized members (the split keeps the enum); the anonymous-struct row no longer blames tags; the `errs` `GoogleAPIError` reason is corrected.

**Supersession:** TASK-0350's 2026-10-05 “no upstream action for tags” is superseded; a comment is added there. Fixed-size arrays, `struct{}`, `chan`/`recover`, `const`, and anonymous/embedded shapes remain non-reportable.

**Checks:** `npm run format:markdown` (rewrote `server/GALA.md`), `npm run format:markdown:check`, `npm run test:markdown-rules` (30 tests), and `npm run test:markdown-format` (30 tests) all pass. Documentation-only: no unit/e2e, Swagger, or code-graph work applies.

## Final verification (2026-10-27)

- The committed probe transpiled from `server/scripts/gala/probes/models-struct-tag/` with `gala transpile --stdlib-module go.gala.fyi/stdlib -i main.gala -o <scratch>/location.go`; the emitted struct kept all seven tags and emitted `Copy`, `Equal`, `Unapply`, and `StructMeta_Location`.
- The scratch module required `go.gala.fyi/stdlib v0.93.0`; its test marshaled the generated `Location` against a handwritten struct, both produced `{"country_code":"US",...}`, `reflect.StructTag.Get("json")` returned `country_code`, and `go test` passed.
- `gh issue view 885` reports `OPEN` at https://github.com/martianoff/gala/issues/885.
- `npm run format:markdown:check`, `npm run test:markdown-rules` (30 tests), `npm run test:markdown-format` (30 tests), `npm run fmt:check`, and `git diff --check` pass.
- All seven acceptance criteria and all eight Definition of Done items are checked.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Filed the GALA block-form struct-tag documentation defect as https://github.com/martianoff/gala/issues/885 and corrected the translation record that had treated struct tags as a wall.

## What changed

- `server/scripts/gala/probes/models-struct-tag/`: the probe now carries the block-form `models.Location` shape with seven `var` fields and interpreted-string JSON tags. `notes.md` records the working spelling, the emitted tag, the JSON comparison, both rejected spellings with exact parse errors, the synthesized members, and the compiler history.
- `.agents/skills/gala-from-go/references/constructs.md`: the `struct tag` row is a `workaround` for the block-form escaped-tag spelling; the shorthand and raw-string forms are parse errors, and `Codec[T]` is the substitute only when the wire shape may move.
- `.agents/skills/gala-from-go/references/gaps.md`: struct tags leave the answered/“what not to file” list; the block-form capability is recorded under “Constructs That Were Never Gaps”; a new “Boundary Gaps Without A Report” section records the synthesized-member boundary and the conditional opt-out request.
- `server/GALA.md`: separate tag-workaround and synthesized-member boundary rows; corrected `#528` report-index row; new `#885` report-index row; prose and cursor now say `set.go` is blocked by `struct{}` and `location.go`/`event.go` by the synthesized members (tags transpile), with `event.go` also needing anonymous embedded structs; `calendar.go`'s consts stay for `swag` and the structs are blocked by synthesized members; the anonymous-struct row and the `errs` `GoogleAPIError` reason no longer blame tags; TASK-0350's tag decision is explicitly superseded.
- `TASK-0350`: comment added recording the supersession by TASK-0369.

## Filed

https://github.com/martianoff/gala/issues/885 — “Document the struct tag a block-form declaration accepts”. It follows the `gaps.md` report template and cites the grammar, `transformStructField`, the two tag tests, the emitted Go/JSON evidence, and the #528/§4 contradiction.

## Verification

- The committed probe transpiles from its directory with `--stdlib-module go.gala.fyi/stdlib`; the emitted struct carries all seven tags plus `Copy`, `Equal`, `Unapply`, and `StructMeta_Location`.
- A scratch module requiring `go.gala.fyi/stdlib v0.93.0` ran a test that marshaled the generated `Location` against a handwritten struct: both produced `{"country_code":"US",...}`, `reflect.StructTag.Get("json")` was `country_code`, and `go test` passed.
- `#885` is OPEN at the filed URL.
- `npm run format:markdown:check`, `npm run test:markdown-rules` (30 tests), `npm run test:markdown-format` (30 tests), and `npm run fmt:check` pass; `git diff --check` is clean.

## Scope

The synthesized-member opt-out report and the `models.Set` report were deliberately not filed, and no generated twin or runtime code changed. Unit/e2e tests are exempt for this documentation-only change; no Swagger annotations or runtime code changed.

## Follow-up

The five-file models batch still depends on the no-API-widening decision; if the rule stays, the opt-out request is the next report to draft.
<!-- SECTION:FINAL_SUMMARY:END -->
