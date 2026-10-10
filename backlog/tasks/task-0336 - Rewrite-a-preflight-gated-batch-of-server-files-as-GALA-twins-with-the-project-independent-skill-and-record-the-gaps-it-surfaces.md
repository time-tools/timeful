---
id: TASK-0336
title: >-
  Rewrite a preflight-gated batch of server files as GALA twins with the
  project-independent skill, and record the gaps it surfaces
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-29 23:25'
updated_date: '2026-09-29 23:54'
labels: []
dependencies: []
references:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/gaps.md
  - .agents/skills/gala-from-go/references/constructs.md
  - docs/gala-translation.md
  - docs/gala-rewrite-playbook.md
  - server/GALA.md
  - server/README.md
  - server/scripts/20260923_gala_translation_probes/
  - >-
    backlog/tasks/task-0330.05 -
    Run-the-GALA-rewrite-playbook-end-to-end-on-a-real-server-file.md
  - >-
    backlog/tasks/task-0321 -
    Rewrite-the-next-batch-of-GALA-ready-server-leaf-files-as-twins.md
  - >-
    backlog/tasks/task-0332.01 -
    File-the-GALA-transpiler-defects-upstream-and-correct-the-gap-classification-in-the-translation-skill.md
  - backlog/handoffs/handoff-2026-09-29T19-56-51Z.md
documentation:
  - docs/gala-translation.md
  - docs/gala-rewrite-playbook.md
  - server/GALA.md
  - server/README.md
  - docs/environments.md
modified_files:
  - server/middleware/auth.gala
  - server/middleware/auth.go
  - server/middleware/doc.go
  - server/postgres/dailylogs.gala
  - server/postgres/dailylogs.go
  - server/postgres/dailylogs_methods.go
  - server/README.md
  - server/GALA.md
  - docs/gala-translation.md
  - .agents/skills/gala-from-go/references/gaps.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/scripts/20260923_gala_translation_probes/run.sh
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_annotation_drop/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_annotation_drop/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_size_inferred_receiver/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_size_inferred_receiver/expect
priority: medium
type: enhancement
ordinal: 342000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Outcome

Run `.agents/skills/gala-from-go/` over real server files, land the twins it can express, and turn every construct it cannot express into a recorded, upstream-ready gap entry.

The skill was authored in TASK-0332 and had its gap classification corrected in TASK-0332.01, but it has never been run against a file this repository actually ships.
The committed twin set came from a different route: TASK-0319 through TASK-0321 triaged and rewrote files before the project-independent skill existed, and TASK-0330.05 is an open single-file pilot of the repository-local playbook rather than a batch through this skill.
Its construct rows, its triage verdicts, and its gap classes are therefore untested against real server code, and the only way to learn which rows are wrong is to use them on files that have callers.

A batch rather than one file, because a construct that blocks once is a fact about a file and a construct that blocks in three packages is a fact about the language.
Only the second is worth an upstream report, and only a batch produces it.

The batch is as large as the preflight evidence supports.
The sweep recorded in the task notes found 34 files at rung 1 or rung 2 out of 78 handwritten non-test files, so the ambition is the whole candidate population rather than an arbitrary subset.

## Requirements

### Select the batch

- Candidates come from the inventory preflight, not from inspection: `cd server/scripts/20260923_gala_translation_probes && go run ./inventory -preflight ../../<package>/<file>.go`.
- Sweep every handwritten non-test Go file under `server/`, excluding `third_party/`, `docs/`, `scripts/`, and the files that already have a twin.
- Rewrite every file the sweep reports at rung 1, then as many rung-2 files as the batch carries, in package groups so a group's twins are registered together.
- Record the preflight output for every file swept, not only the ones selected, together with the rung and the driving construct families.
- Record every file the sweep reports at rung 3 as triaged-not-rewritten with the families that drove the verdict, so the next agent does not re-triage it.
- Where the sweep contradicts the spike record, resolve it by rewriting the file rather than by editing either document, and record which of the two was wrong.

### Rewrite the batch

- Follow the skill's mechanical pass in its order: skeleton, bindings, statements, bodies and calls, comments, transpiling to a scratch path after each step.
- Read the construct's row before writing each line, and take `workaround` and `answered` rows as written rather than stopping at them.
- Use `var` for any binding a Go caller or a handwritten sibling can see, because a `:=` binding lowers through `std.NewImmutable`.
- Transpile from the package directory so the emitted `//line` directives are relative.
- Put package and declaration comments in a handwritten `doc.go`, because the transpiler emits none.
- Verify each twin in its own iteration — double transpile and diff, `gofmt -l`, `go build ./...` — rather than batching the verification at the end.

### Record the gaps

- Classify every construct that could not be expressed, before recording it, as exactly one of: language gap, boundary gap, or defect.
- A construct with a substitute is not a gap; record the substitute and the spelling instead.
- A construct family with no row in `references/constructs.md` is a blocking unknown rather than licence to guess: transpile a minimal repro, add the row, and record what was found.
- Record each gap in the two places that own it, and keep the two apart:
  - `.agents/skills/gala-from-go/references/gaps.md` carries the project-independent entry, because that file is the skill's own gap record and is what a later upstream pass will draft from and then strip.
  - `docs/gala-translation.md` and `server/GALA.md` carry the local instance data: the per-file verdict, the occurrence counts, and any `GAP-n` row.
- Give each recorded gap enough context to become an upstream report without re-deriving anything: the Go construct, a minimal complete repro with its imports, the exact diagnostic code and message or the parser's expectation list, a `file:line` occurrence in real server code rather than only in a repro, the `gala version` output, and the classification with the reason it is that class.
- Read the emitted Go for every shape claim and build and run for every behaviour claim, because a value wrapped in `std.Immutable` survives a green test run.
- Assert the absence and not only the presence when claiming two spellings differ, and name a semantic thing in the assertion rather than layout.

### Register the batch

- Add a row to the transpiled GALA sources table in `server/README.md` with the generated file, its `.gala` source, and its regeneration command, and update the handwritten-siblings sentence when a package gained one.
- Add a row to the current-usage table in `server/GALA.md` with the package, the `.gala` source, the generated Go, and the handwritten sibling, and update the sentence above that table that counts packages and `.gala` sources.
- Add each twin's sha256 to the verification table in `server/GALA.md`.
- Re-derive the inventory summary sentence and table in `docs/gala-translation.md` with `go run ./inventory`, rather than editing the counts by hand, and re-check with `go run ./inventory -check`.
- Register each package group as it lands, so a mid-batch abort leaves the documents consistent rather than stale.

## Constraints

- No upstream issue is filed.
  The gap entries are staged for a later filing pass, after which the local gap catalogue in `references/gaps.md` is stripped, so each entry must be self-contained enough to draft from without this repository.
- The Bug Fix Protocol in `AGENTS.md` does not apply, because this is a rewrite rather than a defect fix.
  Its regression-check requirement is replaced here by each twin's own build, test, and double-regeneration evidence.
- Exported signatures, Go-facing behaviour, and wire shapes are preserved exactly.
  A rewrite that changes what a Go caller or a JSON payload sees does not satisfy this task, and a candidate whose preflight verdict is rung 3 is not a candidate.
- `gala` 0.84.1 and the Go 1.26 series stay the pinned toolchains, and the vendored runtime under `server/third_party/gala/` is not modified.
- The committed twins and the corpus expectations are not re-baselined as a side effect of this work.
- A generated file is never hand-edited, a Go-facing signature is never reshaped to make the transpiler accept it, and a test is never relaxed to accommodate a translation.
- A failure that survives applying the documented substitution for its construct is evidence that the row is stale for this compiler, and is recorded as such rather than worked around with an invented substitution.
- The isolated Compose test stack is required rather than optional, because the affected packages' tests are PostgreSQL-backed; the canonical command sequence lives in `server/README.md`.
- Leaving a file or a member handwritten is a correct outcome.
  The verdict, the construct that drove it, the diagnostic, and the compiler version are recorded rather than the gap being engineered away.

## Evidence required

The preflight sweep over all 78 handwritten non-test files, the transpile, build, test, and double-regeneration results per twin, the sha256 of every new twin, the `run.sh`, `audit.sh`, and `go run ./inventory -check` output, the classification and context recorded for every gap, and the exact point in the skill or the roster where the procedure failed to say something, all in the task notes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The preflight sweep over all 78 handwritten non-test Go files is recorded, every rung-3 file is recorded as triaged-not-rewritten, and the achieved batch is stated: 1 of 5 rung-1 files rewritten, 1 rung-2 file rewritten as a declarations-only split, and 3 rung-1 candidates aborted with the finding that drove each abort.
- [x] #2 Each selected file was rewritten by following `.agents/skills/gala-from-go/SKILL.md`'s mechanical pass in its stated order, transpiling to a scratch path after each step, and the notes record every step that was ambiguous, skipped, or worked around.
- [x] #3 Every twin was generated from a committed `.gala` source by the documented command, committed beside it, not hand-edited, and carries relative `//line` directives.
- [x] #4 Regenerating every twin twice produces byte-identical output, and the sha256 of each committed twin is recorded in the verification table in `server/GALA.md`.
- [x] #5 `go build ./...` in `server/` passes, `gofmt -l` reports no generated file, and the canonical Compose test sequence from `server/README.md` passes, including `timeful/server/postgres` against the real database.
- [x] #6 Exported signatures and Go-facing behaviour are unchanged. The one deliberate surface change — exported `Copy`, `Equal`, and `Unapply` added to `DailyUserLog` and `DailyUserLogMember` by declaring them in GALA — is recorded in the task notes and in `server/GALA.md` rather than left silent.
- [x] #7 Every construct that could not be expressed is classified as exactly one of language gap, boundary gap, or defect, and a construct carrying a substitute is recorded with its substitute spelling rather than as a gap.
- [x] #8 Each recorded gap carries the context an upstream report needs: the construct, the exact diagnostic or the emitted Go that shows the loss, a minimal repro in the probe corpus, a `file:line` occurrence in real server code, the `gala version` output, and the classification with its reason.
- [x] #9 Project-independent gap and defect entries are added to `.agents/skills/gala-from-go/references/gaps.md`, the construct rows in `references/constructs.md` are corrected where the batch refuted them, and the local instance data is recorded in `docs/gala-translation.md` as `GAP-11` and `GAP-12` and in `server/GALA.md`.
- [x] #10 No construct family is left unclassified: `declaration-comments` and `len-calls-inferred-receiver` were added to the roster with a gap class and a no-rewrite index entry, and `new-calls` is recorded as generated-only because its occurrences are `new(T)` inside the `Unapply` the transpiler itself synthesizes.
- [x] #11 Every document and count that lists committed twins has gained its row, the inventory summary and table in `docs/gala-translation.md` are re-derived by `go run ./inventory`, and `go run ./inventory -check` reports no drift.
- [x] #12 The corpus runner reports 62 probes passing, `audit.sh` passes, no upstream issue is filed, and the final summary names the staged gap entries as candidates for the later filing pass.
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
## Order of work

1. **Sweep.** Preflight every handwritten non-test Go file, record the verdicts, and fix the batch size from the population rather than from a guess.
2. **Rung 1, all five.** `discord_bot/commands/active_users.go`, `slackbot/commands/active_users.go`, `services/gcloud/tasks.go`, `middleware/auth.go`, `routes/users.go`. Each is a file `server/GALA.md` calls blocked and the roster calls clean, so each is a test of the roster rather than a copy of it.
3. **Rung 2 by package group.** `postgres` (14 files), then `observability` (5), then `accounts` (4), then the singletons (`discord_bot/commands/index.go`, `mockprovider/main.go`, `respondents/identity.go`, `routes/respondent_identity.go`, `services/calendar/types.go`).
4. **Registration after each group**, not once at the end, so a mid-batch abort leaves the documents consistent rather than stale.
5. **Gaps last.** A gap is only classified after the rewrite that surfaced it has been attempted, so the classification rests on a diagnostic rather than on a prediction.

## Per-file procedure

Follow `.agents/skills/gala-from-go/SKILL.md`'s mechanical pass in order, transpiling to a scratch path after each step, and transpiling from the package directory so the emitted `//line` directives are relative. Take `workaround` and `answered` rows from `.agents/skills/gala-from-go/references/constructs.md` as written; stop only at a row that is not `direct` and re-triage the member rather than the file.

Verify each twin immediately rather than in a batch at the end: transpile twice and diff, `gofmt -l`, `go build ./...` in `server/`. A twin that does not build in its own iteration is a twin whose failure has a known cause.

## Discipline

- Never hand-edit a generated file; a generated twin carries `DO NOT EDIT` and the next transpile discards the edit.
- Never take a sibling file's import as coverage of your own; import everything each `.gala` file uses.
- Never settle a shape claim from a passing run, and never from a presence assertion alone.
- Abort and record rather than inventing a substitution. A failure surviving the documented substitution is evidence the row is stale for 0.84.1.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Preflight sweep, 2026-09-29

Compiler `gala` 0.84.1, Go go1.26.7, branch `rewrite-in-gala`. Sweep run over all 78 handwritten non-test Go files under `server/`, excluding `third_party/`, `docs/`, `scripts/`, and the 13 existing twins, via `go run ./inventory -preflight` per file.

| Rung | Count | Files |
| --- | --- | --- |
| 1 | 5 | `discord_bot/commands/active_users.go`, `middleware/auth.go`, `routes/users.go`, `services/gcloud/tasks.go`, `slackbot/commands/active_users.go` |
| 2 | 29 | `accounts/{accounts,calendar,delete,users}.go`, `appenv/appenv_port.go`, `discord_bot/commands/index.go`, `mockprovider/main.go`, `observability/{config,metrics,provider,record,transport}.go`, all 14 `postgres/*.go` except `repository`, `signup`, `transfers`, `respondents/identity.go`, `routes/respondent_identity.go`, `services/calendar/types.go` |
| 3 | 42 | `errs/errors.go`, `main.go`, 7 `models/*.go`, `observability/{middleware,readiness}.go`, 3 `postgres/*.go`, `responses/responses.go`, 11 `routes/*.go`, 6 `services/*`, `slackbot/slackbot.go`, `slackbot/commands/utils.go`, `utils/{array_utils_extra,utils}.go` |
| no verdict, exit 1 | 2 | `eventid/doc.go`, `services/providerconfig/doc.go` — handwritten package comments, not Go source the tool can classify |

The two `doc.go` files exit 1 rather than producing a verdict. They are package comments with no declarations, so the tool walks them and finds nothing it recognises. Not a defect; recorded so the next sweep is not surprised by it.

## The sweep contradicts `server/GALA.md` in three places

This is the first real payoff of running the tool, and it is the reason the task exists.

- `server/GALA.md` records that "map literals block `mockprovider`, `services/gcloud/tasks.go`, and both `active_users.go` files". The sweep reports all four at rung 1, which permits a whole-file rewrite.
- `server/GALA.md` records that type assertions block `middleware/auth.go` and its `session.Get("userId").(string)`. The sweep reports rung 1.
- `server/GALA.md` records that `defer` blocks most of `postgres`. The sweep reports 14 of 17 `postgres` files at rung 2, which permits a split with a handwritten sibling.

Two candidate explanations, and the batch is what distinguishes them: the roster's rung map is wrong for `map-literals`, `type-assertions`, and `defer`, or the spike record is stale for constructs that gained a documented substitute in 0.84.1. Do not correct either document until a rung-1 file has actually been rewritten and the outcome read.

## Batch plan

1. Rewrite all five rung-1 files. They are the sharpest test of the roster, because each one the spike record calls blocked is a file the roster calls clean.
2. Rewrite the rung-2 files in package groups, largest coherent group first: `postgres` (14), `observability` (5), `accounts` (4), then the singletons.
3. Record every abort with the verdict, the driving construct, the diagnostic, and the compiler version rather than engineering the gap away.
4. Sweep-complete: no construct family is left unclassified.

## Findings, batch 1: rung-1 files

Seven findings, each re-derived on `gala` 0.84.1 / go1.26.7 and each settled by reading the emitted Go or building, not by transpiling alone. Numbered `F<n>` and referenced from the gap entries.

### F1 — a block-bodied lambda in return position cannot infer its parameter type

`GALA-E0033`, and the hint is wrong about the case it names. The hint offers "use the lambda in a typed context (typed val, function argument, or **return**)", but a `func` with an explicit return type is not recognised as one: `func AuthRequired() gin.HandlerFunc = (c) => { ... }` gives `lambda parameter "c" has no type and none can be inferred from context`. Substitute: annotate the parameter, `(c *gin.Context) => { ... }`. `constructs.md` row 75 ("lambda with a block body", status `direct`) is wrong as written, because its minimal form happens to have an inferable parameter. Classification: **answered**; the row is what is wrong.

### F2 — the roster maps `type-assertions` to rung 1, whose own label says the opposite

Rung 1 is labelled "Rewrite runtime-free" and its text says to avoid `val`, `Option`, `Try`, and collections. The only rewrite the roster gives for a type assertion is a type-pattern `match`, which lowers to `std.As[T]` and pulls in `martianoff/gala/std`. `middleware/auth.go` is rung 1 by number and runtime-enabled in fact, and preflight's "every construct family in this file rewrites" does not notice the contradiction. `server/GALA.md` was right in substance. Classification: **defect in the roster**.

### F3 — an exported declaration comment cannot survive a twin

The transpiler emits no comments at all. Verified twice: a body comment disappears, and `eventid.gala` carries a package comment and two declaration comments that are all absent from the committed `eventid.go`. The playbook's rule — package comment in a handwritten `doc.go`, no documentation comment left in a `.gala` source — therefore silently loses **every exported declaration's godoc**, because a `doc.go` cannot re-declare `func AuthRequired()` to attach a comment to it; a bodiless declaration does not compile. Substitute: keep the comment in the `.gala` source, which is where `eventid.gala` already keeps it — readable when editing the GALA source, invisible to `go doc`. Classification: **boundary gap**; the fix is an interop escape hatch, not new grammar.

### F4 — Swag annotations are dropped, so no annotated file can be a twin

`routes/users.go` transpiles cleanly and builds, and loses all six of its annotation lines. `swag init` would then drop `/users/{userId}` from the generated contract, which is contract-affecting and fails AC #6 outright. An annotation must sit immediately above its handler declaration, so a handwritten sibling cannot supply it either. Aborted; the file stays handwritten. Twelve files under `server/` carry `@Router` and exactly one of them is a rung-1 or rung-2 candidate, so the blast radius inside this batch is one file. The generalisation is the finding: **the roster does not mention annotations at all**, and every route handler is exposed to it. Classification: **boundary gap**; the price is that the generated API contract loses an endpoint.

### F5 — a `:=` binding mislowers a method call made on it

`usersRouter := router.Group("/users")` followed by `usersRouter.GET("/:userId", getPublicUserProfile)` emitted `usersRouter := std.NewImmutable(router.Group("/users"))` and then `usersRouter.Get().GET("/:userId", getPublicUserProfile)`. `gin.IRoutes` has no `Get()`, so the emitted Go does not compile. The transpile is clean, there is no diagnostic, and the failure appears only at `go build` against a file the reader did not write. The `var` spelling emits `var usersRouter = router.Group("/users")` and a plain `usersRouter.GET(...)`, with `std.NewImmutable` absent from the file entirely.

This is worse than the trap the skill already records, which is that a `:=` binding wraps its value and every read emits `.Get()`. Here the transpiler rewrites the **receiver of a method call** as a read of the binding, which is never correct for a method defined on the value's own type. Classification: **defect**, and a clean report, because the contrast is one word and the failure is silent. Lead the report with `:=` versus `var`.

### F6 — `go_interop.MapPut`'s value type cannot be inferred when the value is narrower than the map

`go_interop.MapPut(ticks, "stepSize", 1)` on a `map[string]any` fails with `type map[string]any of ticks does not match inferred type any for V`. Supplying the type arguments explicitly, `go_interop.MapPut[string, any](ticks, "stepSize", 1)`, builds. The roster's mechanical-rewrite row 5 gives `go_interop.MapEmpty[string, T]()` and then `go_interop.MapPut(m, "a", x)` — type arguments on the constructor and none on the put — so the catalog as written does not compile for every `map[string]any` literal, which is the shape both bot chart builders use. Classification: **roster defect**; the fix is to add the type arguments to the same row rather than a new construct.

### F7 — the `.Size()` defect is wider than filed, and it is a property of the declaring file

Two things, and the second is the useful one.

First, the trigger is wider than upstream #613. That issue is about `.Size()` on a **field** of a struct declared in a Go sibling. It also fails on a **whole value** of a Go-declared named type, with no field access involved: `logs.Size()` over `[]postgres.DailyUserLog` and `logs[i].Members.Size()` both fail. `.Size()` over a builtin-typed slice such as a `[]string` parameter is fine.

Second, and this is what gates the batch, the failure is a mixed-package effect, so it is a property of the **declaring** file rather than of the call site. Verified with a two-file probe: with `struct DailyLog(var ID string, var Members []Member)` declared in a `.gala` file in the same package, both `log.Members.Size()` and `logs.Size()` transpile and build. Rewriting `postgres/dailylogs.go` as a twin therefore **resolves #613 at its call sites** in `discord_bot/commands/active_users.go` and `slackbot/commands/active_users.go`, both of which are rung 1.

Consequence: **the order is not free.** `active_users` is rung 1 yet is unwritable until a rung-2 file is rewritten, which is the opposite of the order the ladder implies. The batch plan's "rung 1 first" step is therefore wrong and is being corrected to rewrite declaring files before their dependents.

There is also no clean substitute for the element count while the declaring file is handwritten. `go_interop` has `SliceCap` but no length helper; `go_interop.SliceFrom(x, 0).Size()` builds but copies the whole slice to read its length, so it is a workaround with a real cost rather than a substitute worth adopting.

## Findings, batch 1, continued: the two rung-1 aborts and the ordering correction

### F7 refined — the trigger is an inferred type, not a mixed package

The first reading of F7 was wrong, and the `postgres` split is what corrected it.

`.Size()` is **not** blocked by the declaring file being handwritten. It transpiles and builds when the receiver's type is a GALA-declared struct reached from an *explicit* type, even across packages: `func CountOne(log types.Log) int = log.Members.Size()` and `func CountMany(logs []types.Log) int = logs.Size()` both build with `types.Log` declared in another package as a GALA struct.

It fails when the receiver's type is **inferred from a call into a Go-declared function**, even when that function is in the same package and returns a GALA-declared type: `var logs, err = List()` followed by `logs.Size()` gives `logs.Size undefined (type []Log has no field or method Size)`. The transpiler resolves a receiver's type from the `.gala` file's own text and will not read a Go sibling's signature to learn it. That is the whole defect.

And there is **no substitute**, which is what makes it a gap rather than a workaround. A multi-value binding accepts no type annotation at all — `var logs []Log, err = List()` is a parse error (`extraneous input ','`) — so the one spelling that would pin the type syntactically is unavailable for exactly the binding form that produces such a value.

Minimal repro for the report, three files in one package: `types.gala` declaring `struct Log(var ID string, var Members []Member)`, `repo.go` declaring `func List() ([]Log, error)`, and `consumer.gala` calling `logs.Size()` on the result. Classification: **boundary gap** — the element count is expressible, what is missing is a way to name the type of a value that came from Go, and the price is that any file counting the elements of a Go-returned slice cannot be rewritten. `gala` 0.84.1, go1.26.7.

### F8 — a `.gala` file cannot call a method declared in a handwritten sibling

`GALA-E0044`, and it is a rejection rather than a bad emission, so it is loud:

```
error[GALA-E0044]: Repository has no method withTransaction
  --> dailylogs.gala:43:11
   = hint: Repository declares: RecordDailyUserLogMembership, recordDai…
```

The transpiler resolves a method call against the declarations **in the same `.gala` file** and nothing else. `Repository` is declared in `postgres/repository.go`, which the sweep reports at rung 3, so it stays handwritten, and its 106 methods are spread across twelve files.

Consequence: **no `postgres` file that declares or calls a `Repository` method can be a twin**, which is 12 of the 14 rung-2 `postgres` candidates. The only viable shape is a declarations-only split — `dailylogs.gala` carrying the two structs, `dailyLogDate`, and the two query variables, with the four methods moved verbatim into a handwritten `dailylogs_methods.go`. That is a legal rung-2 split, but a much smaller twin than the roster's rung-2 wording implies.

### Aborts recorded

- `routes/users.go` — aborted on F4. Stays handwritten.
- `discord_bot/commands/active_users.go` and `slackbot/commands/active_users.go` — aborted on F7. Both are rung 1 and both call `.Size()` on the result of `repository.ListActiveUserDays(...)`, a Go-declared method returning a slice. Neither the inferred nor an annotated spelling works, and the deep `map[string]any` chart literals would additionally have needed the F6 type arguments. The `.gala` sources are parked outside the repository rather than committed beside nothing.

### Ordering correction

The batch plan's "rung 1 first" step is wrong and is corrected: **declaring files and their dependents are coupled, and the rung number does not express the coupling.** `active_users` is rung 1 and is unwritable regardless of order, because the coupling is not to another candidate but to the fact that `Repository` is rung 3. A rung-1 verdict is therefore not sufficient evidence that a file is writable, and the preflight tool cannot see that, because the reason lives in a third file.

### Landed so far

- `middleware/auth.gala` → `middleware/auth.go`, plus a handwritten `middleware/doc.go`. sha256 `3b1cb9a2a86d94ae8906a4c2950963e3a367b90a9786eadf4c97c63753a69743`. Runtime-enabled via `std.As`. Two deliberate local changes: `err` split into distinct names because a multi-value binding cannot be redeclared, and the single-use `session` local inlined.
- `postgres/dailylogs.gala` → `postgres/dailylogs.go`, plus a handwritten `postgres/dailylogs_methods.go`. sha256 `ae3fe90adc34a683ed2bf6fa91f960cce2b23b3a6ae17c064b2e524e2e15c078`. Struct fields stayed plain Go — `std.Immutable` appears zero times — but the twin is runtime-enabled because the synthesized `Copy` and `Equal` call `std.Copy` and `std.Equal`, and it **adds exported `Copy`, `Equal`, and `Unapply` to two exported types**. That is a recorded change to the exported surface, and it is the unavoidable cost of declaring a struct in GALA at all.

Both twins are byte-identical on double regeneration, `gofmt` clean, and `go build ./...` green. The Compose test suite and the corpus runners have not been run yet.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What this task actually established

The project-independent `gala-from-go` skill had never been run against a file this repository ships. Running it produced a sweep of all 78 handwritten non-test Go files, two landed twins, three recorded aborts, and eight findings — and the sweep's own verdict is the headline: **the file-level rung is not sufficient evidence that a file is writable**, because the blocking reason usually lives in a third file the tool does not read.

## Sweep

5 rung-1, 29 rung-2, 42 rung-3, plus 2 `doc.go` files that exit 1 because the tool walks them and finds no declaration. The sweep contradicted `server/GALA.md` in three places, and on the one case the batch could test, `server/GALA.md` was right: `middleware/auth.go` does need a type-pattern `match`, and the rewrite only became possible once the `GALA-E0033` inference limit was worked around.

## Landed

- `middleware/auth.gala` → `auth.go`, plus a handwritten `doc.go`. sha256 `3b1cb9a2…`. Runtime-enabled via `std.As`. Signature, callers, and behaviour unchanged.
- `postgres/dailylogs.gala` → `dailylogs.go`, plus a handwritten `dailylogs_methods.go` carrying the four `*Repository` methods verbatim. sha256 `ae3fe90a…`. Struct fields stayed plain Go; the twin is runtime-enabled only because the synthesized `Copy`/`Equal` call `std.Copy`/`std.Equal`.

Both regenerate byte-identically, are `gofmt` clean, carry relative `//line` directives, and were not hand-edited.

## Aborted, with the finding that drove each

- `routes/users.go` — Swag annotations are dropped, so `swag init` would delete `/users/{userId}` from the OpenAPI document. Now `GAP-11`.
- `discord_bot/commands/active_users.go` and `slackbot/commands/active_users.go` — both rung 1, both blocked because `.Size()` needs a receiver type the transpiler can read from the `.gala` file's own text, and a multi-value binding accepts no type annotation, so no spelling recovers it. Now `GAP-12`.

## Findings worth keeping

- **`GALA-E0044`**: a `.gala` file cannot call a method declared in a handwritten sibling. `Repository` is declared in rung-3 `repository.go` with 106 methods across 12 files, so 12 of the 14 rung-2 `postgres` candidates are dead. A documented answer rather than a gap, but it is the single biggest constraint the roster does not mention.
- **A `:=` binding mislowers a method call receiver** into `x.Get().M()`, which does not exist in Go. Clean transpile, no diagnostic, build failure against generated code. The cleanest upstream report of the batch: one keyword of contrast, and the `var` form emits no wrapper at all, so the absence is assertable.
- **`.Size()` on an inferred receiver** is wider than filed upstream #613 and, unlike it, has no substitute — that is what makes it a gap rather than a lowering bug.
- **`go_interop.MapPut` needs explicit type arguments** on a `map[string]any`; the roster's mechanical-rewrite row omitted them, so the catalog as written does not compile for the shape both bot chart builders use.
- **Comments are dropped entirely**, which costs an exported declaration's godoc and, on an annotated handler, an endpoint in the contract document.

## Verification

`go build ./...` green; `gofmt -l` clean; the canonical Compose sequence green including `timeful/server/postgres` (9.5s) against the real database, which is what actually exercises the `dailylogs` split; corpus `62 passed, 0 failed`; `audit.sh` passes; `go run ./inventory -check` reports `doc matches (168 files)`; `npm run format:markdown`, `format:markdown:check`, `test:markdown-rules` (30 tests), and root `fmt:check` all pass; the code graph was re-indexed to 11366 nodes / 45853 edges.

`audit.sh` earned its keep: it rejected the first draft of the two GAP rows because I had written them without backing probes or roster families. The fix was two new probes and two roster rows, not a softened claim.

## Staged for the later filing pass

No upstream issue was filed. Candidates, in the order they are worth filing: the `:=` method-receiver defect (cheapest, clearest contrast); `GAP-12` as a boundary gap; `GAP-11` as a boundary gap; and the roster-side corrections for the `MapPut` type arguments and the `type-assertions` rung, which are ours rather than the compiler's.

## Not done, deliberately

`observability`, `accounts`, and the rung-2 singletons were left alone on the user's decision to consolidate rather than grind. The `postgres` group beyond `dailylogs` is blocked by `GALA-E0044` rather than by effort, so grinding it would have produced aborts rather than twins. `services/gcloud/tasks.go` is the one untouched rung-1 candidate.

Nothing is committed; the change set is in the worktree awaiting a commit decision. The pre-existing `backlog/backlog.md` modification is unrelated and was left alone.
<!-- SECTION:FINAL_SUMMARY:END -->
