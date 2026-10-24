### `resource.Bracket` / `Using` re-raise a body panic as an unexported `*std.panicError` instead of the original value

A `panic` inside a `Bracket` body reaches the caller's `recover` as `*std.panicError` rather than as the value that was panicked.
Go's `defer` re-raises the original value, and `Bracket`'s own doc comment says the same thing, so a mixed package that uses `Bracket` for cleanup and a handwritten `recover` for recovery observes a different panic value than the Go spelling would.

**Classification:** defect

**What I tried**

`resource.Bracket(1, (r int) => {}, body)` where `body` calls `go_builtins.Panic("boom")`.
The transpile is clean, the release lambda runs, and the emitted Go is `resource.Bracket(1, func(r int) {}, body)`.
An outer `recover` in a handwritten Go function of the same package then reports `*std.panicError` holding `"boom"`.

**Diagnostic**

No diagnostic; the transpile and the build both succeed.
The observable artifact is the recovered value's type:

```text
recovered type: *std.panicError
value: boom
```

**Minimal repro**

`go.mod`:

```text
module brackettest

go 1.25.0

require go.gala.fyi/stdlib v0.90.0
```

`main.gala`:

```gala
package main

import "go.gala.fyi/stdlib/resource"
import "go.gala.fyi/stdlib/go_builtins"

// Body panics; Bracket is supposed to release and re-raise.
func run() int {
	var body func(int) int = (n) => {
		go_builtins.Panic("boom")
		return n
	}
	return resource.Bracket(1, (r int) => {}, body)
}
```

`harness.go`, handwritten in the same package because in-place `recover` is itself unavailable from GALA:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	defer func() {
		r := recover()
		if r == nil {
			fmt.Println("no panic")
			return
		}
		fmt.Printf("recovered type: %T\nvalue: %v\n", r, r)
		os.Exit(0)
	}()
	fmt.Println(run())
}
```

```sh
cd main && gala transpile --stdlib-module go.gala.fyi/stdlib -i main.gala -o main_gen.go
go mod tidy && go run ./main
# recovered type: *std.panicError
# value: boom
```

The emitted call in `main_gen.go` is unremarkable:

```go
return resource.Bracket(1, func(r int) {
}, body)
```

**What a caller would need**

`Bracket` and `Using` should re-panic with the value the body panicked with, so that an outer `recover` sees `"boom"` exactly as it would under `defer`.
The wrapper itself is currently the obstacle to that: `recoveredError` converts a `string` panic to `errors.New(e)`, and the resulting `*panicError` is unexported with no exported accessor that returns the original value, so a Go caller cannot unwrap back to what was panicked and cannot even name the type to assert on.

`std/types.go` documents the wrapper as deliberate — it carries the panic's stack, and `recoveredError` deliberately returns an existing `*panicError` unchanged so a re-raised panic keeps its original frames.
That is a good reason for `Try` to keep a wrapped error as its `Failure`, and it does not require `Bracket` to put the wrapper where Go's own `recover` expects the raw value: a re-raise can carry the stack in one of the ways `panic` already supports without changing the value the caller observes.
A caller that only inspects the type breaks either way, and `Bracket`'s comment promises the Go behavior, so the current shape is the one that needs the justification.

**Compiler**

`GALA version 0.91.0` (flake rev `a520615610428cd3a152a53e04edf53c71f7c179`).
The behavior was first recorded on `0.87.1` (rev `cd2fdcb5`) and is unchanged.

**Environment**

Go `go1.26.1 linux/amd64`; the program builds and runs with `go.gala.fyi/stdlib v0.90.0`.
Worked around by hand, the same program with a plain `defer` and no `Bracket` re-raises `"boom"` as a `string`, which is the contrast that makes the report legible.

**Searched**

`gala explain --list` and the `GALA-E0036` page (the `defer`/resource code, which names `Bracket` as the replacement and does not mention the panic value).
The specification and the best-practices document for the `resource` combinators, and `resource/resource.gala`'s own comment, which says release runs "even if body panics (release runs, then the failure re-raises)".
The upstream tracker for `panicError`, `Bracket panic`, `recover`, and `resource` — no existing issue.
`resource/resource_test.gala` contains no panic test, so nothing upstream pins the current behavior as intended.
