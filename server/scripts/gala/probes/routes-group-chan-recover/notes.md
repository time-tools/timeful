# routes-group-chan-recover

`routes/group.go`'s `getCalendarAvailabilities` fans out one goroutine per response, collects results through a buffered `chan calendarResult`, and replaces a panicking request's result with an empty one through an in-place `recover`.
GALA has no `chan` type in any position, its `go` statement is a forbidden keyword, and `recover` cannot resume in place.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile -i main.gala -o /tmp/probe-routes-group-chan-recover.go
```

## Observed

```text
error: no viable alternative at input '(chanresult,len('
  --> main.gala:12:34
   |
12 | 	results := make(chan result, len(requestIDs))
   | 	                                ^
   |
```

A minimal `func f(ch chan int) int { return <-ch }` fails at the signature:

```text
error: extraneous input 'int' expecting ')'
  --> main.gala:3:16
   |
 3 | func f(ch chan int) int {
   |                ^^^
   |
```

`defer func() { ... }()` and the `go` statement fail as forbidden statement keywords (`GALA-E0036`), and the in-place `recover` has no resume form:

```text
error: no viable alternative at input 'func(){deferfunc'
  --> main.gala:16:10
   |
16 | 			defer func() {
   | 			      ^^^^
   |

error: no viable alternative at input 'func(){ifrecovered'
  --> main.gala:17:8
   |
17 | 				if recovered := recover(); recovered != nil {
   | 				   ^^^^^^^^^
   |
```

## Expected Go shape

```go
results := make(chan calendarResult, len(requests))
for _, request := range requests {
	request := request
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				results <- calendarResult{publicID: request.publicID}
			}
		}()
		events, _ := calendar.GetUsersCalendarEvents(request.user, request.enabledSet, query.TimeMin, query.TimeMax)
		results <- calendarResult{publicID: request.publicID, events: events}
	}()
}
for i := 0; i < len(requests); i++ {
	calendarEvents := <-results
	...
}
```

The fan-out, the result order, and the panic-to-empty-result behavior are all part of the handler's observable contract.

## Classification

Answered rows (`references/gaps.md`, `constructs.md`): `chan` has no GALA spelling (`concurrent.Future` or `go_interop` for a boundary-free result), `go` maps to `go_interop.Spawn`, and `recover` maps to `Try`, which captures a panic as a value but cannot resume in place.

## Upstream

[#528](https://github.com/martianoff/gala/issues/528) carried the triage and closed upstream on 2026-10-04; `gaps.md` says not to file channel types or `recover`.
No specific follow-up report was found.

## Compiler

- `gala version`: `GALA version 0.87.1`
- flake rev: `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c`
- extraction: `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa`
- date: 2026-10-10

## Re-verified (2026-10-18)

- `gala version`: `GALA version 0.90.0`
- flake rev: `014c2f23e45a2cd631b73cab38c0bdcb5e773347`
- extraction: `0.90.0 bd0c1591b19ce9971d64efd52406bab21723db64807b526c0f716102453bdd90`
- Result: the same refusals for `chan`, `defer func()`, `go`, and `recover`; the expected-token lists now include the new `default` keyword.

## Re-verified (2026-10-21)

- `gala version`: `GALA version 0.92.0`
- flake rev: `4bd2d4e89cb8699e94b1f209356dd679db8d54ac`
- extraction: `0.92.0 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same refusals for `chan`, `defer func()`, `go`, and `recover`; the expected-token lists still name the `default` keyword.

## Re-verified (2026-10-24)

- `gala version`: `GALA version 0.92.1`
- flake rev: `911cc05932486626e684931aac9f1707939a91e6`
- extraction: `0.92.1 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same refusals for `chan`, `defer func()`, `go`, and `recover`; the expected-token lists still name the `default` keyword.

## Re-verified (2026-10-25)

- `gala version`: `GALA version 0.93.0`
- flake rev: `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`
- extraction: `0.93.0 6368c09381c6f9d8d72f1551101ca48207cb2e1c34122627ca928261ad05b13c`
- Result: the same refusals for `chan`, `defer func()`, `go`, and `recover`; the expected-token lists still name the `default` keyword.

## Re-verified (2026-10-29)

- `gala version`: `GALA version 0.95.0`
- flake rev: `173f7d778f103a05553564d841e0daba0126cb16`
- extraction: `0.95.0 7b68112a768d6d4776444c8ab561ce629c219d86748f93f0ca1f485daf7d1648`
- Result: the same refusals for `chan`, `defer func()`, `go`, and `recover`, in seven diagnostics.
  The `make(chan result, len(...))` line now reports `mismatched input ')'` with an expected-token list that includes `default` where the 0.93.0 note recorded a second `chan` diagnostic at the parameter position, and the error-recovered tree adds the enclosing `defer`/`if` lines; the refused constructs are unchanged.
