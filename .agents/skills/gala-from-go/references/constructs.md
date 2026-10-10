# Go constructs in GALA

One row per Go construct: the minimal Go, the GALA spelling that replaces it, whether it works, the diagnostic code when there is one, and the check that pins the row.

Read this file by construct, from a Go snippet you actually have in front of you.
Do not read it top to bottom expecting to learn GALA; `SKILL.md` owns the procedure and this file owns the substitutions.

## How To Read A Row

| Column      | What it means                                                                                                    |
| ----------- | ---------------------------------------------------------------------------------------------------------------- |
| `Go`        | The minimal Go form.                                                                                             |
| `GALA`      | The spelling that replaces it. `—` means nothing in the language expresses it, and the row is a gap.               |
| `Status`    | `direct` works as written, `workaround` preserves the behavior through a helper or a different construct, `answered` is refused with a diagnostic that names the replacement, and `semantic` is refused because the language says so. A `defect` is a construct whose handling is wrong, not one that is missing. [`references/gaps.md`](gaps.md) maps each of these to a gap class. |
| `Code`      | The diagnostic code, or `parse error` when the construct never reaches the resolver. `—` when nothing fires.       |
| `Check`     | A command or a step that re-derives the row on the compiler you have. Run it before you rely on the row.           |

The rows are version-agnostic on purpose.
Each names a code or a check rather than a release, because a construct verdict is a property of the compiler you have and not of a document.
Verify the rows your translation depends on rather than trusting this file, and prefer a transpile of the real file over a transpile of the minimal form, because the minimal form can miss a context effect.

## Declarations And Types

| Construct                | Go                                              | GALA                                                            | Status      | Code       | Check                                                                       |
| ------------------------ | ----------------------------------------------- | --------------------------------------------------------------- | ----------- | ---------- | --------------------------------------------------------------------------- |
| `const`                  | `const Answer = 42`                             | `var Answer = 42`, or `val Answer = 42` for a GALA-internal binding  | workaround | parse error | transpile a file with a `const`; the keyword is not in the grammar       |
| package-level `var`      | `var StdOut *log.Logger`                        | `var StdOut *log.Logger`                                        | direct     | —          | transpile; the emitted Go is a plain `var`                                |
| package-level `val`      | `var StdOut *log.Logger`                        | `val StdOut *log.Logger`                                        | workaround | —          | transpile; the emitted Go is `var StdOut = std.NewImmutable(...)`          |
| local `var`              | `var n, err = f()`                              | `var n, err = f()`                                              | direct     | —          | transpile; the emitted Go is a plain `var`                                |
| local `:=`               | `n, err := f()`                                 | `var n, err = f()`                                              | workaround | —          | transpile both spellings and diff; `:=` is a val-style binding             |
| single `:=`              | `n := f()`                                      | `var n = f()`                                                   | workaround | —          | transpile; `:=` binds through `std.NewImmutable`                           |
| `val`                    | `val x = f()`                                   | `val x = f()`                                                   | direct     | —          | transpile; the emitted Go wraps the result                               |
| type alias                | `type X = Y`                                    | `type X = Y`                                                    | direct     | —          | transpile                                                                   |
| defined type              | `type CalendarType string`                      | `opaque type CalendarType string`; GALA's `type X Y` is an alias and emits `type X = Y`, so a bare `type` silently drops the distinct type | workaround | —          | transpile `type X string` and `opaque type X string` and read the emitted declaration; the opaque form emits runtime-backed `Hash`/`Compare` unless a same-package `.go` sibling declares them |
| method on a defined type  | `func (d DateTime) IsZero() bool`               | `opaque type DateTime int64` with the method; a same-package `.go` sibling declaring `Hash`/`Compare` suppresses the synthesized methods, so the generated Go names no runtime package | answered   | `GALA-E0048` | `gala explain GALA-E0048`, then transpile the `opaque type` with and without the sibling and grep the import block; the sibling run names no runtime package, while the default names `go.gala.fyi/stdlib/std` under `--stdlib-module` (`martianoff/gala/std` without it) |
| struct                    | `type Person struct { Name string }`            | `struct Person(Name string)`, one field per line or comma        | workaround | —          | transpile; non-`var` fields become `std.Immutable[T]`                     |
| struct with `var` fields  | `type Person struct { Name string }`            | `struct Person(var Name string)`                                | direct     | —          | transpile; a `var` field stays a plain Go field                          |
| generic struct            | `type Pair[T any] struct { A, B T }`            | `struct Pair[T any](var A T, var B T)`                          | direct     | —          | transpile; a generic struct gains an extra `Instance` interface and `Is<T>()` method |
| struct tag                | ``Name string `json:"name"` ``                  | `Codec[T]` with `.Rename`, `.Omit`, `.OmitEmpty`               | answered   | parse error | transpile a tagged field; backtick tags are not in the grammar           |
| embedded field            | `type Event struct { Base; Name string }`       | a named field plus explicit delegation, losing promotion        | answered   | parse error | transpile both the bare and the `embed` spellings                        |
| anonymous struct type     | `struct { A string }`                           | a named struct                                                   | workaround | parse error | transpile an anonymous struct in a type position                         |
| `struct{}` in type position | `map[T]struct{}`                             | a named empty struct, which changes the element type identity    | workaround | parse error | transpile `struct{}` as a map value type                                 |
| interface declaration     | `type Repo interface { Get(id string) string }` | `type Repo interface { Get(id string) string }`                  | direct     | —          | transpile; declare it with `type X interface`, not a bare `interface`    |
| `interface{}`             | `Details interface{}`                           | `Details any`                                                    | workaround | parse error | transpile both spellings; only `any` parses                              |
| type sum, tagged union   | a `switch` on a type tag dispatching to per-type constructors | `sealed type T { case A(...); case B(...) }`    | workaround | —          | transpile a sealed type and read the emitted parent struct, the per-variant `Apply` and `Unapply`, and the discriminant |
| `chan` type               | `func send(ch chan int)`                        | `concurrent.Future` or `go_interop` where no channel type crosses the boundary | workaround | parse error | transpile a channel type in a signature, a field, or a local |
| fixed-size array type     | `[16]byte`                                      | `Array[T]`, or a struct when the arity carries meaning         | answered   | parse error | transpile an array type in any position                                  |
| `map` type in type position | `map[K]V`                                    | `map[K]V`                                                       | direct     | —          | transpile a field, a parameter, or an alias                              |
| generic declaration       | `type Set[T comparable] ...`                    | `type Set[T comparable] ...`; struct generics are `struct Box[T any](var V T)` | direct     | —          | transpile a generic function and a generic struct separately              |
| grouped parameters        | `func add(a, b int) int`                        | `func add(a int, b int) int`                                     | workaround | `GALA-E0034` | `gala explain GALA-E0034`, then transpile a grouped parameter list      |
| multi-value return        | `func f() (T, error)`                           | `func f() (T, error)` declaring Go's result list; the body is the results' one GALA value, and only a body that is a Go call with exactly those results emits no `std.Try`/`Tuple` | workaround | —          | transpile a Go-result-list signature once with a Go-call body and once with a `Success`/`Failure` body, and read each emitted import list; an older compiler gives a parse error |
| type arguments            | `MapEmpty[string, int]()`                       | comma-separated, never inferred where nothing pins them          | direct     | —          | transpile and read the emitted type arguments                            |

The `var` versus `val` and `var` versus `:=` distinctions are shape rules, not preferences, and they are the first thing to get right in a file that Go code outside the package can see.
A `struct` declaration's fields are all required at construction unless a field declares a default (`var added []string = nil`); an omitted field is `GALA-E0045` even when the equivalent Go keyed literal may omit it.
`references/gaps.md` and the traps section of `SKILL.md` say what each one costs.

A rejection without a code is still a rejection: some checks report a plain error rather than a `GALA-Exxxx` page, so read the message rather than assuming a code is always there.

A `sealed type` is the one declaration in this file whose emitted form a Go caller cannot use at all, so read the generated struct before you translate a type sum into one.
The parent is one struct with a field per variant, each field a `std.Immutable`, plus a private discriminant and one constant per variant; each variant is its own struct with `Apply` and an `Unapply` returning a `std.Option`; and the parent gains `Copy`, `Equal`, `String`, and an `is<Variant>()` helper per variant.
Construction has to go through a variant's `Apply`, and a match ends in `panic("unreachable")` on the branch no variant can reach, so this is a boundary change and not a local spelling.

## Functions And Lambdas

| Construct                  | Go                                                   | GALA                                            | Status      | Code       | Check                                                              |
| -------------------------- | ---------------------------------------------------- | ----------------------------------------------- | ----------- | ---------- | ------------------------------------------------------------------ |
| block function             | `func f() { ... }`                                   | `func f() { ... }`                              | direct     | —          | transpile                                                          |
| expression function        | `func f() int { return 1 }`                          | `func f() int = 1`                              | direct     | —          | transpile                                                          |
| method                     | `func (r Recv) M()`                                  | `func (r Recv) M()`                             | direct     | —          | transpile a method on a struct                                     |
| pointer receiver           | `func (r *Recv) M()`                                 | `func (r *Recv) M()`                            | workaround | `GALA-E0053` | `gala explain GALA-E0053`, then transpile a pointer method on a `val` field |
| variadic parameter         | `func f(xs ...T)`                                    | `func f(xs ...T)`; build a spread with `ArrayOf(xs...)` | direct | —     | transpile a variadic definition and a call site                    |
| blank parameter            | `func f(_ T)`                                        | `func f(_ T)`                                   | direct     | —          | transpile                                                          |
| function type              | `Execute func(args []string)`                        | the same, as a field or parameter type           | direct     | —          | transpile                                                          |
| lambda                     | `func(x int) int { return x }`                       | `(x) => x`; the parameter is parenthesized      | workaround | `GALA-E0042` | `gala explain GALA-E0042`, then transpile a bare-parameter lambda  |
| lambda with a block body   | `func(x int) int { if x > 0 { return 1 }; return 0 }` | `(x) => { if (x > 0) { return 1 }; return 0 }` | direct     | —          | transpile a lambda whose body is a block                          |
| block-bodied lambda in return position | `func F() func(int) int { return func(x int) int { ... } }` | `func F() func(int) int = (x) => { ... }`; the parameter type is inferred from the written return type | direct | — | transpile the return-position form, run it, and read the emitted parameter type |
| a lambda with no inferable parameter | `val f = (x) => x + 1` | `val f func(int) int = (x) => x + 1`, or annotate the parameter | answered | `GALA-E0033` | `gala explain GALA-E0033`, then transpile a bare `val` initializer and a generic return type `func F[T any]() T = (x) => x` |
| lambda in a composite literal | `Command{Execute: func() string { return "x" }}` | `Command(Execute = () => "x")`                  | workaround | parse error | transpile a Go-style function literal inside a literal, then a lambda |
| declaration ordering       | `f(g())` where `f` is declared later                | a declaration may follow its use                                 | direct     | —          | transpile a call to a function declared below it                   |
| nested function            | `func f() { func g() {} }`                           | a lambda bound to a `val`                                     | answered   | `GALA-E0052` | `gala explain GALA-E0052`, then transpile a nested function      |

A Go function literal written with the Go spelling is not a lambda; only the GALA spelling is, and the distinction is a parse error rather than a diagnostic.

## Control Flow

| Construct              | Go                                              | GALA                                                     | Status      | Code       | Check                                                                    |
| ---------------------- | ----------------------------------------------- | -------------------------------------------------------- | ----------- | ---------- | ------------------------------------------------------------------------ |
| `if` statement         | `if c { ... }`                                  | `if (c) { ... }`; the parenthesized condition is the norm | direct     | —          | transpile                                                               |
| `if` initializer       | `if err := f(); err != nil { ... }`             | a preceding `var err = f()`, then `if (err != nil)`        | workaround | `GALA-E0047` | `gala explain GALA-E0047`, then transpile the `if` with an initializer |
| `if` expression        | `func f() int { if c { return 1 }; return 0 }`   | `func f() int = if (c) { 1 } else { 2 }`                   | direct     | —          | transpile; the condition must be parenthesized                           |
| `switch`                | `switch x { case 0: ... }`                      | `x match { case 0 => ...; case _ => ... }`                  | workaround | parse error | transpile a `switch` with one case, then the same cases as a `match`  |
| `switch` on an identifier | `switch s { case StatusOK: ... }`             | `s match { case StatusOK => ...; case _ => ... }`; a capitalized in-scope name or a qualified name is a stable identifier and compares with `==`, while a lowercase name binds | direct | parse error | transpile a `match` whose case names a capitalized package `var` and read the emitted `obj == Name` |
| `fallthrough`          | `case 0: fallthrough`                            | one arm per pattern, or one arm with a guard; `\|` is an alternative pattern that shares one arm's body rather than falling through (see the next row) | defect     | `GALA-E0017` | transpile `fallthrough` inside a `match` arm                            |
| several values in one `case` | `case 1, 2:`                               | `case 1 \| 2 => ...`; alternatives are tried left to right and nest in extractors and tuples, while `GALA-E0071` rejects an alternative that binds a name or is `_`, `\|` mixed with `+`, `-`, or `^` without parentheses, and `\|` inside a comparison or boolean expression of a pattern | direct     | `GALA-E0071` | transpile `case 1 \| 2` and read the emitted `obj == 1 \|\| obj == 2`; transpile `case x \| y` and read the `GALA-E0071` rejection |
| `match` guard           | `case x > 10:` in a tagless `switch`              | `n match { case x if x > 10 => ...; case _ => ... }`              | direct     | —          | transpile and read the emitted `if true && x > 10`                       |
| `select`               | `select { case v := <-ch: ... }`                 | `concurrent.Future` with a timeout, or `go_interop`          | answered   | parse error | transpile a `select`                                                    |
| `goto`                 | `goto end`                                      | structured control flow                                     | answered   | parse error | transpile a labelled statement                                          |
| `for` counter           | `for i := 0; i < n; i++ { ... }`                 | `for i := 0; i < n; i++ { ... }`                                 | direct     | —          | transpile; `++` and `--` need a mutable binding                      |
| `for` with omitted init | `for ; cond; post { ... }`                       | the same                                                      | direct     | —          | transpile a three-clause `for` whose init slot is empty                |
| `range`                | `for i, v := range xs`                            | the same                                                      | direct     | —          | transpile a `range` over a slice                                       |
| `defer`                | `defer f.Close()`                                | `use f = acquire`, or a `resource` combinator                 | workaround | `GALA-E0036` | `gala explain GALA-E0036`, then transpile the `defer`                 |
| `go` statement          | `go f()`                                         | `go_interop.Spawn(() => f())`, or a `concurrent.Future`       | workaround | `GALA-E0036` | `gala explain GALA-E0036`, then transpile the `go` statement          |
| `break` and `continue`  | `break`                                          | the same                                                      | direct     | —          | transpile a loop with a `break`                                         |
| `:=` reassignment       | `s = s + 1` where `s` came from `:=`              | a new name, or a `var` binding                                | semantic   | —          | transpile an assignment to a `:=`-bound name                            |
| `match` default         | a `switch` with no `default`                      | `case _ => ...` on a non-sealed value; a sealed type's exhaustive match needs no default | workaround | `GALA-E0003` | `gala explain GALA-E0003`, then transpile the same match with and without `case _` |
| `match` on a struct     | `switch p.Type { case ...: }`                     | a struct pattern, which calls its synthesized `Unapply`         | workaround | —          | transpile a struct pattern on a `val` field and on a `var` field separately |

A value-used `match` or `if` expression lowers to an immediately invoked Go closure with no runtime import, as long as its patterns are literals, stable identifiers, guards, or `_`; a typed pattern or an extractor pattern names the runtime.
Read the import block before relying on that, because one typed arm changes the whole file's imports.

`++` and `--` transpile, but only onto a mutable binding: a `val` or a `:=` binding is immutable and rejects them.
The `:=` in a `for` init slot is the exception that reads as inconsistent, because it is mutable there while a plain `:=` binding in a body is not.

## Builtins And Expressions

| Construct                | Go                    | GALA                                                          | Status      | Code       | Check                                                                    |
| ------------------------ | --------------------- | ------------------------------------------------------------- | ----------- | ---------- | ------------------------------------------------------------------------ |
| `len` on a string's bytes | `len(b)`              | `b.ByteSize()`, which emits Go's `len`                        | workaround | `GALA-E0035` | `gala explain GALA-E0035`, then transpile `.ByteSize()` and read the emitted `len` |
| `len` on characters      | `len([]rune(s))`      | `s.Size()`, which emits `utf8.RuneCountInString`              | workaround | `GALA-E0035` | transpile `.Size()` on a non-ASCII string and read the emitted call against `.ByteSize()` |
| `len` on a Go slice      | `len(xs)`             | `xs.Size()`, which emits Go's `len`                           | workaround | `GALA-E0035` | transpile `.Size()` on a Go slice and read the emitted call                |
| `len` on a Go slice whose type was inferred | `var xs, err = f(); len(xs)` | `go_interop.SliceFrom(xs, 0).Size()`, which is a view and allocates nothing | workaround | `GALA-E0035` | transpile the `SliceFrom` spelling on a receiver from `var a, b = f()` and read the emitted `len(go_interop.SliceFrom(...))` |
| `len` on a GALA collection | `len(xs)`           | `xs.Size()`, which stays a method call on the collection       | workaround | `GALA-E0035` | transpile the same call on a `HashMap` and read the difference             |
| `make` a slice           | `make([]T, n)`        | `go_interop.SliceWithSize[T](n)`, or `SliceWithCapacity[T](n)` | workaround | parse error | transpile `make`, then the helper                                         |
| `make` a map             | `make(map[K]V)`       | `go_interop.MapEmpty[K, V]()`, or `MapWithCapacity[K, V](n)`  | workaround | parse error | transpile `make`, then the helper                                         |
| `new`                    | `new(T)`              | `go_interop.New[T]()`, or a zero value                           | workaround | `GALA-E0035` | `gala explain GALA-E0035`, then transpile `new`                          |
| `append` on a Go slice   | `append(xs, v)`       | `go_interop.SliceAppend(xs, v)`; the result is a new slice     | workaround | `GALA-E0035` | transpile `append`, then the helper with the binding written out        |
| `append` on a collection | `append(xs, v)`       | `xs.Append(v)`                                                 | workaround | `GALA-E0035` | transpile the method form on an `Array`                                  |
| `delete`                 | `delete(m, k)`        | `go_interop.MapDelete(m, k)`, or `m.Remove(k)` on a `HashMap`  | workaround | `GALA-E0035` | transpile `delete`, then each form                                        |
| `close`                  | `close(ch)`           | `go_interop.CloseChan(ch)`, or `CloseSignal(s)`                | workaround | `GALA-E0035` | transpile `close`, then each form                                         |
| `copy`                   | `copy(dst, src)`      | `go_interop.SliceCopy(src)`; it returns a new slice             | workaround | `GALA-E0035` | transpile `copy`, then the helper                                         |
| `cap`                    | `cap(s)`              | `go_interop.SliceCap(s)`                                       | workaround | `GALA-E0035` | transpile `cap`, then the helper                                          |
| `panic`                  | `panic("boom")`       | `Try(...)` for a recoverable failure, or `go_builtins.Panic` in statement position | workaround | `GALA-E0035` | transpile `panic`, then each form in statement position             |
| `recover`                | `recover()`           | `Try(...)` captures the panic as a value but cannot resume in place | answered | `GALA-E0035` | transpile `recover`                                                      |
| `[]byte(s)`              | `[]byte(s)`           | `go_interop.ToBytes(s)`; `string(b)` needs no helper            | workaround | `GALA-E0040` | `gala explain GALA-E0040`, then transpile both directions               |
| slice expression         | `args[from:to]`       | `go_interop.Slice`, `SliceFrom`, `SliceTo`, `SliceTake`, `SliceDrop` | workaround | parse error | transpile one slice expression, then the helper for its arity         |
| string truncation        | `value[:n]`           | `string(go_interop.SliceTake(go_interop.ToBytes(s), n))` for a byte count, or a rune-accumulating loop for a character count | workaround | parse error | transpile `value[:n]`; then decide which count the original meant: the `ToBytes`/`SliceTake` form copies bytes exactly, while `out += string(s[i])` re-encodes a byte at or above 0x80 rather than copying it |
| map index                | `m[k]`                | `m[k]` on a Go map; `m.Get(k)` on a `HashMap`                   | direct     | —          | transpile an index on each container type                                |
| type assertion           | `s, ok := v.(string)` | `v match { case s: string => s; case _ => "" }`, which emits `std.As` and imports the runtime | workaround | parse error | transpile the assertion, then the type-pattern `match`, and grep the import block |
| function-type field read | `h.Execute()`         | the same, when the struct is a GALA declaration                 | direct     | —          | transpile; a struct declared in a handwritten sibling behaves differently |
| string interpolation     | `fmt.Sprintf("a %s", b)` | `s"a $b"` or `f"$b%.2f"`; `s"..."` picks the verb from the value's type (`%d` for an integer, `%v` for a string field) | direct     | —          | transpile both interpolation forms and read the emitted verbs against the original format string |
| comparison and arithmetic | `a == b`, `x + 1`     | the same                                                         | direct     | —          | transpile; `==` on a collection value is not the Go meaning            |

A bare builtin is a hard error rather than a style violation, and the check is resolver-aware: a name the program declares itself is left alone.
Nearly every substitute in this table lives in a package `gala doc` cannot describe, so read the helper list from the transpiler's own diagnostic hint instead; [Pinning A Row](#pinning-a-row) says why that package is the exception.

`.Size()` is not one spelling with one meaning: it lowers to a rune count on a string, to Go's `len` on a Go slice, and to a method call on a GALA collection.
A row that says `.Size()` without naming the receiver is not a row, because the two lowerings count different things and the one that is correct depends on which the Go code originally asked for.

The receiver's *spelling* decides whether the call lowers at all, independently of what the receiver is.
A receiver whose type is written out resolves even when the type is declared in another package as a GALA struct, and a receiver whose type is inferred from a call into a Go-declared function does not resolve even when that function returns a GALA-declared type in the same package.
So the same `len` over the same slice is a `workaround` or a wall depending only on how the binding was written.
`go_interop.SliceFrom(xs, 0).Size()` reaches the length in both cases, so this is a substitute and not a gap; a multi-value binding cannot be annotated because the grammar's single `(type)?` slot sits after the whole name list, and a single-value binding can, so the annotation is not the only door.
`go_interop` ships `MapLen` and `SliceCap` and no length helper for a slice or a string, so a `string` receiver from a Go sibling has no zero-cost spelling.
The same warning applies to a member the program does not have: a missing method is refused with a diagnostic that enumerates what the type does declare, so read that list rather than guessing the name.

## Containers

The container type decides which spelling applies, not whether the file is styled as Go-shaped or GALA-shaped.

| Container            | Construct                                       | GALA                                                                        | Status      | Code       | Check                                                                    |
| -------------------- | ----------------------------------------------- | --------------------------------------------------------------------------- | ----------- | ---------- | ------------------------------------------------------------------------ |
| Go slice             | `[]T{a, b}`                                     | `go_interop.SliceOf(a, b)`, `SliceEmpty[T]()`, or `SliceWithSize[T](n)`      | workaround | `GALA-E0007` | `gala explain GALA-E0007`, then transpile the literal and the helpers   |
| Go slice             | `[]T{}`                                         | `go_interop.SliceEmpty[T]()`                                                 | workaround | `GALA-E0007` | transpile the empty literal and the helper                                |
| `Array`              | `[]T{a, b}`                                     | `ArrayOf(a, b)`, or `EmptyArray[T]()`                                        | workaround | `GALA-E0007` | transpile both container types with the same elements                     |
| Go map               | `map[K]V{"a": x}`                               | `go_interop.MapEmpty[K, V]()` then `MapPut`, one call per entry             | workaround | `GALA-E0008` | `gala explain GALA-E0008`, then transpile the literal and the helpers   |
| `HashMap`            | `map[K]V{"a": x}`                               | `EmptyHashMap[K, V]().Put("a", x)`                                          | workaround | `GALA-E0008` | transpile both container types with the same entries                      |
| Go map lookup        | `m[k]`                                          | `m[k]`                                                                       | direct     | —          | transpile a lookup on a Go map                                            |
| Go map lookup with comma-ok | `v, ok := m[k]`                          | `go_interop.OptionFromMap(m, k)` then `match`/`ForEach`/`GetOrElse`; annotate the lambda parameter when the map's type is not resolvable from the `.gala` file | workaround | — | transpile the comma-ok lookup and the `OptionFromMap` form and read the emitted `std.Option`; for a map whose type came from a handwritten sibling, read the leaked `T` in the unannotated lambda |
| `HashMap` lookup     | `m[k]`                                          | `m.Get(k)`, or a `Some`/`None` match                                         | workaround | —          | transpile the lookup and read the emitted spelling                        |
| map or slice type in an expression | `EmptyHashMap[string, []byte]()`   | name a GALA type such as `Array[T]` or `HashMap[K, V]`; text is `string`     | workaround | `GALA-E0040` | `gala explain GALA-E0040`, then transpile the expression                 |

## Errors And Failure

| Construct                   | Go                                    | GALA                                                          | Status      | Code       | Check                                                                 |
| --------------------------- | ------------------------------------- | ------------------------------------------------------------- | ----------- | ---------- | --------------------------------------------------------------------- |
| error check after a call    | `if err != nil { return err }`        | `if (err != nil) { ... }` with a preceding `var` binding        | direct     | —          | transpile the check                                                |
| failure as a value          | `v, err := f(); if err != nil {...}`  | `f()` used as a value is already a `Try[T]` (the transpiler emits `std.GoTry`); take it apart with `match`, or handle the ends with `OnFailure`/`OnSuccess`/`GetOrElse` | direct | — | transpile the call under `match`, `OnFailure`, and `GetOrElse` and read the emitted `std.GoTry`; `Try(f())` instead emits `std.Try[T]{}.Apply` over a panicking closure, so it is the panic-catching form rather than this one |
| failure as a value, void call | `err := f()`                        | `FromError(f())`                                                | direct     | —          | transpile `FromError(f())` and `Try(f())` and read `std.FromError` against `std.Try[std.Void]{}.Apply`; the `Try` form catches a panic the call itself raised, which `FromError` propagates |
| a Go call's plain value     | `n := f()` then `n + 1`               | the call is one GALA value; unwrap it with `Get` or match on it  | workaround | `GALA-E0049` | `gala explain GALA-E0049`, then use the value as a plain value     |
| a Go call's `Try` or `Tuple` | `val t = f()` then `t.Take(0)`      | `Try` has no `Take`; a Go multi-result call is one value         | answered   | `GALA-E0044` | `gala explain GALA-E0044`, then transpile the method call         |
| `panic` in a tail position  | `func f() int { panic("x") }`        | `go_builtins.Panic("x")` as a statement, then return            | workaround | `GALA-E0035` | transpile the statement form and the tail form separately         |
| `recover` in a handler      | `if r := recover(); r != nil {...}`  | `Try(...)` per iteration; there is no in-place resume            | answered   | `GALA-E0035` | transpile `recover`                                                  |

`Try` and `FromError` are not interchangeable for a void Go call, and the choice is observable at run time rather than at transpile time, so it is worth a real run rather than a transpile alone.

## Concurrency And Resources

| Construct             | Go                                | GALA                                                                  | Status      | Code       | Check                                                                       |
| --------------------- | --------------------------------- | --------------------------------------------------------------------- | ----------- | ---------- | --------------------------------------------------------------------------- |
| scoped cleanup        | `defer f.Close()`                 | `use f = acquire` after a single-value acquire                        | workaround | `GALA-E0036` | transpile the `use` form and read the emitted `defer`                   |
| cleanup after a checked acquire | `defer f.Close()` | `resource.Using[R, A](res, (x) => ...)`; a GALA-declared resource type infers both arguments, a Go-declared one needs them written out | workaround | `GALA-E0036` | transpile the same body over a GALA-declared and over a `.go`-declared resource type, with and without the type arguments, and read the emitted lambda parameter |
| non-`Closeable` cleanup | `defer os.RemoveAll(d)`          | `resource.Bracket(resource, release, body)`                           | workaround | `GALA-E0036` | transpile the `Bracket` form                                             |
| critical section      | `mu.Lock(); defer mu.Unlock()`     | `resource.Bracket(init, release, body)` or `resource.WithLock`         | workaround | `GALA-E0036` | transpile the `WithLock` form                                           |
| goroutine             | `go f()`                           | `go_interop.Spawn(() => f())`                                         | workaround | `GALA-E0036` | transpile the `Spawn` form                                              |
| one async result      | `ch := make(chan T, 1); <-ch`      | `concurrent.Future(f())` then `Await` or `Map`                        | workaround | `GALA-E0035` | transpile the `Future` form                                             |
| timeout               | `select { case <-time.After(d): }`  | `future.WithTimeout(d)`, or `future.AwaitFor(d)`, or `FirstCompletedOf` | workaround | parse error | `gala doc concurrent`, then transpile a `select` and the timeout form |
| capture across a boundary | `go func(){ use mu, xs }()`     | snapshot the captured state into a `val` first                        | workaround | `GALA-E0037` | `gala explain GALA-E0037`, then transpile a closure that captures a mutable |
| a lock held in a `val` field | `struct C(mu sync.Mutex)`     | declare the field `var`                                                | workaround | `GALA-E0053` | `gala explain GALA-E0053`, then transpile a pointer method on it      |
| a method on a `use`-bound value | `f.Name()` after `use f = os.Open(p)` | handle the error first, then bind with `use`                     | answered   | `GALA-E0044` | transpile the two-step form                                            |

`use` accepts a single-value acquire, so a call that returns `(T, error)` has to have its error handled before the binding, and the resulting binding is a `Try` rather than the value, which is why calling a method on it fails.

A resource combinator's result type argument is optional in practice, and what is not optional is that the resource type be resolvable from `.gala` text.
A GALA-declared resource type infers both arguments with none written out, which is how upstream's own example calls it, and `WithLock` infers either way; a resource type declared in a handwritten `.go` sibling binds the body parameter to `any` and does not build.
The enclosing function's return type is not stripped by any of that — it is emitted unchanged in every variant checked — so a claim that the omission costs the signature is wrong, and the general defect there is `func F() = <expr>` never inferring a return type at all.
A partial type-argument list is its own failure: it emits the transpiler's own type-parameter name as a declared Go type.
The body's parameter still binds correctly, which is what makes this hard to see in the source; transpile the form you intend to use and read the emitted function's own signature, because a clean transpile here is not evidence of a correct one.

The `Spawn` asymmetry is documented in `gala explain GALA-E0037`'s own escape-hatch section, in the runtime source, and in the concurrency-safety document; what is missing is the caveat in the best-practices document and in the forbidden-keyword table, which present `Spawn` as a plain drop-in for `go f()`, so there is nothing to file about the behaviour.
The capture guard is boundary-specific, and which boundary decides whether you get a diagnostic or a race.
A `concurrent.Future` body that captures a reassignable binding is refused before codegen, while `go_interop.Spawn` accepts the same capture, emits an unchecked goroutine, and lets the enclosing function return before the body has run.
Treat the two rows as independent and verify the spawn form by running the result under the race detector.

## Packages And Files

These rows are about a package rather than about a construct, and they only bite in a package that mixes a translated file with a handwritten sibling.

| Construct                                  | Go                                          | GALA                                             | Status      | Code          | Check                                                              |
| ------------------------------------------ | ------------------------------------------- | ------------------------------------------------ | ----------- | ------------- | ------------------------------------------------------------------ |
| an import the file does not declare         | `import "os"` in the file that uses it      | the same import in every file that uses it       | direct      | `GALA-E0023`  | transpile a file that omits an import its sibling declares, once with a `.gala` sibling and once with a handwritten `.go` one, and read the refusal each time |
| an unqualified GALA-runtime name the file does not import | `SliceOf(1)` with no `go_interop` import | the same call, plus the import that declares it | direct      | `GALA-E0023`  | transpile a bare runtime name in a file whose sibling dot-imports it, and read the hint that names the declaring package |
| a method on a GALA type declared in a handwritten sibling | `func (r Repo) Save()` in a `.go` sibling, `struct Repo()` in the `.gala` file | none: the method has to move into the `.gala` file, which usually means keeping the member handwritten | defect | `GALA-E0044` | transpile both, changing only the sibling's extension, and read the diagnostic whose hint claims the type declares no methods |
| a bare name in a declared-type position | `func size(r Response) int` where an import also exports `Response` | declare the type in the same `.gala` file, or name it through a handwritten constructor | defect | — | transpile with the declaration in a sibling `.gala` and a field that exists only on the local type, and read the emitted parameter type |
| a method call on a binding | `x := f()` then `x.M()` | `var x = f()`; `val` mislowers the same way | defect | — | transpile the two spellings over a pointer-receiver method, assert `std.NewImmutable` and `AddrOfCopy` are absent from the `var` output, and read the `go build` error for the other |
| a declaration in a handwritten sibling      | `helper()` defined in the `.go` file        | referenced by name from the `.gala` file         | direct      | —             | transpile both files; the name resolves, and the members behind it are the transpiler's problem |
| an import alias                             | `store "example.com/project/store"`         | the same; a call whose signature the transpiler resolves needs the qualifier to name the package, because a differently named alias loses the signature | direct | — | transpile a `(T, error)` call behind a differently named alias with an annotated `OnFailure` lambda and read the call without `std.GoTry`; re-spell the alias as the package's own name and read `std.GoTry` |

A name declared in another file of the package — `.gala` or `.go` — is invisible to the transpiler's type resolution, and the two directions of that are not symmetric: an unknown receiver passes through and Go resolves it, while a GALA-declared receiver is judged against a method set that excludes a sibling's.
A package's imports are checked one file at a time and emitted per file, so an import one file omits is refused at the transpile rather than left for the Go compiler, and that holds whether or not a sibling declares it and whichever language that sibling is written in.
An alias is direct, and a sibling declaration is visible as a name; neither of those two rows is a claim that the transpiler knows what the name is.
In particular, a call behind an alias whose name differs from the package's own name loses its resolved signature: on the pinned rev, `pgstore "timeful/server/postgres"` made `pgstore.DefaultRepository()` an unknown type, so `.OnFailure((err) => ...)` was refused with `GALA-E0033` unannotated and emitted the call without `std.GoTry` when the lambda was annotated, which `go build` rejects, while `postgres "timeful/server/postgres"` resolved and emitted `std.GoTry` (checked 2026-10-10; `server/discord_bot/init.gala` and the bot command twins dropped the `pgstore` spelling for this reason).

## Pinning A Row

A construct verdict is a property of the compiler on `PATH`, and it changes between releases.
Re-derive the rows your translation depends on rather than trusting this file, using the cheapest check that answers the question:

```sh
gala version                          # what the rows below were checked against
gala explain --list                   # every diagnostic code the compiler knows
gala explain GALA-Exxxx               # the full page for one code, including a minimal repro
gala transpile -i <file>.gala -o <scratch>.go   # the verdict for one construct
gala doc <package>                    # what a GALA-source package exports
```

`gala doc` describes GALA-source packages only, so a package whose source is handwritten Go is not found by it and its helpers have to be read from a transpiler diagnostic instead.
The two packages this costs you the most are the interop ones: they are handwritten Go where every other runtime package has GALA sources, and they are the packages that supply the substitute for nearly every builtin row in this file, so a reader sent to `gala doc` for a helper list arrives at a package that is not there.
`gala transpile` to a scratch path rather than to `/dev/stdout`, which the writer cannot open on every platform.
To test a shape claim rather than an acceptance claim, read the emitted Go; to test an acceptance claim, build and run it, because a wrapper that is invisible to the program is still a wrapper.
