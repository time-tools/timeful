---
id: TASK-0337
title: Reproduce the unreported GALA blockers minimally and file the upstream reports
status: Done
assignee: []
created_date: '2026-09-30 11:50'
updated_date: '2026-09-30 16:46'
labels: []
dependencies: []
priority: high
type: chore
ordinal: 343000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Outcome

Every GALA blocker that TASK-0336 recorded but did not file is re-derived on the pinned compiler with a minimal, self-contained reproduction, checked against upstream master for an existing fix or issue, classified as bug or feature request, and filed as an upstream report on `martianoff/gala`.

`martianoff` answered the roster in #528 on 2026-09-24 with a per-gap equivalent table and explicitly invited two own issues: "Doc comments in generated Go, and defined types for non-struct types. Both are fair asks and worth their own issues."
Neither was filed, and neither has an issue in the tracker, which carries only #528 and #611-#613 as open.

The maintainer also asked twice for a minimal mixed-package case for the two findings they could not reproduce (#528, 2026-09-23: "if you can share a minimal case with the sibling file included, I'll take another look").
TASK-0336 produced exactly that case and it is still unanswered.

## Requirements

### Reproduce before reporting

- Each candidate is a minimal, self-contained program under a scratch directory, not a file from this repository.
- Each reproduction states the pinned compiler (`gala version`), the Go toolchain, and the transpile and build outcome.
- A claim about emitted Go is settled by reading the emitted file; a claim about behaviour is settled by running it.
- A claim that two spellings differ asserts the absence as well as the presence, and the marker names a semantic thing rather than layout.
- A reproduction that does not reproduce on the pinned compiler is recorded as such and no issue is filed from it.

### Check upstream before drafting

- Search the tracker for an existing issue or an open pull request on the same construct, and read `martianoff`'s answer in #528 before classifying anything already answered as a gap.
- Check master for a fix landing after `0.84.1`, so a report is not filed against behaviour that is already fixed upstream.
- A construct #528 already gave an equivalent for is not refiled; it is recorded as answered.

### Candidate set

- A `:=` binding whose value has a method: the transpiler lowers the receiver into an unwrap read, so a clean transpile emits Go that does not build.
- Cross-boundary resolution: a `.gala` file resolving a method or a receiver type declared in a handwritten `.go` sibling of the same package. This is the minimal case #528 asked for, and it covers both `GALA-E0044` and the `.Size()` case that has no substitute.
- Comments: the transpiler emits none, so an exported declaration loses its godoc and a `swag` annotation loses a documented endpoint. Invited by #528.
- Defined types for non-struct types (`GALA-E0048` covers the method half; the type half is missing). Invited by #528.
- A resource combinator whose result type argument is omitted: the emitted function loses its return type instead of failing where the type argument was written.
- `go_interop.Spawn` accepting a capture that `concurrent.Future` refuses with `GALA-E0037`.
- A block-bodied lambda in return position: `GALA-E0033` names the return type as a context that permits inference, and it does not.
- `go_interop`: `MapLen` exists and no slice-length helper does, and `MapPut` cannot infer its value type on a `map[string]any`.

### Report

- One issue per construct, titled with the Go construct and the emitted artifact.
- Each body follows `.agents/skills/gala-from-go/references/gaps.md`'s report template: classification, what was tried, the exact diagnostic or the emitted Go, a minimal repro with its imports, what a caller needs, `gala version`, environment, and what was searched.
- A defect report leads with the contrast between the two spellings and asserts the absence.
- Each issue cross-references #528 and, where it extends one, #611, #612 or #613.
- No issue is filed for a construct #528 answered, and a report whose repro does not reproduce is withdrawn rather than softened.

## Constraints

- `gala` 0.84.1 and the Go 1.26 series stay the pinned toolchains, and `server/third_party/gala/` is not modified.
- Generated files are never hand-edited; a repro lives outside the repository or as a probe in the corpus.
- The local gap catalogue keeps its entries until the filing pass strips them, and each entry is updated with its issue number.
- Filing is outward-facing: every issue is drafted, reviewed against its reproduction, and only then filed.

## Evidence required

The reproduction, diagnostic, and build output per candidate; the upstream search result per candidate; the filed issue URL per candidate; and the reason each non-filed candidate was not filed, in the task notes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every candidate has a minimal self-contained reproduction re-derived on gala 0.84.1, with its transpile, emitted-Go, and build outcome recorded
- [x] #2 Every candidate is checked against upstream master and the issue tracker, and every construct #528 already answered is recorded as answered rather than refiled
- [x] #3 Each reproduction that fails to reproduce is recorded as such and produces no issue
- [x] #4 Each filed issue follows the gaps.md report template, leads a defect report with the two-spelling contrast, asserts absence as well as presence, and cross-references #528 or the issue it extends
- [x] #5 The two issues #528 explicitly invited are filed: comments in generated Go, and defined types for non-struct types
- [x] #6 The minimal mixed-package case martianoff asked for in #528 is supplied and linked from the relevant issue
- [x] #7 Every candidate is classified as exactly one of bug, feature request, documentation defect, or already-answered, with the reason
- [x] #8 The local gap catalogue entries that become filed issues are updated with their issue numbers and the local documents stay consistent
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Filed upstream, 2026-09-30

Eight issues and one comment, all re-derived on `gala` 0.84.1 / go1.26.7 with minimal two-file repros. Drafts are kept outside the repository at `/tmp/nix-shell.SfTLK3/opencode/gala-repro/drafts/`; the two headline reproductions were re-run by hand before filing.

| Issue | Class | Construct |
| --- | --- | --- |
| #614 | bug | `:=` / `val` binding: the receiver of a method call is lowered into an unwrap (`x.Get().Rename(...)`), and `std.AddrOfCopy` plus `GALA-E0053` are skipped when the receiver's type is declared in a same-package `.go` sibling. Clean transpile, `cannot call pointer method` at build. `var` emits no wrapper at all. |
| #616 | bug | A bare name in a declared-type position resolves to an imported package over a package-local declaration. #528's second "couldn't reproduce" item, and it needs no handwritten sibling: two `.gala` files. Silent signature substitution when the two types share a field. |
| #615 | bug | False-positive `GALA-E0044`: a GALA-declared type whose method is declared in a handwritten `.go` sibling. The hint (`Repo declares no methods`) is false and contradicts the `GALA-E0044` page's own stand-down text. |
| #618 | bug | `resource.Using` / `Bracket` degrade the body lambda's parameter to `any` when the resource type is Go-declared; a partial type-argument list emits the transpiler's own `A` / `R` as a Go type. |
| #617 | bug | `func F() = <expr>` never infers a return type: the emitted Go function is void and its body returns a value. |
| #619 | feature request | Comments in generated Go. Maintainer-invited in #528. 3,156 authored comment lines across 53 vendored `.gala` files, 0 in the 53 generated `.go` files; `gala doc` has the prose and `go doc` does not; `ast.CommentGroup` appears 0 times in the transpiler. Includes the `swag` annotation consequence and the `gala-lint` "Undocumented export" contradiction. |
| #621 | feature request | No newtype for non-struct types. Maintainer-invited in #528. Also refutes one part of the #528 answer: `Codec[T]` cannot restore a scalar wire shape, because `StructMeta.EncodeFields` hard-codes the object braces (`.Rename("Value", "")` yields `{}`, not `1500`). |
| #620 | bug | `var (a, b) = <tuple>` is accepted by the grammar and panics with `GALA-E0017`; the `val` twin transpiles. |
| #613 comment | — | The substitute `go_interop.SliceFrom(x, 0).Size()` works at 0 allocs, so the price is documentation rather than impossibility; `.ByteSize()` fails identically; a second trigger (a local binding, not field access) reproduces; `go_interop` has `SliceCap` and `MapLen` and no slice-length or string-length helper. |

Deliberately not filed: `go_interop.Spawn`'s missing unchecked-capture caveat in `GALA_BEST_PRACTICES.MD:14,:105` and `GALA.MD:3176`. It is a documentation defect, but the asymmetry is already documented in four other places.

## Refuted, so nothing was filed

Three of this repository's recorded claims did not survive a minimal reproduction, and each was wrong in the direction of over-claiming:

- **The `GALA-E0033` hint is correct.** A block-bodied lambda in return position with an unannotated parameter transpiles, builds and runs; all three contexts the hint names work, including the return. Only a bare untyped `val` and a generic type parameter are refused, and both refusals are right.
- **`go_interop.MapPut` infers.** Eight shapes over `map[string]any`, `map[string]string`, `map[any]any`, a map-typed parameter, a map returned from a Go sibling, a Go-declared struct field and both binding keywords: all transpile, build and run. The roster row that omitted the type arguments was wrong.
- **GAP-12 is a workaround, not a boundary gap.** `go_interop.SliceFrom(logs, 0).Size()` lowers to `len(go_interop.SliceFrom(logs, 0))`, builds, runs, and allocates nothing (`SliceFrom` is `s[from:]`, O(1)). A multi-value binding still takes no type annotation, but a single-value binding does, and the workaround exists for every receiver kind except `string`.

## Local documents corrected, 2026-09-30

Thirteen new probes pin the new claims and the refutations, and four documents were corrected against them.

Probes: `blocked_receiver_unwrap_samepkg`, `blocked_receiver_unwrap_val`, `emit_var_receiver_no_wrapper`, `blocked_bare_name_type_position`, `blocked_e0044_go_sibling_method`, `blocked_resource_go_sibling_type`, `blocked_expression_body_no_return`, `blocked_var_tuple_pattern`, `pass_lambda_return_inference`, `pass_map_put_inference`, `pass_slice_from_size_len`, `pass_expression_body_typed`, and `pass_var_tuple_val_form`. The corpus was 75 probes after this pass, and every new `build_fail` and `transpile_fail` probe reproduces the diagnostic quoted in its filed issue on the pinned compiler.

Corrections:
- `references/constructs.md:76` — the block-bodied-lambda row was refuted by its own verification instruction; the return position is `direct` and a new row carries the two genuine `GALA-E0033` refusals. The inferred-receiver `len` row moves from no-substitute to a `workaround` naming `go_interop.SliceFrom(xs, 0).Size()`. The resource rows lose three wrong claims: the result type argument is optional, `WithLock` is unaffected, and the enclosing return type is never stripped. Four new rows cover the filed package-level and binding defects.
- `references/gaps.md` — the `.Size()` entry leaves "Gaps Worth Filing" and moves to "Constructs That Were Never Gaps" with the substitute and where it is established; the block-bodied-lambda entry moves there too; the `:=` receiver entry gains its two corrections and its issue number; five new filed entries replace the "not filed" text; every "What Not To File" bullet that would have invited a duplicate is now explicit.
- `docs/gala-translation.md` — GAP-12 is deleted and the row reclassified as a workaround, so GAP-1 to GAP-11 stay contiguous; the roster gains `binding-method-receivers`, `var-tuple-destructuring`, `expression-body-no-return`, `gala-type-go-sibling-method`, and `bare-type-name-resolution`; five mechanical-rewrite rows and one no-rewrite index row are added; the `MapPut` row no longer claims required type arguments; seven new subsections in "Workarounds and contested verdicts" carry the reasoning, including the `SliceFrom` substitute and the `pass_lambda_return_inference` refutation.
- `server/GALA.md` — the four known-limit bullets now name their issue and carry the corrections, and the revisit list drops GAP-12 for the eight filed reports.

Verification for this pass: `run.sh` 75 passed 0 failed; `audit.sh` passes all six checks; `go run ./inventory -check` reports `doc matches (168 files)`; the inventory unit tests pass; `npm run format:markdown`, `format:markdown:check`, `test:markdown-rules` (30 tests), and root `fmt:check` all pass; `go build ./...` in `server/` is clean. No Go source in the server module changed, so the code graph does not need refreshing and no swagger regeneration applies.

## Definition of Done items that do not apply, and why

Three checklist items are satisfied by not applying rather than by running a command, and the reasoning is recorded so a later reader does not read the tick as a run that happened.

- **#3 e2e** — the change set is Markdown plus the probe corpus, which is a separate Go module (`server/scripts/20260923_gala_translation_probes/go.mod`) with no frontend or server-runtime artifact in it, so the browser suite has nothing to exercise. The executable suite for this area is the corpus itself, and it passed 75 of 75 after this pass.
- **#5 Swagger** — no handler, route, or annotation changed. The annotation finding is recorded as an upstream gap precisely because no annotated file can be a twin, so `routes/users.go` is untouched and the contract document is unaffected.
- **#6 code graph** — no Go source in the indexed server module changed. The new corpus files are `.gala` sources, `expect` markers, `expected.out`, and runner-written `.go` fixtures that are gitignored, and `run.sh` is a shell script, so a re-index would record nothing new.

## Follow-up TASK-0337.01, 2026-09-30

A review of this change set found the substance sound but the documents themselves left inconsistent: superseded text left beside its replacement, a table row that lost its cells, sentences cut mid-clause, and roster citations that named a line past the end of a file or the remedy rather than the defect.
Every check in the Definition of Done passed on that state, because none of them compares a table row against its header, reads a sentence for terminal punctuation, or resolves a prose pointer to a file, so an earlier draft of these notes carried a "Local documents that are now known to be wrong" section listing work as outstanding; that section is gone because the work it listed is what this task's own notes record as done above, and the inconsistency it could not see was repaired by TASK-0337.01.
That follow-up repaired the documents and added the seventh `audit.sh` check, which compares every Markdown table row in the five GALA documents against its own header.
The corpus is 77 probes after the follow-up rather than the 75 recorded above, and the two it added are `emit_comments_dropped`, which pins the block comment syntax that parses and is then dropped so GAP-11 is an emit gap rather than a grammar gap, and `blocked_bare_name_type_gala_sibling`, which pins the sibling-`.gala` layout #616 was filed against and that this corpus did not cover.
Nothing upstream changed in the follow-up: the eight issues #614 through #621 and the #613 comment are exactly as filed.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: review
created: 2026-09-30 13:25
---
## Review of the change set, 2026-09-30

Reviewed the staged change set against the pinned toolchain and the upstream tracker, and left the worktree unmodified.

**Substantively sound.** All eight issues #614 through #621 exist on `martianoff/gala`, filed 2026-09-30, and the #613 comment exists and supplies the minimal mixed-package case the maintainer asked for in #528, so acceptance criterion #6 holds. `run.sh` passes 75 of 75 on `gala` 0.84.1 / go1.26.7, `audit.sh` passes all six checks, `go run ./inventory -check` reports no drift across 168 files, the inventory unit tests pass, and `format:markdown:check`, `test:markdown-rules` and root `fmt:check` are green. The three refutations recorded in the notes all re-derived correctly on the pinned compiler. The verification recorded in the notes is real.

**Not consistent, which acceptance criterion #8 requires.** Twelve defects, four of them mechanical damage from a replacement that was inserted without deleting the old text: four bullets in `server/GALA.md` carry their new body followed by the superseded body, three of them contradicting themselves; `server/GALA.md` has an empty bullet and a sentence stranded at the end of an unrelated bullet; one table row in `constructs.md` lost all its cells; two sentences in `constructs.md` stop mid-clause. The rest are unbacked claims, including a pointer to evidence that lives only in a temporary directory, a roster citation to a line number past the end of the file, and a representative occurrence that names the remedy rather than the defect.

**The operational finding is the one that matters for the follow-up:** every automated check passes on the broken state, because none of them compares a table row against its header, reads a sentence for terminal punctuation, or resolves a prose pointer to a file. The checks in the Definition of Done cannot catch this class of defect, so the follow-up adds the table-row check rather than relying on re-reading.

**Deliberately not actioned here.** Committing and task-state changes are left to the user, and the eight filed issues are outward-facing and were treated as correct immutable inputs; nothing upstream was filed, edited, or withdrawn. `backlog/backlog.md` carries an unrelated pre-existing worktree edit that was left alone.

Follow-up work is recorded as the sub-task TASK-0337.01, which carries the full defect inventory, the constraints, and the verification commands.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What landed

Eight upstream issues on `martianoff/gala`, one substantial comment on #613, and the local documents corrected so no file still claims something a reproduction has since refuted.
A follow-up, TASK-0337.01, then repaired the documents this pass left mechanically inconsistent and added the check that catches that class of defect; it is summarised at the end.

## Filed

Six bugs, each with a two-file minimal reproduction re-derived on `gala` 0.84.1 / go1.26.7, every `build_fail` and `transpile_fail` case now pinned by a probe in the corpus:

- **#614** — a method call on a `:=` or `val` binding is lowered into an unwrap of the receiver (`x.Get().Rename(...)`). `std.AddrOfCopy` and `GALA-E0053` are both skipped when the receiver's type is declared in a same-package `.go` sibling, which makes the failure silent; the same source with the type imported emits `AddrOfCopy` and builds.
- **#616** — a bare name in a declared-type position resolves to an imported package over a package-local declaration. This is the second of #528's two "couldn't reproduce" findings, and it needs no handwritten Go at all: two `.gala` files. When the two types share a field name, both the transpile and the build succeed with the wrong exported signature.
- **#615** — false-positive `GALA-E0044` for a GALA-declared type whose method is in a `.go` sibling; the hint claims the type declares no methods, and the `GALA-E0044` page asserts the opposite of what happens.
- **#618** — `resource.Using`/`Bracket` degrade the body lambda's parameter to `any` when the resource type is Go-declared, and a partial type-argument list emits the transpiler's own `A`/`R` as a Go type.
- **#617** — `func F() = <expr>` never infers a result type, so the emitted Go function is void and returns a value.
- **#620** — `var (a, b) = <tuple>` is accepted by the grammar and panics with `GALA-E0017`, where the `val` twin transpiles.

Two feature requests the maintainer explicitly invited in #528 and that nobody had filed:

- **#619** comments in generated Go — 3,156 authored comment lines across the 53 vendored `.gala` files, 0 in the 53 generated `.go` files, `ast.CommentGroup` absent from the transpiler entirely, `gala doc` holding the prose that `go doc` lacks, a `swag` annotation deleting an endpoint from the OpenAPI document, and the project's own lint rule mandating comments the emitter deletes.
- **#621** newtype over a non-struct type — with evidence that #528's "the wire shape belongs to the codec" answer does not extend to a scalar wrapper, because `StructMeta.EncodeFields` hard-codes the object braces.

On **#613** the comment supplies the minimal mixed-package case the maintainer asked for, the working zero-allocation substitute that changes the price from "cannot be written" to "written down nowhere", `.ByteSize()` failing identically, and a second trigger that is a local binding rather than field access.

## Refuted, and corrected locally

Three recorded claims did not survive reproduction, each wrong in the direction of over-claiming: the `GALA-E0033` return-position claim (the hint is right; our row was wrong), `go_interop.MapPut` type-argument inference (it infers), and GAP-12 (a workaround, not a boundary gap — the vendored `go_interop.SliceFrom` is a one-line reslice at `server/third_party/gala/go_interop/types.go:116`, so the substitute allocates nothing, and nothing in this repository benchmarks allocations).

## Local state

15 new probes, 77 passing. GAP-12 deleted, five roster families added, five rewrite rows and one index row added, and `constructs.md`, `gaps.md`, `SKILL.md`, `docs/gala-translation.md` and `server/GALA.md` corrected with their issue numbers.

`run.sh` 77/77; `audit.sh` passes all seven checks; `go run ./inventory -check` reports no drift; inventory unit tests pass; markdown format and rule checks pass; root `fmt:check` passes; `go build ./...` in `server/` clean. No Go source in the server module changed, so the code graph and Swagger are untouched.

## Follow-up TASK-0337.01

A review found the substance sound but the edit pass had left superseded text beside its replacement, one table row with no cells, two sentences cut mid-clause, one pointer to a temporary directory, and roster citations that named a line past the end of a file or the remedy rather than the defect.
Every Definition of Done check passed on that state, because none of them compares a table row against its header, reads a sentence for terminal punctuation, or resolves a prose pointer to a file.
The follow-up repaired the five documents and the runner, added two probes (`emit_comments_dropped`, which pins that the block comment syntax parses and is then dropped, and `blocked_bare_name_type_gala_sibling`, which pins the sibling-`.gala` layout #616 was filed against), and gave `audit.sh` a seventh check that compares every Markdown table row in the five GALA documents against its own header.

Nothing is committed; the change set is in the worktree, and the pre-existing `backlog/backlog.md` modification is unrelated and was left alone.
The follow-up changed no upstream state: the eight issues #614 through #621 and the #613 comment are exactly as filed.
<!-- SECTION:FINAL_SUMMARY:END -->
