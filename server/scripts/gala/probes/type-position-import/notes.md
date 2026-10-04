# Type-position import hole probe

An unimported GALA runtime type name in a type position is not refused at transpile time, so the emitted Go carries the bare name and `go build` blames the generated file through its `//line` directive.
This is the finding tracked as [#648](https://github.com/martianoff/gala/issues/648) and reported from TASK-0340.

## Invocation

From this directory:

```sh
gala transpile -i main.gala -o /tmp/type-position-import/main.go
```

The transpile exits `0` with no diagnostic.
To see the failure, transpile into a scratch module whose `go.mod` resolves `martianoff/gala` to the vendored runtime and build it:

```sh
mkdir -p /tmp/type-position-import && cd /tmp/type-position-import
cp <repo>/server/scripts/gala/probes/type-position-import/main.gala .
gala transpile -i main.gala -o main.go
go mod init probe
go mod edit -require=martianoff/gala@v0.0.0 -replace=martianoff/gala=<repo>/server/third_party/gala
go build ./...
```

## Observed on the pinned compiler

Compiler: `GALA version 0.84.1`, flake rev `e2e28c318ff1eb606f7f607199b629d93b78ab1f`, checked 2026-10-04.
Transpile: clean, exit `0`.
Build: exit `1` with `main.gala:4: undefined: Future`, `main.gala:13: undefined: Future`, `main.gala:20: undefined: Future`, `main.gala:5: undefined: Future`, and `main.gala:7: undefined: Future`.
The generated Go keeps `Future[int]` in the struct field, the package variable, and the parameter, and imports only `martianoff/gala/std`.

## Expected shape

Either a transpile-time refusal, in the shape the value-position check already produces with `GALA-E0023`, or a diagnostic naming the package that declares `Future` and the import the file is missing.

## Upstream

Filed as [#648](https://github.com/martianoff/gala/issues/648) on 2026-09-30 and closed on 2026-10-03, so the fix exists upstream but is not in the pinned rev.
A sync to a flake rev at or after the fix retires this probe.
