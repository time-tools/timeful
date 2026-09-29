# Gaps And How To Report Them

A gap is a Go construct GALA cannot express, or cannot express in a shape a Go caller can use.
This file owns two things: the vocabulary for classifying a gap, and the shape of a report worth filing.

Read [`constructs.md`](constructs.md) first.
It says what each construct does today, and a construct with a `workaround` row is not a gap.

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

`const` is the one construct here with nothing standing in for it.
`var` and `val` cover every use except a value that has to remain a Go compile-time constant, which is the whole of the gap, and it is a small one.
A file whose only `const` is a named number transliterates cleanly; report it only when a constant has to stay constant for a Go reader or a wire format.

## Constructs That Were Never Gaps

Two of these were listed as blocked before they were checked against the language, and they are kept here because the reasoning is what stops them coming back.

A Go function literal inside a composite literal is a parse error in the Go spelling only.
A GALA lambda in that position transpiles and builds, so the construct was never missing and there was nothing to file.

Assignment to a `:=` binding is a documented semantic, not a gap.
`docs/GALA.MD` states that `:=` bindings are immutable, and the rejection is that rule working as written; `var` is the mutable form and the migration is mechanical.
A construct the language deliberately refuses is a fact about the language, not a gap in it.

A construct the transpiler accepts and then lowers wrongly is neither of these.
It is a defect, and it belongs under [Defects Observed While Pinning Rows](#defects-observed-while-pinning-rows).

## Defects Observed While Pinning Rows

These are not gaps.
Each is a construct GALA is supposed to handle, where the handling is wrong, so each is an issue rather than a request, and none of them is a reason to leave the file handwritten without saying so.
All three below are filed, and the contrast in each one is what makes the fix legible, so a new report on the same construct should lead with the contrast rather than the repro.

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

<`gala explain --list`, the specification sections you read, and the upstream triage issue>
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
- A gap that needs new grammar or new semantics and has a working substitute for every use in the code being translated; that is a request to be filed only when the substitute genuinely cannot carry the code.
