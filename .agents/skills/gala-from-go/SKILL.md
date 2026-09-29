---
name: gala-from-go
description: Translate a Go file to GALA, showing each Go construct and whether GALA can express it: use when the user asks how to rewrite, port, or convert Go code in GALA, asks whether a Go construct is expressible in GALA, or asks what a GALA gap is and how to report one.
---

# GALA From Go

GALA transpiles to Go, so a package can hold both a `.gala` file and a handwritten `.go` sibling.
That is the fact this skill is built on: a construct GALA cannot express does not disqualify a file, it moves a member out of it.
The outcome of a translation is therefore a package, not a verdict on a file.

This skill translates Go into GALA and reports what GALA cannot express.
It does not govern how a repository stores or commits the result; that belongs to whatever convention the project already has.

## The Three Files

| Question                                                             | File that owns it                                      |
| -------------------------------------------------------------------- | ------------------------------------------------------ |
| In what order do I do this, and what does each step decide?           | this file                                              |
| Given this Go construct, what is the GALA spelling, and does it work? | [`references/constructs.md`](references/constructs.md) |
| Is this a gap, how is it classified, and how do I report it?          | [`references/gaps.md`](references/gaps.md)             |

Ask each file only the question it owns, and read a construct's row before writing a line of GALA.
A construct that works and a construct GALA refuses are both plain Go until the row says otherwise, and the two produce opposite outcomes.

## Before Starting

Establish three things, because each of them silently invalidates a translation rather than failing loudly.

Record the compiler version and keep it for the whole session:

```sh
gala version
```

The construct rows in this skill are version-agnostic, and that is a deliberate choice rather than a hedge.
A construct verdict is a property of the compiler you have, and a document that pins a release goes stale the moment the release rolls.
Every row names a diagnostic code or a check instead, so the way to trust a row is to run its check.

Note where you transpile.
The transpiler emits `//line` directives naming the path it was given, so transpiling from a directory other than the package directory writes directives that do not resolve from the generated file and then show up as a diff on every regeneration.
Always transpile from the package directory, and transpile to a scratch path while you are iterating.

Read the project's own rules before touching its layout.
A repository that commits the generated file, one that regenerates on build, and one that forbids generated code in a directory all want different things from the same `.gala` source.

## Triage Before Translating

Give every file one of three verdicts, and take the first that fits.
The verdicts are about the file; the gap classes in [`references/gaps.md`](references/gaps.md) are about the constructs inside it, and the two are separate questions.

1. **Rewrite it whole.**
   Every construct has a row with a `direct` or `workaround` status, and the Go callers tolerate the shapes the transpiler emits.
2. **Split it, with a handwritten sibling.**
   Some members are expressible and some are not, so translate the expressible members into the `.gala` file and keep the rest in a `.go` sibling in the same package.
   This is the rung most blocked files actually reach, and it is available precisely because GALA transpiles to Go.
3. **Keep it handwritten.**
   A shape is blocked and splitting would leave less in the translated file than the split is worth, or a Go-facing API or wire format would change.

Two shape rules decide whether a file can be rewritten whole, and both are about what escapes the package rather than about taste.

- An exported package-level `val` emits `var X = std.NewImmutable(...)`, so a Go caller has to call `.Get()`; keep `var` for anything that crosses into Go.
- A `:=` binding is a val-style binding, not a neutral one: it lowers through `std.NewImmutable` and every read emits `.Get()`.
  Use `var` for anything a Go caller or a sibling `.go` file can see.

A construct family with no row in [`references/constructs.md`](references/constructs.md) is a blocking unknown, not a licence to guess.
Transpile the minimal form yourself, add the row, and record what you found.

## The Mechanical Pass

Work one construct at a time, in this order, and transpile to a scratch path after each step so a parse error costs nothing.

1. Skeleton.
   Translate the package clause, the imports with their aliases, and the signatures, with bodies empty.
   Transpile it.
   A signature is where the first blocked construct usually shows up, and finding out now rather than after the bodies are written is the whole point of the order.
2. Bindings.
   Translate declarations, and prefer `var` for anything visible from Go.
   Multi-value bindings are the one place where `var` and `:=` differ in emitted shape, so transpile both spellings once and read the difference rather than assuming.
3. Statements.
   Translate control flow, applying the mechanical substitutions in [`references/constructs.md`](references/constructs.md) one row at a time.
4. Bodies and calls.
   Translate the calls, choosing the helper or the GALA-native form by which container type the value actually is, not by how the file is styled.
5. Comments.
   The transpiler emits no documentation comments, so package and declaration comments belong in a handwritten `doc.go`.

Stop at the first construct whose row is not `direct` and apply the triage verdicts again for the member rather than the file, because a member GALA cannot carry is a split, not a failure.
A `workaround` or an `answered` row is not that case: the row already names the spelling to use, so take it and keep going.

## Verify

A translation that transpiles is not a translation that works, and the difference is where this skill spends most of its attention.

```sh
# from the package directory
gala transpile -i <name>.gala -o /tmp/<name>.go    # iterate to a scratch path
gala transpile -i <name>.gala -o <name>.go         # write the real file
gofmt -l <name>.go                                # formatting
gala transpile -i <name>.gala -o /tmp/again.go     # transpile twice
diff <name>.go /tmp/again.go                      # regeneration is byte-identical
go build ./...                                    # from the module root
go test ./...                                     # the project's own suite
```

Read the emitted Go for every shape claim and build and run for every acceptance claim.
That distinction is not pedantry: a value wrapped in `std.Immutable` is invisible to a program that only passes it around, so a green test run does not disprove that the wrapper is there.
When a claim is about what the transpiler emits, read the file; when it is about whether the program is correct, run it.

## Stop And Report

Leaving a file or a member handwritten is a correct outcome, not a failure of this procedure.
Stop and report when any of these holds, and say which one applies:

- The file's triage verdict is "keep it handwritten".
- The translation would change a Go API or a wire format that a caller outside the package depends on.
- A member's construct has no row, and moving it to a sibling would leave less translated than the split is worth.
- A transpile, build, or test failure survives applying the row's documented substitution, which is evidence the row is stale for this compiler and not permission to invent a replacement.
- The only way forward is a different compiler version, which re-baselines every generated file in the project and is a separate exercise.

When you stop, report the verdict, the construct that drove it, the diagnostic, and the compiler version.
Do not proceed by hand-editing a generated file, by reshaping a Go-facing signature to satisfy the transpiler, or by relaxing a test.

If a construct turns out to be inexpressible rather than merely awkward, [`references/gaps.md`](references/gaps.md) owns the classification and the report template.
Classify before reporting, because a construct with a substitute is not a gap and an issue that ignores a substitute gets answered with the substitute.

## Traps

Each of these is a way the translation goes wrong without producing an error at the point of the mistake.

- **A `:=` binding is a `val` binding.**
  It lowers through `std.NewImmutable` and every read emits `.Get()`.
  A program that only passes the value around still runs, so the wrapper survives a green test and shows up as a build failure in a Go sibling or a Go caller.
  Use `var` for anything visible from Go.
- **A package-level `val` is not a plain Go var.**
  It emits `var X = std.NewImmutable(...)`, and a Go caller needs `.Get()`.
  Inside the GALA file it already reads as a plain value, so writing `.Get()` there produces a double unwrap that fails to build.
- **A non-`var` struct field is not a plain Go field.**
  It becomes `std.Immutable[T]`, and the struct gains synthesized `Copy`, `Equal`, and `Unapply` members.
  A generic struct additionally gains an `Instance` interface and an `Is<Type>()` method, so check the emitted member set for a non-generic and a generic struct separately rather than assuming one set covers both.
  Declare `var` for any field Go code reads.
- **A struct declared in a handwritten sibling is not a GALA struct.**
  Constructing it positionally, calling `.Size()` or `.ByteSize()` on one of its fields, or calling a method on it can transpile and then fail to build, because the transpiler emits a call on a type it does not know.
  This is a mixed-package effect and does not reproduce in a package with no handwritten sibling, so a repro in isolation proves nothing either way.
- **A struct pattern over `var` fields transpiles and then fails to build.**
  A `match` on a struct calls its synthesized `Unapply`, and the emitted read of a `var` field carries a `.Get()` that the plain Go field does not have; the same match over non-`var` fields builds.
  Check both field kinds rather than assuming they behave alike, and if it reproduces, report it as a defect: the two field kinds differ only in their wrapper, and the transpiler unwraps the wrong one.
- **A `match` needs a default case unless the scrutinee is a sealed type.**
  A match over a struct that looks exhaustive is still rejected, while an exhaustive match over a sealed type is not.
- **`++` and `--` work, but only on a mutable binding.**
  A `val` or a `:=` binding is immutable, so a post statement or an increment on one is rejected.
  A counter that is never reassigned is a missed `val`, and a counter that is reassigned through a `var` is correct.
- **`==` on a collection is not Go's `==`.**
  A collection value lowers to a wrapper, so comparing two of them emits a comparison of the wrappers and the Go compiler rejects it.
  Use the collection's own equality method.
- **A float format verb on an integer does not fail; it mangles.**
  `f"$n%.2f"` where `n` is an integer lowers to a `Sprintf` with a float verb on an integer, which prints Go's `%!f(int=...)` marker rather than raising an error.
  That matches Go, so the translation is faithful, but it means a mangled format string is not evidence that the translation is wrong.
- **A `use` binding is not the acquired value.**
  `use x = acquire` takes a single-value acquire, so a call returning `(T, error)` has to have its error handled first, and the resulting binding is a `Try`, which is why calling a method on it fails.
- **`Try` and `FromError` are not interchangeable for a void Go call.**
  Both transpile cleanly and the difference shows at run time, so a transpile alone does not settle it.
- **A bare builtin or statement keyword is a hard error, not a style finding.**
  `len`, `append`, `make`, `copy`, `delete`, `close`, `cap`, `new`, `panic`, and `recover` are `GALA-E0035`; `defer`, `go`, `select`, `goto`, `fallthrough`, and `chan` are `GALA-E0036`.
  Run `gala explain GALA-Exxxx` for the code's page and its minimal repro.
- **`gala doc` does not describe a package whose source is handwritten Go.**
  It answers for GALA-source packages, so do not send an agent to a Go-backed package to read its helper list; the transpiler's own diagnostic hint lists the sanctioned helpers instead.
- **A transpiler internal error is not a verdict.**
  `GALA-E0017` means the transpiler emitted Go it could not parse back, which is a defect in a rejection path rather than an answer about the construct.
  Treat it as "this construct is rejected badly", not as "this construct is accepted".

## Never

- Never pin a construct verdict to a release; verify the row on the compiler you have.
- Never claim a construct works because it transpiles; read the emitted Go or run the program.
- Never pick a container helper by the file's style rather than by the value's actual container type.
- Never reshape a Go-facing signature, a struct tag, or a wire format to make the transpiler accept the file.
- Never relax a test to accommodate a translation.
- Never file a gap report for a construct that has a substitute; classify it first.
- Never copy a row from a stale document without running its check.
- Never hand-edit a generated file; it carries a `DO NOT EDIT` header and the next transpile discards the edit without a trace.
- Never transpile from a directory other than the package directory, because the emitted `//line` directives name the path the transpiler was given.
