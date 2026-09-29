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
| defined type              | `type CalendarType string`                      | `type CalendarType string`; conversions work                     | direct     | —          | transpile a conversion such as `int64(Millis(5))`                        |
| method on a defined type  | `func (d DateTime) IsZero() bool`               | a struct wrapping the value, or a plain function               | answered   | `GALA-E0048` | `gala explain GALA-E0048`, then transpile a method on a named scalar      |
| struct                    | `type Person struct { Name string }`            | `struct Person(Name string)`, one field per line or comma        | workaround | —          | transpile; non-`var` fields become `std.Immutable[T]`                     |
| struct with `var` fields  | `type Person struct { Name string }`            | `struct Person(var Name string)`                                | direct     | —          | transpile; a `var` field stays a plain Go field                          |
| generic struct            | `type Pair[T any] struct { A, B T }`            | `struct Pair[T any](var A T, var B T)`                          | direct     | —          | transpile; a generic struct gains an extra `Instance` interface and `Is<T>()` method |
| struct tag                | ``Name string `json:"name"` ``                  | `Codec[T]` with `.Rename`, `.Omit`, `.OmitEmpty`               | answered   | parse error | transpile a tagged field; backtick tags are not in the grammar           |
| embedded field            | `type Event struct { Base; Name string }`       | a named field plus explicit delegation, losing promotion        | answered   | parse error | transpile both the bare and the `embed` spellings                        |
| anonymous struct type     | `struct { A string }`                           | a named struct                                                   | workaround | parse error | transpile an anonymous struct in a type position                         |
| `struct{}` in type position | `map[T]struct{}`                             | a named empty struct, which changes the element type identity    | workaround | parse error | transpile `struct{}` as a map value type                                 |
| interface declaration     | `type Repo interface { Get(id string) string }` | `type Repo interface { Get(id string) string }`                  | direct     | —          | transpile; declare it with `type X interface`, not a bare `interface`    |
| `interface{}`             | `Details interface{}`                           | `Details any`                                                    | workaround | parse error | transpile both spellings; only `any` parses                              |
| `chan` type               | `func send(ch chan int)`                        | `concurrent.Future` or `go_interop` where no channel type crosses the boundary | workaround | parse error | transpile a channel type in a signature, a field, or a local |
| fixed-size array type     | `[16]byte`                                      | `Array[T]`, or a struct when the arity carries meaning         | answered   | parse error | transpile an array type in any position                                  |
| `map` type in type position | `map[K]V`                                    | `map[K]V`                                                       | direct     | —          | transpile a field, a parameter, or an alias                              |
| generic declaration       | `type Set[T comparable] ...`                    | `type Set[T comparable] ...`; struct generics are `struct Box[T any](var V T)` | direct     | —          | transpile a generic function and a generic struct separately              |
| grouped parameters        | `func add(a, b int) int`                        | `func add(a int, b int) int`                                     | workaround | `GALA-E0034` | `gala explain GALA-E0034`, then transpile a grouped parameter list      |
| multi-value return        | `func f() (T, error)`                           | `Tuple[T, error]` for GALA-internal callers                      | workaround | parse error | transpile a multi-value signature                                        |
| type arguments            | `MapEmpty[string, int]()`                       | comma-separated, never inferred where nothing pins them          | direct     | —          | transpile and read the emitted type arguments                            |

The `var` versus `val` and `var` versus `:=` distinctions are shape rules, not preferences, and they are the first thing to get right in a file that Go code outside the package can see.
`references/gaps.md` and the traps section of `SKILL.md` say what each one costs.

A rejection without a code is still a rejection: some checks report a plain error rather than a `GALA-Exxxx` page, so read the message rather than assuming a code is always there.

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
| `switch` on an identifier | `switch s { case StatusOK: ... }`             | an `if`/`else` comparison chain; a `case` naming a constant binds instead of comparing | workaround | parse error | transpile a `match` whose case names a declared constant             |
| `fallthrough`          | `case 0: fallthrough`                            | combine the patterns in one arm with `\|`, or restructure     | defect     | `GALA-E0017` | transpile `fallthrough` inside a `match` arm                            |
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

`++` and `--` transpile, but only onto a mutable binding: a `val` or a `:=` binding is immutable and rejects them.
The `:=` in a `for` init slot is the exception that reads as inconsistent, because it is mutable there while a plain `:=` binding in a body is not.

## Builtins And Expressions

| Construct                | Go                    | GALA                                                          | Status      | Code       | Check                                                                    |
| ------------------------ | --------------------- | ------------------------------------------------------------- | ----------- | ---------- | ------------------------------------------------------------------------ |
| `len` on bytes           | `len(b)`              | `b.ByteSize()`                                                 | workaround | `GALA-E0035` | `gala explain GALA-E0035`, then transpile `len` on a string and a slice |
| `len` on characters      | `len([]rune(s))`      | `s.Size()`                                                     | workaround | `GALA-E0035` | transpile `.Size()` on a non-ASCII string and compare with `.ByteSize()` |
| `len` on a collection    | `len(xs)`             | `xs.Size()`                                                    | workaround | `GALA-E0035` | transpile `.Size()` on a slice                                          |
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
| map index                | `m[k]`                | `m[k]` on a Go map; `m.Get(k)` on a `HashMap`                   | direct     | —          | transpile an index on each container type                                |
| type assertion           | `s, ok := v.(string)` | `v match { case s: string => s; case _ => "" }`                 | workaround | parse error | transpile the assertion, then the type-pattern `match`                |
| function-type field read | `h.Execute()`         | the same, when the struct is a GALA declaration                 | direct     | —          | transpile; a struct declared in a handwritten sibling behaves differently |
| string interpolation     | `fmt.Sprintf("a %s", b)` | `s"a $b"` or `f"$b%.2f"`                                       | direct     | —          | transpile both interpolation forms                                       |
| comparison and arithmetic | `a == b`, `x + 1`     | the same                                                         | direct     | —          | transpile; `==` on a collection value is not the Go meaning            |

A bare builtin is a hard error rather than a style violation, and the check is resolver-aware: a name the program declares itself is left alone.
`gala doc go_interop` does not work, so read the helper list from the transpiler's own diagnostic hint rather than from `gala doc`.

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
| `HashMap` lookup     | `m[k]`                                          | `m.Get(k)`, or a `Some`/`None` match                                         | workaround | —          | transpile the lookup and read the emitted spelling                        |
| map or slice type in an expression | `EmptyHashMap[string, []byte]()`   | name a GALA type such as `Array[T]` or `HashMap[K, V]`; text is `string`     | workaround | `GALA-E0040` | `gala explain GALA-E0040`, then transpile the expression                 |

## Errors And Failure

| Construct                   | Go                                    | GALA                                                          | Status      | Code       | Check                                                                 |
| --------------------------- | ------------------------------------- | ------------------------------------------------------------- | ----------- | ---------- | --------------------------------------------------------------------- |
| error check after a call    | `if err != nil { return err }`        | `if (err != nil) { ... }` with a preceding `var` binding        | direct     | —          | transpile the check                                                |
| failure as a value          | `v, err := f(); if err != nil {...}`  | `Try(f())` then `match` on `Success`/`Failure`, or `GetOrElse`   | direct     | —          | transpile both `match` and `GetOrElse` forms                        |
| failure as a value, void call | `err := f()`                        | `FromError(f())`                                                 | direct     | —          | transpile a void Go call through `FromError` and run it            |
| a Go call's plain value     | `n := f()` then `n + 1`               | the call is one GALA value; unwrap it with `Get` or match on it  | workaround | `GALA-E0049` | `gala explain GALA-E0049`, then use the value as a plain value     |
| a Go call's `Try` or `Tuple` | `val t = f()` then `t.Take(0)`      | `Try` has no `Take`; a Go multi-result call is one value         | answered   | `GALA-E0044` | `gala explain GALA-E0044`, then transpile the method call         |
| `panic` in a tail position  | `func f() int { panic("x") }`        | `go_builtins.Panic("x")` as a statement, then return            | workaround | `GALA-E0035` | transpile the statement form and the tail form separately         |
| `recover` in a handler      | `if r := recover(); r != nil {...}`  | `Try(...)` per iteration; there is no in-place resume            | answered   | `GALA-E0035` | transpile `recover`                                                  |

`Try` and `FromError` are not interchangeable for a void Go call, and the choice is observable at run time rather than at transpile time, so it is worth a real run rather than a transpile alone.

## Concurrency And Resources

| Construct             | Go                                | GALA                                                                  | Status      | Code       | Check                                                                       |
| --------------------- | --------------------------------- | --------------------------------------------------------------------- | ----------- | ---------- | --------------------------------------------------------------------------- |
| scoped cleanup        | `defer f.Close()`                 | `use f = acquire` after a single-value acquire                        | workaround | `GALA-E0036` | transpile the `use` form and read the emitted `defer`                   |
| cleanup after a checked acquire | `defer f.Close()`         | `resource.Using[*os.File, T](f, (x) => ...)`                          | workaround | `GALA-E0036` | transpile with and without explicit type arguments and compare         |
| non-`Closeable` cleanup | `defer os.RemoveAll(d)`          | `resource.Bracket(resource, release, body)`                           | workaround | `GALA-E0036` | transpile the `Bracket` form                                             |
| critical section      | `mu.Lock(); defer mu.Unlock()`     | `resource.Bracket(init, release, body)` or `resource.WithLock`         | workaround | `GALA-E0036` | transpile the `WithLock` form                                           |
| goroutine             | `go f()`                           | `go_interop.Spawn(() => f())`                                         | workaround | `GALA-E0036` | transpile the `Spawn` form                                              |
| one async result      | `ch := make(chan T, 1); <-ch`      | `concurrent.Future(f())` then `Await` or `Map`                        | workaround | `GALA-E0035` | transpile the `Future` form                                             |
| timeout               | `select { case <-time.After(d): }`  | `future.WithTimeout(d)`, or `future.AwaitFor(d)`, or `FirstCompletedOf` | workaround | parse error | `gala doc concurrent`, then transpile a `select` and the timeout form |
| capture across a boundary | `go func(){ use mu, xs }()`     | snapshot the captured state into a `val` first                        | workaround | `GALA-E0037` | `gala explain GALA-E0037`, then transpile a closure that captures a mutable |
| a lock held in a `val` field | `struct C(mu sync.Mutex)`     | declare the field `var`                                                | workaround | `GALA-E0053` | `gala explain GALA-E0053`, then transpile a pointer method on it      |
| a method on a `use`-bound value | `f.Name()` after `use f = os.Open(p)` | handle the error first, then bind with `use`                     | answered   | `GALA-E0044` | transpile the two-step form                                            |

`use` accepts a single-value acquire, so a call that returns `(T, error)` has to have its error handled before the binding, and the resulting binding is a `Try` rather than the value, which is why calling a method on it fails.

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
`gala transpile` to a scratch path rather than to `/dev/stdout`, which the writer cannot open on every platform.
To test a shape claim rather than an acceptance claim, read the emitted Go; to test an acceptance claim, build and run it, because a wrapper that is invisible to the program is still a wrapper.
