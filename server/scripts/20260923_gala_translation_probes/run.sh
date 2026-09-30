#!/usr/bin/env bash
# Scripted probe corpus for the Go-to-GALA translation roster in docs/gala-translation.md.
#
# Each probes/<name>/ directory holds a main.gala and an expect file:
#   KIND=pass            transpile and go build must succeed; RUN=yes also runs the
#                        binary and diffs stdout against expected.out
#   KIND=transpile_fail  gala transpile must fail; CODE and ERR are checked against
#                        the diagnostic
#   KIND=build_fail      transpile must succeed and go build must fail on ERR
#   KIND=emit            transpile must succeed and the text of main.gen.go must
#                        satisfy every CONTAINS and ABSENT marker
#
# Which kind a new claim needs:
#   - a claim about what the program does            -> KIND=pass (add RUN=yes)
#   - a claim that something is rejected             -> KIND=transpile_fail
#   - a claim that the emitted Go does not compile   -> KIND=build_fail
#   - a claim about the shape of the emitted Go      -> KIND=emit
# A construct can be accepted with the wrong shape, so the two rejection kinds
# never carry a shape claim, and a wrapper the program never observes is invisible
# to an output assertion, so KIND=pass never carries one either.
#
# KIND=emit markers are literal substrings of main.gen.go, and both keys are
# repeatable, so one probe can pin several markers:
#   CONTAINS=<marker>  the marker must appear in the emitted Go
#   ABSENT=<marker>    the marker must not appear in the emitted Go
# A marker names a semantic thing: a wrapper constructor call, a synthesized
# interface or method, a method call on a binding. Never use whitespace,
# indentation, or a line break, because the emitted Go is formatted and carries
# //line directives naming the probe's own source.
# The polarity lives in the key name rather than in a separate mode key, so a
# marker needs no escaping and the expect file reads as an assertion list.
#
# Generated Go files (main.gen.go), runner-written fixtures, logs, and binaries are
# gitignored by explicit path in .gitignore; only the .gala sources, expect files,
# expected.out files, go.mod, .gitignore, and this runner are committed.
#
# The expectations pin GALA 0.84.1. Set GALA_PROBES_ALLOW_ANY_VERSION=1 to run against
# another compiler and see which expectations moved.
# The go build expectations also pin the Go 1.26 toolchain series; set
# GALA_PROBES_ALLOW_ANY_GO=1 to run with another toolchain.
# KIND=emit pins codegen rather than behaviour, so it is the kind most tightly
# coupled to the compiler release: a GALA bump is expected to fail those probes,
# and the re-baseline procedure is in docs/gala-translation.md under
# "When to revisit".

set -uo pipefail

cd "$(dirname "$0")"

EXPECTED_GALA_VERSION="0.84.1"
actual_version="$(gala version 2>/dev/null | awk '{print $NF}')"
if [[ "$actual_version" != "$EXPECTED_GALA_VERSION" && "${GALA_PROBES_ALLOW_ANY_VERSION:-0}" != "1" ]]; then
    printf 'expected gala %s, found %s\n' "$EXPECTED_GALA_VERSION" "${actual_version:-none}" >&2
    printf 'the checked-in expectations pin that compiler; set GALA_PROBES_ALLOW_ANY_VERSION=1 to override\n' >&2
    exit 2
fi

EXPECTED_GO_SERIES="go1.26"
actual_go_version="$(go version 2>/dev/null | awk '{print $3}')"
actual_go_series="${actual_go_version%.*}"
if [[ "$actual_go_series" != "$EXPECTED_GO_SERIES" && "${GALA_PROBES_ALLOW_ANY_GO:-0}" != "1" ]]; then
    printf 'expected %s.x, found %s\n' "$EXPECTED_GO_SERIES" "${actual_go_version:-none}" >&2
    printf 'the go build expectations embed compiler diagnostics; set GALA_PROBES_ALLOW_ANY_GO=1 to override\n' >&2
    exit 2
fi

LOG_DIR=".logs"
rm -rf "$LOG_DIR"
mkdir -p "$LOG_DIR"

passed=0
failed=0
failed_names=()

# The bare-name probes share one colliding fixture: an imported package that also
# exports Response, and the handwritten constructor that returns the local type.
# Writing it once here is what keeps a probe from drifting away from the others.

write_collide_package() {
    mkdir -p fixtures/collide
    cat >fixtures/collide/response.go <<'GO'
package collide

type Response struct {
	Status int
}

func Default() Response {
	return Response{Status: 500}
}
GO
}

write_response_helper() {
    cat >"probes/$1/helper.go" <<'GO'
package main

type Response struct {
	Status int
}

func newResponse(status int) *Response {
	return &Response{Status: status}
}
GO
}

materialize_fixtures() {
    case "$1" in
    pass_name_collision_workaround | contested_bare_response_name | blocked_bare_name_type_position)
        write_collide_package
        write_response_helper "$1"
        ;;
    blocked_bare_name_type_gala_sibling)
        write_collide_package
        ;;
    pass_exported_val)
        cat >"probes/$1/check.go" <<'GO'
package main

import "fmt"

func init() {
	fmt.Println(ExportedAnswer.Get())
}
GO
        ;;
    pass_struct_methods)
        cat >"probes/$1/check.go" <<'GO'
package main

import "fmt"

func init() {
	c := Counter{N: 1}
	fmt.Println(c.Next())
}
GO
        ;;
    blocked_bare_name_type_position)
        cat >"probes/$1/helper.go" <<'GO'
package main

type Response struct {
	Status int
}

func newResponse(status int) *Response {
	return &Response{Status: status}
}
GO
        ;;
    blocked_receiver_unwrap_samepkg | blocked_receiver_unwrap_val | emit_var_receiver_no_wrapper)
        cat >"probes/$1/types.go" <<'GO'
package main

type Greeter struct{ Name string }

func (g Greeter) Hello() string { return "hello " + g.Name }

func (g *Greeter) Rename(name string) { g.Name = name }

func NewGreeter(name string) Greeter { return Greeter{Name: name} }
GO
        ;;
    blocked_e0044_go_sibling_method)
        cat >"probes/$1/types.go" <<'GO'
package main

func (r Repo) Save() error { return nil }
GO
        ;;
    blocked_resource_go_sibling_type)
        cat >"probes/$1/types.go" <<'GO'
package main

type Res struct{ name string }

func (r Res) Close() error { return nil }

func (r Res) Name() string { return r.name }

func OpenRes(name string) Res { return Res{name: name} }
GO
        ;;
    pass_slice_from_size_len)
        cat >"probes/$1/types.go" <<'GO'
package main

func listLogs() ([]string, error) {
	return []string{"a", "b", "c"}, nil
}
GO
        ;;
    blocked_size_inferred_receiver)
        cat >"probes/$1/types.go" <<'GO'
package main

type Log struct {
	ID      string
	Members []Member
}

type Member struct {
	Email string
}

func listLogs() ([]Log, error) {
	return nil, nil
}
GO
        ;;
    contested_size_field_go_sibling)
        cat >"probes/$1/types.go" <<'GO'
package main

type Command struct {
	Usage string
}
GO
        ;;
    esac
}

expect_value() {
    sed -n "s/^$2=//p" "$1" | head -n 1
}

# expect_values returns every value of a key, one per line, in file order. The
# single-valued keys above still take the first value; only the repeatable
# KIND=emit marker keys read more than one.
expect_values() {
    sed -n "s/^$2=//p" "$1"
}

report_ok() {
    printf 'PASS %s: %s\n' "$1" "$2"
    passed=$((passed + 1))
}

report_fail() {
    printf 'FAIL %s: %s\n' "$1" "$2"
    failed=$((failed + 1))
    failed_names+=("$1")
}

printf 'go %s; gala %s; %s probes\n\n' "${actual_go_version:-unknown}" "${actual_version:-unknown}" "$(find probes -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')"

for dir in probes/*/; do
    name="$(basename "$dir")"
    expect_file="$dir/expect"
    kind="$(expect_value "$expect_file" KIND)"
    code="$(expect_value "$expect_file" CODE)"
    err="$(expect_value "$expect_file" ERR)"
    want_run="$(expect_value "$expect_file" RUN)"
    note="$(expect_value "$expect_file" NOTE)"
    mapfile -t contains_markers < <(expect_values "$expect_file" CONTAINS)
    mapfile -t absent_markers < <(expect_values "$expect_file" ABSENT)
    transpile_log="$LOG_DIR/$name.transpile.log"
    build_log="$LOG_DIR/$name.build.log"
    run_log="$LOG_DIR/$name.run.log"
    emit_log="$LOG_DIR/$name.emit.txt"

    materialize_fixtures "$name"

    (cd "$dir" && gala transpile -i main.gala -o main.gen.go) >"$transpile_log" 2>&1
    transpile_status=$?

    if [[ -f "$dir/sibling.gala" && $transpile_status -eq 0 ]]; then
        (cd "$dir" && gala transpile -i sibling.gala -o sibling.gen.go) >>"$transpile_log" 2>&1
        transpile_status=$?
    fi

    if [[ -n "$note" ]]; then
        printf '  note: %s\n' "$note"
    fi

    case "$kind" in
    transpile_fail)
        if [[ $transpile_status -eq 0 ]]; then
            report_fail "$name" "transpile unexpectedly succeeded"
            continue
        fi
        if [[ -n "$code" ]] && ! grep -qF "$code" "$transpile_log"; then
            report_fail "$name" "diagnostic is missing $code"
            sed -n '1,6p' "$transpile_log"
            continue
        fi
        if [[ -n "$err" ]] && ! grep -qF "$err" "$transpile_log"; then
            report_fail "$name" "diagnostic is missing: $err"
            sed -n '1,6p' "$transpile_log"
            continue
        fi
        report_ok "$name" "transpile rejected as expected"
        printf '  %s\n' "$(head -n 1 "$transpile_log")"
        ;;
    build_fail)
        if [[ $transpile_status -ne 0 ]]; then
            report_fail "$name" "transpile unexpectedly failed"
            sed -n '1,6p' "$transpile_log"
            continue
        fi
        go build -o "$LOG_DIR/$name.bin" "./probes/$name" >"$build_log" 2>&1
        build_status=$?
        if [[ $build_status -eq 0 ]]; then
            report_fail "$name" "go build unexpectedly succeeded"
            continue
        fi
        if [[ -n "$err" ]] && ! grep -qF "$err" "$build_log"; then
            report_fail "$name" "build error is missing: $err"
            sed -n '1,6p' "$build_log"
            continue
        fi
        report_ok "$name" "transpiled, then go build failed as expected"
        printf '  %s\n' "$(grep -m1 -F "$err" "$build_log")"
        ;;
    pass)
        if [[ $transpile_status -ne 0 ]]; then
            report_fail "$name" "transpile failed"
            sed -n '1,6p' "$transpile_log"
            continue
        fi
        go build -o "$LOG_DIR/$name.bin" "./probes/$name" >"$build_log" 2>&1
        build_status=$?
        if [[ $build_status -ne 0 ]]; then
            report_fail "$name" "go build failed"
            sed -n '1,6p' "$build_log"
            continue
        fi
        if [[ "$want_run" == "yes" ]]; then
            "./$LOG_DIR/$name.bin" >"$run_log" 2>&1
            run_status=$?
            if [[ $run_status -ne 0 ]]; then
                report_fail "$name" "binary exited $run_status"
                sed -n '1,6p' "$run_log"
                continue
            fi
            if [[ ! -f "$dir/expected.out" ]]; then
                report_fail "$name" "RUN=yes but expected.out is missing"
                continue
            fi
            if ! diff -u "$dir/expected.out" "$run_log" >"$LOG_DIR/$name.diff" 2>&1; then
                report_fail "$name" "output differs from expected.out"
                cat "$LOG_DIR/$name.diff"
                continue
            fi
            report_ok "$name" "transpiled, built, and output matched"
        else
            report_ok "$name" "transpiled and built"
        fi
        ;;
    emit)
        if [[ $transpile_status -ne 0 ]]; then
            report_fail "$name" "transpile failed"
            sed -n '1,6p' "$transpile_log"
            continue
        fi
        if [[ ${#contains_markers[@]} -eq 0 && ${#absent_markers[@]} -eq 0 ]]; then
            report_fail "$name" "KIND=emit without a CONTAINS or ABSENT marker"
            continue
        fi
        if [[ ! -f "$dir/main.gen.go" ]]; then
            report_fail "$name" "transpile succeeded but wrote no main.gen.go"
            continue
        fi
        { printf -- '--- %s\n' "$dir/main.gen.go"; cat -n "$dir/main.gen.go"; } >"$emit_log" 2>&1
        # Every failing assertion is collected, not just the first, so one run
        # shows the whole drift; report_fail prints the FAIL line, then the
        # collected detail follows the way the other kinds' excerpts do.
        emit_detail=()
        emit_failures=0
        for marker in "${contains_markers[@]}"; do
            if [[ -z "$marker" ]]; then
                emit_detail+=("  empty CONTAINS marker")
                emit_failures=$((emit_failures + 1))
            elif ! grep -qF -- "$marker" "$dir/main.gen.go"; then
                emit_detail+=("  no match for CONTAINS=$marker")
                emit_failures=$((emit_failures + 1))
            fi
        done
        for marker in "${absent_markers[@]}"; do
            if [[ -z "$marker" ]]; then
                emit_detail+=("  empty ABSENT marker")
                emit_failures=$((emit_failures + 1))
            elif grep -qF -- "$marker" "$dir/main.gen.go"; then
                emit_detail+=("  matched ABSENT=$marker, which the emitted Go must not contain")
                while IFS= read -r hit; do
                    emit_detail+=("  $hit")
                done < <(grep -nF -C 2 -- "$marker" "$dir/main.gen.go")
                emit_failures=$((emit_failures + 1))
            fi
        done
        if [[ $emit_failures -gt 0 ]]; then
            report_fail "$name" "$emit_failures of $(( ${#contains_markers[@]} + ${#absent_markers[@]} )) emitted-text assertion(s) did not hold"
            printf '%s\n' "${emit_detail[@]}"
            printf '  the generated Go, numbered; a full copy is at %s\n' "$emit_log"
            sed -n '2,61p' "$emit_log" | sed 's/^/  /'
            continue
        fi
        report_ok "$name" "transpiled, and the emitted Go held ${#contains_markers[@]} presence and ${#absent_markers[@]} absence assertion(s)"
        ;;
    *)
        report_fail "$name" "unknown KIND: $kind"
        ;;
    esac
done

printf '\n%d passed, %d failed\n' "$passed" "$failed"
if [[ $failed -gt 0 ]]; then
    printf 'failed: %s\n' "${failed_names[*]}"
    exit 1
fi
