# Gaps And How To Report Them

A gap is a Go construct GALA cannot express, or cannot express in a shape a Go caller can use.
This file owns two things: the vocabulary for classifying a gap, and the shape of a report worth filing.

Read [`constructs.md`](constructs.md) first.
It says what each construct does today, and a construct with a `workaround` row is not a gap.

## Classify Before Reporting

A row is a gap only if it has no substitute.
Classify it into exactly one of these, because the class determines whether there is anything to file at all.

| Class             | Meaning                                                                                       | File it? |
| ----------------- | --------------------------------------------------------------------------------------------- | -------- |
| direct            | The construct is written as it stands; no substitution is involved.                             | No       |
| workaround        | Only an interop helper or a plain-Go spelling preserves the behavior.                           | No       |
| analog            | A GALA-native construct expresses the same behavior, such as `match`, `Try`, `use`, or `.ByteSize()`. | No    |
| boundary gap      | GALA expresses the behavior, but not as a Go-shaped drop-in, so a Go API or a wire format changes. | Yes  |
| language gap      | GALA has no syntax or no semantics for the construct, so nothing in the language expresses it.     | Yes      |

The two reportable classes differ in what a fix looks like.
A language gap needs new grammar or new semantics.
A boundary gap needs an interop escape hatch: a way to produce or accept the Go shape, even if the GALA side is a different shape.
A boundary gap is real even though the behavior is expressible, because the thing that breaks is a caller on the other side of the boundary, and a caller cannot be changed from inside the transpiler.

Do not promote a `workaround` to a gap because the workaround is ugly.
A helper that compiles and preserves the behavior is an answer, and a report that ignores it is answered with "use the helper".
Do not demote a boundary gap to a workaround because the repository has a local convention for living with it; a local convention is evidence about impact, not about whether the gap exists.

## Blocked On The Compiler In Hand

The rows below were each re-derived by transpiling a minimal repro on the compiler in hand, and each is a gap rather than a workaround.
Re-verify a row before filing it, because a release may have lifted it.

| Construct                    | What happens                                        | Code       | What a fix needs                                                             |
| ---------------------------- | --------------------------------------------------- | ---------- | ---------------------------------------------------------------------------- |
| `const`                      | parse error; the keyword is not in the grammar       | parse error | a const declaration whose value stays a Go compile-time constant            |
| backtick struct tag          | parse error in a field list                          | parse error | tag preservation on emitted structs                                          |
| `switch`                     | parse error, single-case or multi-case                | parse error | a `switch` form, or a documented rule that `match` is the only answer        |
| embedded field               | parse error, bare and `embed` spellings alike         | parse error | embedded-field syntax and field promotion                                    |
| `struct{}` in a type position | parse error                                          | parse error | `struct{}` as a type expression                                              |
| anonymous struct type        | parse error in a type position                        | parse error | anonymous struct type syntax                                                 |
| fixed-size array type        | parse error in any position                           | parse error | `[N]T` grammar                                                               |
| type assertion               | parse error                                           | parse error | a type-assertion expression form                                             |
| `interface{}`                | parse error; only `any` parses                        | parse error | acceptance of the Go spelling, for wire-facing signatures                   |
| `select`                     | parse error                                           | parse error | a `select` statement, or a documented multiplexing analog                    |
| `goto`                       | parse error                                           | parse error | a labelled-jump form, or a documented answer                                 |
| `fallthrough` in a `match` arm | internal transpiler error, not a diagnostic      | `GALA-E0017` | a diagnostic for the arm, and then a documented answer                       |
| method on a defined type     | rejected; the declared type is an alias to a non-local type | `GALA-E0048` | an explicit newtype declaration that carries methods                     |
| `if` initializer             | rejected                                              | `GALA-E0047` | an initializer slot, or acceptance that `Try` plus `match` is the only answer |
| `recover`                    | forbidden builtin; `Try` captures a panic but cannot resume in place | `GALA-E0035` | an in-place `recover`                              |
| `multi-value return`         | parse error in a signature                            | parse error | multi-value return signatures, or an interop escape hatch                    |
| `chan` type in a signature, field, or local | parse error                          | parse error | channel types at the Go boundary                                             |
| function literal inside a composite literal | parse error, Go spelling only          | parse error | acceptance of the Go spelling; a GALA lambda in that position parses        |
| grouped parameters           | rejected                                              | `GALA-E0034` | acceptance of the Go spelling                                                |
| nested function            | rejected                                              | `GALA-E0052` | a local-function form, or a documented answer                                 |
| assignment to a `:=` binding | rejected; the binding is immutable, with no diagnostic code | —      | a mutable rebinding form, or acceptance that `var` is the only answer        |
| a `match` over a struct      | rejected for want of a default case                    | `GALA-E0003` | exhaustive matching over a non-sealed type, or a documented rule            |

The `fallthrough` row is a distinct kind of finding.
A construct that is correctly rejected with a diagnostic and a hint is a documented gap.
A construct that makes the transpiler emit unparseable Go and then reports an internal error is a defect in the rejection path, and it is worth reporting as one, because the same class of bug will silently accept other invalid input.
The hint text on `GALA-E0017` names the upstream issue tracker; include the generated Go the transpiler left behind, since it is the evidence.

## Defects Observed While Pinning Rows

These are not gaps.
Each is a construct GALA is supposed to handle, where the handling is wrong, so each is an issue rather than a request, and none of them is a reason to leave the file handwritten without saying so.

**`fallthrough` inside a `match` arm produces an internal error instead of a diagnostic.**
A bare `fallthrough` statement is correctly rejected as a forbidden statement keyword, but the same word inside a `match` arm passes the parser, reaches codegen, and produces Go that does not parse, so the transpiler reports an internal transpile error.
Report it with the generated Go the transpiler wrote for inspection, because that file is the whole evidence.

**A `match` on a struct whose fields are declared `var` emits an unwrap the field does not have.**
The struct's synthesized extractor is called, and the emitted read of a `var` field carries a `.Get()`; the same match over the same fields declared without `var` builds and runs.
The two spellings differ only in whether the field is wrapped, so unwrapping the unwrapped one is a codegen bug rather than a documented rule.
Report both spellings, because the contrast is what makes the bug legible.

**`.Size()` and `.ByteSize()` on a field of a struct declared in a handwritten Go sibling do not build.**
The transpiler cannot know the field's type, so it emits a method call on a plain Go type that has no such method.
The same calls on a GALA-declared struct lower correctly, so the trigger is a mixed package rather than the construct itself.
Worth reporting with the mixed-package repro, and worth knowing as a limitation when deciding whether a member can move into a `.gala` file beside a `.go` sibling.

## Report Template

The upstream repository has a single open issue that triages Go constructs GALA does not have, and a report belongs there rather than in a new issue.
Check that issue first for an existing entry, and add to it when the construct is already listed; open a new issue only when the construct is genuinely new to it.

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

Do not file these, and do not spend a report slot on them:

- A construct with a `workaround` row in [`constructs.md`](constructs.md).
- A style preference, such as preferring a GALA collection to a Go slice where the value only ever meets a Go API.
- A row you have not transpiled on the compiler you have.
- A failure that a transpile-and-run on a minimal repro does not reproduce, which is a signal to look for a context effect in the real file rather than to report the isolated result.
- A `GALA-E0017` internal error for input that already has a correct diagnostic, which duplicates a row above instead of adding one.
