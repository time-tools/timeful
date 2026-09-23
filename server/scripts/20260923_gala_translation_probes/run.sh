#!/usr/bin/env bash
# Scripted probe corpus for the Go-to-GALA translation roster in docs/gala-translation.md.
#
# Each probes/<name>/ directory holds a main.gala and an expect file:
#   KIND=pass            transpile and go build must succeed; RUN=yes also runs the
#                        binary and diffs stdout against expected.out
#   KIND=transpile_fail  gala transpile must fail; CODE and ERR are checked against
#                        the diagnostic
#   KIND=build_fail      transpile must succeed and go build must fail on ERR
#
# Generated Go files (main.gen.go), runner-written fixtures, logs, and binaries are
# gitignored by explicit path in .gitignore; only the .gala sources, expect files,
# expected.out files, go.mod, .gitignore, and this runner are committed.
#
# The expectations pin GALA 0.81.0. Set GALA_PROBES_ALLOW_ANY_VERSION=1 to run against
# another compiler and see which expectations moved.
# The go build expectations also pin the Go 1.26 toolchain series; set
# GALA_PROBES_ALLOW_ANY_GO=1 to run with another toolchain.

set -uo pipefail

cd "$(dirname "$0")"

EXPECTED_GALA_VERSION="0.81.0"
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

materialize_fixtures() {
    case "$1" in
    pass_name_collision_workaround | contested_bare_response_name)
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
    transpile_log="$LOG_DIR/$name.transpile.log"
    build_log="$LOG_DIR/$name.build.log"
    run_log="$LOG_DIR/$name.run.log"

    materialize_fixtures "$name"

    (cd "$dir" && gala transpile -i main.gala -o main.gen.go) >"$transpile_log" 2>&1
    transpile_status=$?

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
