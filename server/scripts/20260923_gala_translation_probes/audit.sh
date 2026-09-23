#!/usr/bin/env bash
# Read-only consistency audit for the Go-to-GALA translation roster in
# docs/gala-translation.md.
#
# It needs no GALA or Go toolchain and checks:
#   1. probe directory integrity: main.gala, expect, a valid KIND, and
#      expected.out whenever RUN=yes
#   2. no orphan probes and no roster link that does not resolve to a directory
#   3. every genuine-gap row names a failing, non-contested probe
#   4. every language/boundary roster row has a gap row and vice versa
#   5. every GAP-n referenced from server/GALA.md exists here and vice versa
#
# The corpus runner (run.sh) pins the toolchains; this script only checks the
# artifacts against each other.

set -uo pipefail

cd "$(dirname "$0")"

DOC="../../../docs/gala-translation.md"
SERVER_GALA="../../GALA.md"

failures=0

fail() {
    printf 'FAIL %s\n' "$1"
    failures=$((failures + 1))
}

ok() { printf 'ok   %s\n' "$1"; }

expect_value() {
    sed -n "s/^$2=//p" "$1" | head -n 1
}

if [[ ! -f "$DOC" ]]; then
    printf 'missing %s\n' "$DOC" >&2
    exit 2
fi
if [[ ! -f "$SERVER_GALA" ]]; then
    printf 'missing %s\n' "$SERVER_GALA" >&2
    exit 2
fi

# --- 1. probe directory integrity -----------------------------------------

probe_count=0
for dir in probes/*/; do
    name="$(basename "$dir")"
    probe_count=$((probe_count + 1))
    if [[ ! -f "$dir/main.gala" ]]; then
        fail "$name: missing main.gala"
    fi
    if [[ ! -f "$dir/expect" ]]; then
        fail "$name: missing expect"
        continue
    fi
    kind="$(expect_value "$dir/expect" KIND)"
    case "$kind" in
    pass)
        if [[ "$(expect_value "$dir/expect" RUN)" == "yes" && ! -f "$dir/expected.out" ]]; then
            fail "$name: RUN=yes but expected.out is missing"
        fi
        ;;
    transpile_fail | build_fail)
        if [[ -z "$(expect_value "$dir/expect" CODE)" && -z "$(expect_value "$dir/expect" ERR)" ]]; then
            fail "$name: $kind without CODE or ERR"
        fi
        ;;
    *)
        fail "$name: unknown KIND '$kind'"
        ;;
    esac
done
if [[ ! -d probes || "$probe_count" -eq 0 ]]; then
    fail "no probe directories under probes/"
else
    ok "probe directories ($probe_count)"
fi

# --- 2. orphan probes and unresolved links ---------------------------------

before=$failures
for dir in probes/*/; do
    name="$(basename "$dir")"
    if ! grep -qF "$name" "$DOC"; then
        fail "orphan probe: $name is not referenced in $DOC"
    fi
done
while IFS= read -r linked; do
    if [[ ! -d "probes/$linked" ]]; then
        fail "unresolved roster link: probes/$linked/ does not exist"
    fi
done < <(grep -o '/probes/[a-z0-9_]*/' "$DOC" | sort -u | sed 's#^/probes/##; s#/$##')
if [[ $failures -eq $before ]]; then
    ok "no orphan probes, all roster links resolve"
fi

# --- 3. gap rows name failing, non-contested probes ------------------------

gap_rows="$(mktemp)"
trap 'rm -f "$gap_rows"' EXIT
awk -F'|' '
function trim(s) { gsub(/^[ \t]+/, "", s); gsub(/[ \t]+$/, "", s); return s }
/^\| GAP-/ {
    id = trim($2); kind = trim($3)
    fams = trim($4); gsub(/`/, "", fams); gsub(/ *, */, ",", fams)
    probes = ""
    contested = 0
    n = split($0, parts, "`")
    for (i = 2; i <= n; i += 2) {
        t = parts[i]
        if (t ~ /^blocked_/) probes = probes (probes == "" ? "" : ",") t
        if (t ~ /^contested_/) contested = 1
    }
    print id "\x1f" kind "\x1f" fams "\x1f" probes "\x1f" contested
}' "$DOC" >"$gap_rows"

gap_count=0
before_gaps=$failures
declare -A gap_kind=()
declare -A gap_family=()
declare -A gap_probe_ok=()
while IFS=$'\x1f' read -r id kind fams probes contested; do
    [[ -z "$id" ]] && continue
    gap_count=$((gap_count + 1))
    case "$kind" in
    language | boundary) ;;
    *) fail "$id: kind '$kind' is not language or boundary" ;;
    esac
    if [[ "$contested" == "1" ]]; then
        fail "$id: a contested probe may not back a gap claim"
    fi
    if [[ -z "$probes" ]]; then
        fail "$id: no blocking probe named"
    fi
    IFS=',' read -r -a probe_list <<<"$probes"
    for probe in "${probe_list[@]}"; do
        if [[ ! -d "probes/$probe" ]]; then
            fail "$id: probe $probe does not exist"
            continue
        fi
        kind_value="$(expect_value "probes/$probe/expect" KIND)"
        if [[ "$kind_value" == "pass" ]]; then
            fail "$id: probe $probe is KIND=pass, not a blocker"
        fi
        gap_probe_ok["$probe"]=1
    done
    IFS=',' read -r -a fam_list <<<"$fams"
    for fam in "${fam_list[@]}"; do
        [[ -n "$fam" ]] && gap_family["$fam"]=1
    done
    gap_kind["$id"]=$kind
done <"$gap_rows"
if [[ "$gap_count" -eq 0 ]]; then
    fail "no GAP-n rows found in $DOC"
fi
for i in $(seq 1 "$gap_count"); do
    if [[ -z "${gap_kind[GAP-$i]:-}" ]]; then
        fail "GAP-$i is missing; gap IDs must be contiguous from GAP-1"
    fi
done
if [[ $failures -eq $before_gaps ]]; then
    ok "gap rows name failing, non-contested probes"
fi

# --- 4. class-to-gap consistency -------------------------------------------

before_classes=$failures
declare -A family_class=()
while IFS=$'\t' read -r fam cls; do
    [[ -n "$fam" ]] && family_class["$fam"]="$cls"
done < <(awk -F'|' '
function trim(s) { gsub(/^[ \t]+/, "", s); gsub(/[ \t]+$/, "", s); return s }
/^\| Construct/ && /Gap class/ {
    for (i = 2; i < NF; i++) {
        h = trim($i)
        if (h == "Gap class") gc = i
        if (h == "Construct") fc = i
    }
    in_table = 1
    next
}
in_table && /^\|/ {
    if ($0 ~ /^\| *-+/) next
    fam = trim($fc); gsub(/`/, "", fam)
    cls = trim($gc)
    if (fam != "") print fam "\t" cls
    next
}
{ in_table = 0 }' "$DOC")

if [[ "${#family_class[@]}" -eq 0 ]]; then
    fail "no Gap class column found in $DOC"
fi
for fam in "${!family_class[@]}"; do
    cls="${family_class[$fam]}"
    case "$cls" in
    language | boundary)
        if [[ -z "${gap_family[$fam]:-}" ]]; then
            fail "$fam is classified $cls but has no GAP-n row"
        fi
        ;;
    "—" | analog | workaround) ;;
    *) fail "$fam: unknown Gap class '$cls'" ;;
    esac
done
for fam in "${!gap_family[@]}"; do
    cls="${family_class[$fam]:-}"
    if [[ -z "$cls" ]]; then
        fail "$fam appears in a GAP-n row but not in the roster"
    elif [[ "$cls" != "language" && "$cls" != "boundary" ]]; then
        fail "$fam appears in a GAP-n row but is classified '$cls'"
    fi
done
if [[ "${#family_class[@]}" -gt 0 && "$failures" -eq "$before_classes" ]]; then
    ok "gap rows, blocking probes, and Gap classes agree"
fi

# --- 5. GAP-n cross-references with server/GALA.md --------------------------

before=$failures
for i in $(seq 1 "$gap_count"); do
    if ! grep -qF "GAP-$i" "$SERVER_GALA"; then
        fail "GAP-$i is not referenced from $SERVER_GALA"
    fi
done
while IFS= read -r referenced; do
    if ! grep -qF "$referenced" "$DOC"; then
        fail "$SERVER_GALA references $referenced, which is not in $DOC"
    fi
done < <(grep -o 'GAP-[0-9]*' "$SERVER_GALA" | sort -u)
if [[ $failures -eq $before ]]; then
    ok "$gap_count GAP IDs cross-referenced with server/GALA.md"
fi

# --- report -----------------------------------------------------------------

if [[ $failures -gt 0 ]]; then
    printf '\naudit failed: %d problem(s)\n' "$failures"
    exit 1
fi
printf '\naudit passed\n'
