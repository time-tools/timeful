# Gaps And How To Report Them

A gap is a Go construct GALA cannot express, or cannot express in a shape a Go caller can use.
This file owns two things: the vocabulary for classifying a gap, and the shape of a report worth filing.

Read [`constructs.md`](constructs.md) first.
It says what each construct does today, and a construct with a `workaround` row is not a gap.

## Where To Look Before You Decide

Four upstream places carry the answers, and consulting them is cheaper than filing something already answered.

- The language specification, whose section numbers the rows in [`constructs.md`](constructs.md) cite for a refusal.
- The best-practices document, for the intended spelling where a construct has more than one.
- The generated documentation index, which is a digest of both and is the fastest place to find a section number you do not have.
- `gala explain --list` and the page for one code, which is the compiler's own account of itself and outranks every document for what the compiler currently does.

One upstream page is worth naming so you do not cite it: the Go-interoperability feature page claims that bare `len`, `make`, and `cap` work, and the specification and the compiler both contradict it.
A report that leans on that page gets corrected against you, so treat it as evidence of intent rather than as a description of behavior.

## Classify Before Reporting

A row is a gap only if it has no substitute.
Classify it into exactly one of these, because the class determines whether there is anything to file at all.
The first four are the classes [`constructs.md`](constructs.md) records as a `Status`; the last two are the ones that belong in a report, and neither has a `Status` because neither is a property of a single row.

| Class             | Status in `constructs.md` | Meaning                                                                                       | File it? |
| ----------------- | ------------------------- | --------------------------------------------------------------------------------------------- | -------- |
| direct            | `direct`                  | The construct is written as it stands; no substitution is involved.                             | No       |
| workaround        | `workaround`              | Only an interop helper or a plain-Go spelling preserves the behavior.                           | No       |
| analog            | `answered`                | A GALA-native construct expresses the same behavior, such as `match`, `Try`, `use`, or `.ByteSize()`. | No    |
| documented answer | `answered`, `semantic`    | The construct is refused, and either a diagnostic or the specification names the replacement.   | No       |
| boundary gap      | —                         | GALA expresses the behavior, but not as a Go-shaped drop-in, so a Go API or a wire format changes. | Yes  |
| language gap      | —                         | GALA has no syntax or no semantics for the construct, so nothing in the language expresses it.     | Yes      |

A construct whose handling is wrong rather than missing is a fourth thing, recorded in `constructs.md` as `defect` and listed under [Defects Observed While Pinning Rows](#defects-observed-while-pinning-rows).
It is a report, but a report about the compiler rather than about the language, which is why it is not a gap class.

The two reportable classes differ in what a fix looks like.
A language gap needs new grammar or new semantics.
A boundary gap needs an interop escape hatch: a way to produce or accept the Go shape, even if the GALA side is a different shape.
A boundary gap is real even though the behavior is expressible, because the thing that breaks is a caller on the other side of the boundary, and a caller cannot be changed from inside the transpiler.

A substitute can carry a cost of its own, and the cost belongs in the report.
Naming a struct where the Go code had an anonymous one is a substitution, and it can change the element type's identity, raise the Go language version the emitted code needs, or drop a promotion a caller relied on.
"Has a substitute" and "costs nothing" are different claims, so a boundary gap whose substitute has a price is still a boundary gap, and the price is the part the maintainer cannot infer.
Record what the substitute changed, not only that one exists.

`documented answer` is the class most often mistaken for a gap, because a parse error and a `GALA-Exxxx` rejection both look like a wall.
Run `gala explain GALA-Exxxx` and read the hint before classifying anything that has a code: the hint is the answer, and a report that restates it is a duplicate.
A parse error has no hint, so for those the question is whether the specification or the standard library already names a replacement, which is the question the `answered` rows in [`constructs.md`](constructs.md) record.

Do not promote a `workaround` to a gap because the workaround is ugly.
A helper that compiles and preserves the behavior is an answer, and a report that ignores it is answered with "use the helper".
Do not promote a `documented answer` to a gap because the answer is inconvenient, and do not demote a boundary gap to a workaround because the repository has a local convention for living with it; a local convention is evidence about impact, not about whether the gap exists.

## Gaps Worth Filing

There is very little here, and that is the finding rather than an absence.

Of the Go constructs GALA refuses, almost every one is a documented answer rather than a gap: the construct is rejected, and either a diagnostic names the replacement or the specification states the rule.
Those live as `answered` and `semantic` rows in [`constructs.md`](constructs.md), which is also where the replacement spelling is, so there is no reason to list them a second time.

A construct reaches this section only if it has no substitute at all.

| Construct | What happens | What a fix needs |
| --------- | ------------ | ---------------- |
| `const`   | parse error; the keyword is not in the grammar | a const declaration whose value stays a Go compile-time constant |
| a comment on a declaration | the transpiler emits no comments, so a doc comment or a `swag` annotation written in the `.gala` source is absent from the emitted Go | a comment-preserving emit, or an interop escape hatch letting a generated file carry a Go-facing comment |

`const` is the one construct here with nothing standing in for it.
`var` and `val` cover every use except a value that has to remain a Go compile-time constant, which is the whole of the gap, and it is a small one.
A file whose only `const` is a named number transliterates cleanly; report it only when a constant has to stay constant for a Go reader or a wire format.

A comment on a declaration is the second entry, and it is a boundary gap rather than a defect because both places it hurts are on the Go side of the boundary, which a transpiler cannot change.
The two places are far apart.
A doc comment on an exported declaration disappears from `go doc`, and it cannot be rescued by a handwritten sibling, because the only spelling is a bodiless re-declaration and that is a Go error.
A `swag` annotation disappears from the generated OpenAPI document, and it cannot be rescued either, because an annotation has to sit immediately above its declaration; a file that loses one silently drops a documented endpoint from the contract.
The substitute is to keep the comment in the `.gala` source, which is where it stays readable to whoever edits that source and invisible to every Go tool.
A *package* comment is the one exception to "unrescuable", because it needs no adjacency: a handwritten `doc.go` carries it and `go doc` picks it up.
Filed as upstream [#619](https://github.com/martianoff/gala/issues/619), after the maintainer named it in #528 as one of two genuinely missing capabilities.

Only the comment entry has been filed, as upstream [#619](https://github.com/martianoff/gala/issues/619); the `const` entry is deliberately unfiled, because a value that has to stay a Go compile-time constant is a rare shape and the row above already names the substitute for every other use.
Both entries were reproduced on a two-file minimal case, and the reproductions live in this repository's probe corpus rather than here: [`blocked_annotation_drop`](../../../server/scripts/20260923_gala_translation_probes/probes/blocked_annotation_drop/) pins the dropped annotation and [`blocked_const`](../../../server/scripts/20260923_gala_translation_probes/probes/blocked_const/) pins the parse error.

## Constructs That Were Never Gaps

These are listed because each one was recorded as blocked before it was checked against the language or the corpus, and the reasoning is what stops them coming back.

A Go function literal inside a composite literal is a parse error in the Go spelling only.
A GALA lambda in that position transpiles and builds, so the construct was never missing and there was nothing to file.

Assignment to a `:=` binding is a documented semantic, not a gap.
The language specification states that `:=` bindings are immutable, and the rejection is that rule working as written; `var` is the mutable form and the migration is mechanical.
A construct the language deliberately refuses is a fact about the language, not a gap in it.

A block-bodied lambda in return position infers its parameter type from the written return type.
`func F() func(int) int = (x) => { return x + 1 }` transpiles, builds, runs, and emits `func(x int) int`; the same holds for a return type written through an alias, on a method, and for a multi-parameter function type.
`GALA-E0033`'s hint names three contexts that permit an unannotated parameter — a typed `val`, a function argument, and a return — and all three hold.
The two shapes that are refused are a bare untyped `val` initializer, where there is no type to draw from, and a generic type parameter such as `func F[T any]() T = (x) => x`, where the concrete type is not yet known; both refusals are correct.
Read the hint, then run the case the hint names, before calling any of them a defect.

`go_interop.MapPut` infers its type arguments from the map argument.
`go_interop.MapPut(m, "stepSize", 1)` over a `map[string]any` transpiles, builds and runs, and so do `map[string]string`, `map[any]any`, a map-typed parameter, a map returned from a Go sibling, a Go-declared struct field, and both binding keywords.
The explicit `MapPut[K, V](m, k, v)` form also builds, so it is optional rather than required, and a helper that documents no inference rule is not a report.

`.Size()` on a receiver whose type was inferred from a Go call has a substitute, so it is a workaround rather than a gap.
The transpiler types a receiver from the `.gala` file's own text and does not read a Go sibling's signature, so the call is emitted against a type it never resolved:

```gala
var logs, err = List()   // List is declared in repo.go and returns []Log
return logs.Size()        // logs.Size undefined (type []Log has no field or method Size)
```

A receiver whose type is *written out* resolves even across packages, and the inferred case is reachable without naming the type at all:

```gala
return go_interop.SliceFrom(logs, 0).Size()   // builds; emits len(go_interop.SliceFrom(logs, 0))
```

The substitute allocates nothing, and what establishes that is a source line rather than a benchmark: the vendored helper is a one-line reslice at [`go_interop/types.go:116`](../../../server/third_party/gala/go_interop/types.go), whose own doc comment calls it `O(1)`, so the earlier reading of it as a whole-slice copy was wrong.
A multi-value binding still takes no type annotation, because the grammar's single `(type)?` slot sits after the whole name list, but a single-value binding does take one, so the annotation is not the only door.
The lowering defect itself is a report, filed as upstream [#613](https://github.com/martianoff/gala/issues/613); what is not a report is the conclusion that the element count cannot be read.

A construct the transpiler accepts and then lowers wrongly is none of these.
It is a defect, and it belongs under [Defects Observed While Pinning Rows](#defects-observed-while-pinning-rows).

## Defects Observed While Pinning Rows

These are not gaps.
Each is a construct GALA is supposed to handle, where the handling is wrong, so each is an issue rather than a request, and none of them is a reason to leave the file handwritten without saying so.
Every entry below is filed, and the contrast in each one is what makes the fix legible, so a new report on the same construct should lead with the contrast rather than the repro.

Two things make a defect report land, and both are about the evidence rather than the diagnosis.

Lead with the artifact the transpiler itself generated that states the correct shape.
The compiler emits the declaration and the extraction it later mislowers, so its own output contains the specification of what it should have done; a report that quotes that line needs no argument about intent, because the contradiction is inside one file.
A reader who has not seen your code can then check the claim against the compiler rather than against your reasoning.

Choose your assertion to match the claim, and notice that the two directions are not interchangeable.
A check that a marker is *present* pins a shape the compiler produced, so it cannot back a claim that a construct is blocked; a check that a marker is *absent* pins the shape's absence, and a construct is blocked exactly when the thing it needs is not emitted.
The same asymmetry is why a program that prints the right answer proves nothing about whether a wrapper is there.
And a report needs a probe that fails for the construct it names: a construct the corpus cannot reproduce is a claim about a compiler you no longer have.

**`fallthrough` inside a `match` arm produces an internal error instead of a diagnostic.** [#611](https://github.com/martianoff/gala/issues/611).
A bare `fallthrough` statement is correctly rejected as a forbidden statement keyword, but the same word inside a `match` arm passes the parser, reaches codegen, and produces Go that does not parse, so the transpiler reports an internal transpile error.
The transpiler writes that unparseable output to a file and names it in the error, so the report carries its own evidence; include the file's contents rather than describing them.

**A `match` on a struct whose fields are declared `var` emits an unwrap the field does not have.** [#612](https://github.com/martianoff/gala/issues/612).
The struct's synthesized extractor is called, and the emitted read of a `var` field carries a `.Get()`; the same match over the same fields declared without `var` builds and runs.
The transpiler generates both the struct and the extractor, so the generated `Unapply` return type already states whether the field is wrapped; the match lowering ignores it and unwraps unconditionally.
Report both spellings, because the contrast is what makes the bug legible.

**`.Size()` on a field of a struct declared in a handwritten Go sibling does not build.** [#613](https://github.com/martianoff/gala/issues/613), filed as a documentation and diagnostics gap rather than a bug.
The transpiler cannot know the field's type, so it emits a method call on a plain Go type that has no such method.
The same call on a GALA-declared struct lowers correctly, so the trigger is a mixed package rather than the construct itself, and it holds for a plain `string` field as much as for a slice.
Three facts since the filing, all checked on 0.84.1: `.ByteSize()` fails on the same shapes, so both documented replacements break together; the failure also fires for a **local binding** whose type came from a same-package Go function, not only for field access; and `go_interop.SliceFrom(x, 0).Size()` builds and runs on every shape except a `string`, at zero allocations.
So the price is that the substitute is written down nowhere, not that the length cannot be read.
It is also worth knowing as a limitation when deciding whether a member can move into a `.gala` file beside a `.go` sibling.

**A method call on a `:=` or `val` binding is lowered into an unwrap of the receiver.** [#614](https://github.com/martianoff/gala/issues/614).

```gala
var usersRouter = router.Group("/users")
usersRouter.GET("/:userId", getPublicUserProfile)   // emits usersRouter.GET(...), builds
```

```gala
usersRouter := router.Group("/users")
usersRouter.GET("/:userId", getPublicUserProfile)   // emits usersRouter.Get().GET(...), does not build
```

The transpile is clean, there is no diagnostic, and the failure is a `go build` error against generated code the reader did not write.
The `var` form emits no `std.NewImmutable` anywhere in the file, and that absence is what distinguishes the two lowerings rather than merely the presence of a wrapper in one of them.

Three corrections to the earlier reading of this, each with a probe behind it.
`val` emits the same artifact as `:=`, so `var` is the only safe spelling of the three rather than one of two.
The build failure needs a **pointer**-receiver method: a value-receiver method on the unwrapped copy builds and runs, so the same source is either a build failure or a quiet copy.
And the trigger is narrower than it looks — the transpiler emits `std.AddrOfCopy(x.Get()).M()` for a GALA-declared receiver and for an imported one, and skips it only for a type declared in a same-package handwritten `.go` sibling, which is also where `GALA-E0053` is suppressed.

**`GALA-E0044` rejects a method declared in a handwritten `.go` sibling, and its hint is false.** [#615](https://github.com/martianoff/gala/issues/615).
The receiver's method set is read from the file being transpiled only, so a `struct` declared in the `.gala` file is judged against a method set that excludes a sibling's.
The refusal is a false positive: the emitted Go would compile, and changing the sibling's extension from `.go` to `.gala` makes the identical declarations transpile and build.
The hint reads `Repo declares no methods`, which is untrue, and the `GALA-E0044` page's own stand-down text asserts the opposite of what happens — "a type declared in the package being compiled is fully known and is still checked".
The check fires exactly where its metadata is weakest, which is the scope a fix needs.

**A bare name in a declared-type position resolves to an imported package.** [#616](https://github.com/martianoff/gala/issues/616).
A return type, a parameter type, or a `var` annotation spelled as a bare name takes an imported package's qualifier whenever any import of the file exports that name, while a constructor position resolves to the package-local declaration.
Two files of `.gala` are enough; a handwritten sibling is not required, which also corrects the guess that this needs a mixed package.
When the two types share a field name, `gala transpile` and `go build` both succeed and the exported signature is the wrong one, so the evidence is a field that exists only on the local type.

**`func F() = <expr>` never infers a result type.** [#617](https://github.com/martianoff/gala/issues/617).
The emitted Go function is void and its body returns a value, so `go build` reports `too many return values` against generated code; `func F() T = <expr>` is correct.
A related earlier claim is wrong and worth keeping in mind: omitting a resource combinator's result type argument does *not* strip the enclosing function's return type, in any of the eleven variants checked.

**`resource.Using` and `resource.Bracket` degrade the body lambda's parameter to `any`.** [#618](https://github.com/martianoff/gala/issues/618).
A GALA-declared resource type infers both type arguments, which is why upstream's own example calls `Using(Handle(...), (h) => ...)` with none written out, and `WithLock` infers either way.
With the resource's type declared in a handwritten `.go` sibling, the transpile is clean and the emitted Go does not build, and a *partial* type-argument list emits the transpiler's own parameter name as a declared Go type.
PR #606 already made this shape a `GALA-E0033`; here the expected parameter type resolves to nil rather than unusable, so the guard does not fire.

**`var (a, b) = <tuple>` panics the transpiler.** [#620](https://github.com/martianoff/gala/issues/620).
`varDeclaration` accepts the grammar's `tuplePattern` and the transformer asserts an `IdentifierListContext` that is nil, so the call ends in a `GALA-E0017` internal error while the `val` spelling of the identical program transpiles, builds and runs.
A construct the grammar accepts should not reach an internal error, and the hint names no line the author can fix.


## Report Template

The upstream repository triages Go constructs GALA does not have in one open issue, and a language gap belongs there rather than in a new issue.
That issue is not a general defect queue: a construct it has already answered is answered, and the pull request that closed it named several more as having documented replacements.
A defect is a different kind of thing from a gap, and it wants its own issue, because the fix is a change in the compiler rather than a request for a feature.
Check the tracker for an existing issue on the construct before drafting either, and open a new one only when the search comes back empty.

Search before drafting, not before filing.
The search decides whether a construct is a new issue, a comment on an existing one, or already fixed, and drafting first spends the contrast framing that the report depends on.

A report is worth reading when someone who has not seen the code can reproduce it and tell whether it still happens.
Fill every field; a field you could not fill is itself information, so write `unknown` rather than dropping it.

```markdown
### <the Go construct, in Go>

<one or two lines: what it is, and where it appears in real code rather than only in a repro>

**Classification:** language gap | boundary gap

**What I tried**

<the GALA spelling that comes closest, and what it does instead>

**Diagnostic**

<the exact code and message, or the parser's expectation list when there is no code,
or the path to the unparseable output when the transpiler reported an internal error>

**Minimal repro**

<the smallest complete program, with its import lines, that shows it>

**What a caller would need**

<for a boundary gap: the Go-shaped signature, wire format, or field that has to keep working,
and who breaks when it changes>

**Compiler**

<`gala version` output>

**Environment**

<Go toolchain, and whether the program builds and runs when the construct is worked around by hand>

**Searched**

<`gala explain --list`, the specification and best-practices sections you read,
the generated documentation index, and the upstream triage issue>
```

Two habits make a report land.

Reproduce minimally, then say where the real code hits it, because a repro with no real occurrence is a feature request and a repro with an occurrence is a blocker, and the maintainer triages those differently.
Check `gala explain --list` before filing, because a construct that is forbidden on purpose has a code and a hint and belongs in a rewrite guide rather than in an issue.

## What Not To File

Do not file these, and do not spend a report slot on them.

The first four are the ones that look like walls and are not; each has already been filed and answered upstream, and re-filing one spends the credibility a real gap needs.

- A construct with a `workaround` row in [`constructs.md`](constructs.md), which is where the replacement spelling is.
- A construct with an `answered` row there: `switch`, struct tags, embedded fields, anonymous and `struct{}` types, fixed-size arrays, type assertions, `interface{}`, `select`, `goto`, a method on a defined type, the `if` initializer, `recover`, multi-value returns, channel types, grouped parameters, and nested functions.
- A construct with a `semantic` row there, such as assigning to a `:=` binding.
- A construct whose `gala explain GALA-Exxxx` page already states the fix, which is a documented answer rather than a gap.
- A construct re-derived on the compiler you have and found working, which this file lists under [Constructs That Were Never Gaps](#constructs-that-were-never-gaps): a block-bodied lambda in a typed return position, and `go_interop.MapPut` type-argument inference.
- A receiver typed from a Go call, whose substitute costs nothing because the vendored `go_interop.SliceFrom` is a one-line reslice; report the lowering defect instead, and the defect is filed.

Then the ones that are not about a construct at all:

- A style preference, such as preferring a GALA collection to a Go slice where the value only ever meets a Go API.
- A row you have not transpiled on the compiler you have.
- A failure that a transpile-and-run on a minimal repro does not reproduce, which is a signal to look for a context effect in the real file rather than to report the isolated result.
- A `GALA-E0017` internal error for input that already has a correct diagnostic, which is a defect to report once rather than a gap to file.
- A gap whose only evidence is a shape that was *found*, such as a program's output or a marker asserted present, because neither distinguishes "blocked" from "accepted with the wrong shape".
- A gap that needs new grammar or new semantics and has a working substitute for every use in the code being translated; that is a request to be filed only when the substitute genuinely cannot carry the code.
