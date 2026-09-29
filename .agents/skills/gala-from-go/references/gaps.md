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
| `.Size()` on a value whose type was inferred from a Go call | the transpiler resolves a receiver's type from the `.gala` file's own text and does not read a Go sibling's signature, so the call is emitted against a type it never resolved | type resolution across the Go boundary, or an accepted type annotation on a multi-value binding |

`const` is the one construct here with nothing standing in for it.
`var` and `val` cover every use except a value that has to remain a Go compile-time constant, which is the whole of the gap, and it is a small one.
A file whose only `const` is a named number transliterates cleanly; report it only when a constant has to stay constant for a Go reader or a wire format.

The other two are boundary gaps, and both are about the Go side of the boundary rather than the GALA side, which is what makes them gaps rather than defects.

A comment is dropped wherever it sits, and the two places it hurts are far apart.
A doc comment on an exported declaration disappears from `go doc`, and it cannot be rescued by a handwritten sibling, because a sibling cannot re-declare a function to attach a comment to it.
A `swag` annotation disappears from the generated OpenAPI document, and it cannot be rescued either, because an annotation has to sit immediately above its declaration; a file that loses one silently drops a documented endpoint from the contract.
The substitute is to keep the comment in the `.gala` source, which is where it stays readable to whoever edits that source and invisible to every Go tool.
That is a real substitute with a real price, so this is a boundary gap rather than a defect: what breaks is a tool on the other side of the boundary, and a tool cannot be changed from inside the transpiler.

The `.Size()` case is a narrower reading of the receiver-resolution defect recorded below under [Defects Observed While Pinning Rows](#defects-observed-while-pinning-rows), and the two are worth reading together.
The defect says the receiver is not typed; this says there is no way to type it yourself, which is what turns a lowering bug into a gap.
A receiver whose type is written out resolves fine, even when the type is declared in another package as a GALA struct:

```gala
func CountMany(logs []types.Log) int = logs.Size()   // builds
```

A receiver whose type is *inferred* from a call into a Go-declared function does not, even when that function sits in the same package and returns a GALA-declared type:

```gala
var logs, err = List()   // List is declared in repo.go and returns []Log
return logs.Size()        // logs.Size undefined (type []Log has no field or method Size)
```

The obvious repair is to annotate the binding, and it is closed: a multi-value binding takes no type annotation, so `var logs []Log, err = List()` is a parse error.
`go_interop` has `SliceCap` but no length helper either, and `go_interop.SliceFrom(x, 0).Size()` resolves the receiver only by copying the whole slice to read its length.
So there is no spelling, and the price is that any file counting the elements of a slice returned by a Go function cannot be translated at all.

## Constructs That Were Never Gaps

Two of these were listed as blocked before they were checked against the language, and they are kept here because the reasoning is what stops them coming back.

A Go function literal inside a composite literal is a parse error in the Go spelling only.
A GALA lambda in that position transpiles and builds, so the construct was never missing and there was nothing to file.

Assignment to a `:=` binding is a documented semantic, not a gap.
The language specification states that `:=` bindings are immutable, and the rejection is that rule working as written; `var` is the mutable form and the migration is mechanical.
A construct the language deliberately refuses is a fact about the language, not a gap in it.

A construct the transpiler accepts and then lowers wrongly is neither of these.
It is a defect, and it belongs under [Defects Observed While Pinning Rows](#defects-observed-while-pinning-rows).

## Defects Observed While Pinning Rows

These are not gaps.
Each is a construct GALA is supposed to handle, where the handling is wrong, so each is an issue rather than a request, and none of them is a reason to leave the file handwritten without saying so.
All three below are filed, and the contrast in each one is what makes the fix legible, so a new report on the same construct should lead with the contrast rather than the repro.

Two things make a defect report land, and both are about the evidence rather than the diagnosis.

Lead with the artifact the transpiler itself generated that states the correct shape.
The compiler emits the declaration and the extraction it later mislowers, so its own output contains the specification of what it should have done; a report that quotes that line needs no argument about intent, because the contradiction is inside one file.
A reader who has not seen your code can then check the claim against the compiler rather than against your reasoning.

Choose your assertion to match the claim, and notice that the two directions are not interchangeable.
A check that a marker is *present* pins a shape the compiler produced, so it cannot back a claim that a construct is blocked; a check that a marker is *absent* pins the shape's absence, and a construct is blocked exactly when the thing it needs is not emitted.
The same asymmetry is why a program that prints the right answer proves nothing about whether a wrapper is there.

**`fallthrough` inside a `match` arm produces an internal error instead of a diagnostic.** Filed.
A bare `fallthrough` statement is correctly rejected as a forbidden statement keyword, but the same word inside a `match` arm passes the parser, reaches codegen, and produces Go that does not parse, so the transpiler reports an internal transpile error.
The transpiler writes that unparseable output to a file and names it in the error, so the report carries its own evidence; include the file's contents rather than describing them.

**A `match` on a struct whose fields are declared `var` emits an unwrap the field does not have.** Filed.
The struct's synthesized extractor is called, and the emitted read of a `var` field carries a `.Get()`; the same match over the same fields declared without `var` builds and runs.
The transpiler generates both the struct and the extractor, so the generated `Unapply` return type already states whether the field is wrapped; the match lowering ignores it and unwraps unconditionally.
Report both spellings, because the contrast is what makes the bug legible.

**`.Size()` on a field of a struct declared in a handwritten Go sibling does not build.** Filed as a documentation and diagnostics gap, not a bug.
The transpiler cannot know the field's type, so it emits a method call on a plain Go type that has no such method.
The same call on a GALA-declared struct lowers correctly, so the trigger is a mixed package rather than the construct itself, and it holds for a plain `string` field as much as for a slice.
The specification and the lint skill both sanction `.Size()` as the replacement for `len` with no caveat about an unresolved receiver, and the failure surfaces only as a Go compiler error against generated code the reader did not write, which is what makes it a documentation and diagnostics gap.
It is also worth knowing as a limitation when deciding whether a member can move into a `.gala` file beside a `.go` sibling.

**A `:=` binding mislowers a method call made on it.** Not filed.
The `SKILL.md` trap that records a `:=` binding as a `val` binding says the value wraps and that every read emits `.Get()`.
The receiver of a method call is worse than a read: the transpiler treats the call as a read of the binding and unwraps the receiver, which is never right for a method defined on the value's own type.

```gala
var usersRouter = router.Group("/users")
usersRouter.GET("/:userId", getPublicUserProfile)   // emits usersRouter.GET(...), builds
```

```gala
usersRouter := router.Group("/users")
usersRouter.GET("/:userId", getPublicUserProfile)   // emits usersRouter.Get().GET(...), does not build
```

The transpile is clean, there is no diagnostic, and the failure is a `go build` error against generated code the reader did not write.
The two-spelling contrast is one keyword and the emitted artifact states the correct shape in the first case, so a report needs no argument about intent.
Assert the absence as well as the presence: the `var` form emits no `std.NewImmutable` anywhere in the file, and that absence is what distinguishes the two lowerings rather than merely the presence of a wrapper in one of them.

**A block-bodied lambda in return position cannot infer its parameter type.** Not filed; the diagnostic already names the fix.
`func F() T = (x) => { ... }` is refused with `GALA-E0033`, even though the hint offers "a typed val, function argument, or **return**" and the return type is written out.
Annotating the parameter, `(x T) => { ... }`, is accepted.
The `constructs.md` row for a block-bodied lambda is marked `direct` on the strength of a minimal form whose parameter happened to be inferable, so the row needs the annotated spelling alongside the bare one.

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

Then the ones that are not about a construct at all:

- A style preference, such as preferring a GALA collection to a Go slice where the value only ever meets a Go API.
- A row you have not transpiled on the compiler you have.
- A failure that a transpile-and-run on a minimal repro does not reproduce, which is a signal to look for a context effect in the real file rather than to report the isolated result.
- A `GALA-E0017` internal error for input that already has a correct diagnostic, which is a defect to report once rather than a gap to file.
- A gap whose only evidence is a shape that was *found*, such as a program's output or a marker asserted present, because neither distinguishes "blocked" from "accepted with the wrong shape".
- A gap that needs new grammar or new semantics and has a working substitute for every use in the code being translated; that is a request to be filed only when the substitute genuinely cannot carry the code.
