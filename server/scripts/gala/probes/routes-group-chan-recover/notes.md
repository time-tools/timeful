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
