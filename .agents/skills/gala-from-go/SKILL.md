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
The runtime policy is part of those rules: a project that consumes the published stdlib module passes `gala transpile --stdlib-module go.gala.fyi/stdlib` so the generated imports name the module pinned in its `go.mod`, and a project that keeps generated Go runtime-free rejects the same imports.

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

Three shape rules decide whether a file can be rewritten whole, and all three are about what escapes the package rather than about taste.

- An exported package-level `val` emits `var X = std.NewImmutable(...)`, so a Go caller has to call `.Get()`; keep `var` for anything that crosses into Go.
- A `:=` binding is a val-style binding, not a neutral one: it lowers through `std.NewImmutable` and every read emits `.Get()`.
  `val` emits the same thing, so `var` is the only safe spelling of the three.
  Use `var` for anything a Go caller or a sibling `.go` file can see.
- A `sealed type` is the largest of the three: it emits one merged struct whose every field is a `std.Immutable`, a private discriminant with one constant per variant, an `Apply` and an `Unapply` per variant, an `is<Variant>()` helper per variant, and `Copy`, `Equal`, and `String` on the parent.
  A variant's `Unapply` returns a `std.Option`, and a match over the type ends in `panic("unreachable")` on the branch no variant can reach.
  There is no form a Go caller can construct positionally, so a file that introduces one reshapes its package's exported surface as surely as a `val` does.
  Read the emitted struct before committing to it rather than trusting this summary.

The three rules are about a Go-facing surface, not about which constructs a rewrite may use.
Every runtime construct is available to a translated file: `val` and `:=` bindings, `Try`, `Option`/`Either`, sealed types, the immutable collections, and the `go_interop` helpers all transpile, and a project on the published stdlib writes their imports as `go.gala.fyi/stdlib/...` with `gala transpile --stdlib-module go.gala.fyi/stdlib`.
The spellings that remain boundary-changing for a Go-facing surface are: an exported `val` behind `std.Immutable`, a sealed type's merged struct, a non-`var` struct field behind `std.Immutable[T]`, and a `Codec` where a struct tag defined the wire shape.

### The Mixed Package

Verdict 2 creates a package that holds a `.gala` file beside a handwritten `.go` sibling, and that is where the transpiler stops being a function of the file you are translating.
Four rules belong to the package rather than to the file; the first is loud at the transpile, and the others fail somewhere other than it, which is what makes them expensive:

- Every `.gala` file imports everything it uses, including a package a sibling file already imports.
  Import resolution is per file, so an import one file omits is refused at transpile time with `GALA-E0023` rather than carried into the generated file for `go build` to find.
  Whether a sibling imports the package does not change that: a `.gala` sibling, a handwritten `.go` sibling, and no sibling at all produce the same diagnostic, so there is no case where a missing import is invisible until the build.
  An unqualified GALA-runtime name is the same rule from the other side, and its hint names the package that declares the name and both spellings of the fix.
- A declaration in a handwritten sibling is visible to the `.gala` file as a name but is not a GALA declaration, so the transpiler emits calls against a type whose members it never saw.
  [`references/gaps.md`](references/gaps.md) records the case known to fail and why a repro in isolation settles nothing either way.
- One receiver's methods have to stay in one language across the package.
  A method declared on a receiver in a `.gala` file becomes that receiver's whole method set for every *importer* of the package, because the importer resolves the type from the package's GALA index and never merges the methods its handwritten `.go` siblings declare; the importer's call to a sibling-declared method is then refused with a false `GALA-E0044` (checked 2026-10-10 on rev `cd2fdcb5`, where a `repository.gala` carrying eighteen methods broke five committed twins until they moved back to a handwritten sibling).
  Translate the methods into `.gala` only when the receiver's methods all move with them, and keep the declaration split's methods handwritten otherwise.
- A bare name that both a translated file and an imported package export is contested rather than settled: on the compiler this skill was measured against, the local declaration wins, and a report of the opposite is a fact about the reporter's compiler rather than about the language.
  Do not depend on the outcome either way: never name such a type in a `.gala` file, and call a handwritten constructor in the sibling instead, so the ambiguity never gets a chance to arise.

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
   Transpile a declaration comment and read the emitted Go before deciding where comments live, because compilers differ: the compiler this skill was last measured against emits a comment attached to a declaration and the package comment, and drops comments inside a function body.
   Where a compiler drops a declaration comment, the `doc.go` route recovers only the package comment, because a declaration comment and a `swag` annotation have to sit immediately above their declaration and the only spelling in another file is a bodiless re-declaration, which does not compile.
   An annotated handler then has to stay handwritten, which is a boundary gap rather than a defect in the file.

Stop at the first construct whose row is not `direct` and apply the triage verdicts again for the member rather than the file, because a member GALA cannot carry is a split, not a failure.
A `workaround` or an `answered` row is not that case: the row already names the spelling to use, so take it and keep going.

## Verify

A translation that transpiles is not a translation that works, and the difference is where this skill spends most of its attention.

```sh
# from the package directory; a project on the published stdlib carries
# --stdlib-module go.gala.fyi/stdlib on every transpile:
gala transpile --stdlib-module go.gala.fyi/stdlib -i <name>.gala -o /tmp/<name>.go  # iterate to a scratch path
gala transpile --stdlib-module go.gala.fyi/stdlib -i <name>.gala -o <name>.go       # write the real file
gofmt -l <name>.go                                # formatting
gala transpile --stdlib-module go.gala.fyi/stdlib -i <name>.gala -o /tmp/again.go   # transpile twice
diff <name>.go /tmp/again.go                      # regeneration is byte-identical
go build ./...                                    # from the module root
go test ./...                                     # the project's own suite
```

Read the emitted Go for every shape claim and build and run for every acceptance claim.
That distinction is not pedantry: a value wrapped in `std.Immutable` is invisible to a program that only passes it around, so a green test run does not disprove that the wrapper is there.
When a claim is about what the transpiler emits, read the file; when it is about whether the program is correct, run it.

### Checking A Shape Claim

Three kinds of claim need three different checks, and no one of them substitutes for another.
A rejection settles only that something was refused, a passing run settles only what the program observed, and only the emitted text settles what was generated.
Pick the check the claim needs before deciding it is true.

Two rules make reading the emitted text reliable, and both come from the same property: the file is generated, so its shape is not where the author's attention went.

- Assert the absence, not only the presence.
  A marker that must be present proves the wrapper is there; a marker that must be *absent*, taken from the same program in the other spelling, is what proves the two spellings differ.
  A transpiler that started wrapping both would pass every presence check and every output assertion in the suite, and fail only the absence.
- Name a semantic thing in the marker, never layout.
  A wrapper constructor call, a synthesized type or method name, and a call on a binding all survive formatting.
  Indentation, a line break, and a column do not, because the output is formatted and carries `//line` directives naming your own source.

The import block is the cheapest read in the file and the most useful one, because it is the transpiler's own account of what the translation cost.
A file that emits no runtime import pulled in nothing; a file that emits one has moved a dependency into the generated code, whatever the source looked like.
Read it before deciding whether a rewrite is worth its blast radius, and when the project adopted the published stdlib module confirm every runtime import names that module rather than `martianoff/gala`.

Finally, read the emitted `//line` directives themselves.
They name the path the transpiler was given, so relative directives are the evidence that you transpiled from the package directory rather than a claim that you did.

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
  The worst case is the method-call *receiver*, which is not a read at all: `x := f()` then `x.M()` emits `x.Get().M()`, which does not exist in Go when `M` has a pointer receiver.
  `val` emits the identical artifact, so `var` is the only safe spelling of the three, and it is the shape to reach for whenever a binding's value has methods.
- **A package-level `val` is not a plain Go var.**
  It emits `var X = std.NewImmutable(...)`, and a Go caller needs `.Get()`.
  Inside the GALA file it already reads as a plain value, so writing `.Get()` there produces a double unwrap that fails to build.
- **A non-`var` struct field is not a plain Go field.**
  It becomes `std.Immutable[T]`, and the struct gains synthesized `Copy`, `Equal`, and `Unapply` members.
  A generic struct additionally gains an `Instance` interface and an `Is<Type>()` method, so check the emitted member set for a non-generic and a generic struct separately rather than assuming one set covers both.
  Declare `var` for any field Go code reads.
- **A `sealed type` is not a struct with a tag.**
  One merged struct carries every variant's fields as `std.Immutable`, plus a private discriminant and one constant per variant, and each variant gets `Apply` and `Unapply` with `is<Variant>()` helpers on the parent.
  The parent gains `Copy`, `Equal`, and `String`, and a Go caller cannot construct it positionally at all, so a sealed type is a boundary change and not a local rewrite.
- **A `val` holding a collection reads through two unwraps.**
  A `HashMap` bound with `val` and read with `Get` emits `m.Get().Get(k)`: the outer call unwraps the binding and the inner one is the lookup.
  It builds and returns the right value, so nothing reports it until a Go caller or a sibling reads the name it is bound to.
- **A struct declared in a handwritten sibling is not a GALA struct.**
  Constructing it positionally or calling a method on it can transpile and then fail to build, because the transpiler emits a call on a type it does not know.
  A method call on such a type passes through and Go resolves it, so the failure is about the *type's* origin, not the call's.
  The `.Size()`/`.ByteSize()` lowering over such a receiver, and over one inferred from a Go call, was the defect filed as [#613](https://github.com/martianoff/gala/issues/613) and is fixed on the compiler this skill was last measured against; run the corresponding row's check rather than assuming either behavior.
  The mirror direction is worse: a type declared in the `.gala` file whose method is declared in the sibling is *refused* with `GALA-E0044` and a hint that claims the type declares no methods, although Go accepts the program ([#615](https://github.com/martianoff/gala/issues/615)).
  This is a mixed-package effect and does not reproduce in a package with no handwritten sibling, so a repro in isolation proves nothing either way.
- **A bare name in a declared-type position can resolve to an import.**
  A return type, a parameter type, or a `var` annotation takes an imported package's qualifier whenever any import of the file exports that name, even when the package declares the name itself; a constructor position resolves correctly, so the two can disagree inside one file.
  Declare the type in the same `.gala` file, or reach it through a handwritten constructor, and never assume a bare name in a type position means the local type ([#616](https://github.com/martianoff/gala/issues/616)).
- **A sibling file's import is not your import.**
  A qualified name resolves against the imports of the file that writes it, so a file that omits an import its sibling already has is refused with `GALA-E0023` rather than transpiled and left for `go build` to report against the `.gala` file through its `//line` directive.
  Read `undefined: <pkg>` as "this file does not import it" rather than "the package does not import it", because a `.gala` sibling, a handwritten `.go` sibling, and no sibling all give the same diagnostic.
  An unqualified runtime name reads the same way, with a hint naming the declaring package, so there is nothing to infer from which sibling imported what.
- **A struct pattern over `var` fields transpiles and then fails to build.**
  A `match` on a struct calls its synthesized `Unapply`, and the emitted read of a `var` field carries a `.Get()` that the plain Go field does not have; the same match over non-`var` fields builds.
  Check both field kinds rather than assuming they behave alike, and if it reproduces, report it as a defect: the two field kinds differ only in their wrapper, and the transpiler unwraps the wrong one.
- **`type X Y` is an alias, not a defined type.**
  It emits `type X = Y`, so a Go `type Environment string` translated with the same spelling silently becomes interchangeable with every `string`, and every caller still builds.
  Use `opaque type X Y` for a distinct type, and check the emitted declaration for the absence of `=` rather than trusting a green build.
- **`|` in a pattern is an alternative, not bitwise OR, and it has restrictions.**
  `case 1 | 2` matches either value and emits `obj == 1 || obj == 2`; `GALA-E0071` rejects an alternative that binds a name or is `_`, `|` mixed with `+`, `-`, or `^` without parentheses, and `|` inside a comparison or boolean expression of a pattern.
  On a compiler that predates the fix the same spelling was an expression pattern that emitted `obj == 1|2` and matched only `3`, so run the row's check rather than trusting either behavior.
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
- **A byte-boundary truncation has no spelling, and the substitute changes the count.**
  `s[:n]` is a parse error, so a truncation has to become a hand-written loop, and a loop that accumulates runes stops at character boundaries rather than the byte boundary the Go code asked for.
  The rewrite builds, runs, and returns a different length for any non-ASCII input, so decide which of the two meanings the code wanted before writing the loop.
- **A resource combinator needs the resource type resolvable from `.gala` text, or the body parameter binds to `any`.**
  A GALA-declared resource type infers both type arguments with none written out, and `WithLock` infers either way, but a resource type declared in a handwritten `.go` sibling needs both arguments written out, or the body parameter binds to `any` and the generated file does not build ([#618](https://github.com/martianoff/gala/issues/618)).
  A partial type-argument list is its own failure, emitting the transpiler's own type-parameter name as a declared Go type.
  The enclosing function's return type is emitted unchanged in every variant checked; `func F() = <expr>` is the shape that never infers a result ([#617](https://github.com/martianoff/gala/issues/617)).
  A function-typed result cannot be spelled as an explicit type argument at all: `resource.Bracket[context.CancelFunc, func()]` is a parse error, so name the type (`type CloseFn func()`) and pass it, or omit both arguments and annotate the binding (`var closeFn func() = resource.Bracket(cancel, release, body)`), which infers the result from the annotation.
- **A `resource.Bracket` body's panic is re-raised as an error, not as Go's panic value.**
  The body's panic becomes a `Try` Failure holding the runtime's unexported `panicError`, the release still runs on every exit path, and re-raising the Failure delivers that error to an outer `recover`, so a `string` panic arrives as a `*panicError` rather than as a `string`.
  `Unwrap` still reaches the original and `Error()` prints the same message, so a checker that only reads the message sees no difference, while one that inspects the recovered value's type does.
- **A `use` binding is not the acquired value.**
  `use x = acquire` takes a single-value acquire, so a call returning `(T, error)` has to have its error handled first, and the resulting binding is a `Try`, which is why calling a method on it fails.
- **The capture guard covers one concurrency boundary and not the other.**
  A closure that captures a reassignable binding is refused when it crosses into a `concurrent.Future`, with a diagnostic naming the race.
  The same capture is accepted by `go_interop.Spawn`: it transpiles, the enclosing function returns before the spawned body has run, and the program's own output shows the zero value.
  Run a build of the result under the race detector rather than trusting the guard to have covered both.
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
- Never let a sibling file's import stand in for your own, because the check is per file and a missing import is refused rather than carried.
- Never settle a shape claim from a passing run, and never settle it from a presence assertion alone.
