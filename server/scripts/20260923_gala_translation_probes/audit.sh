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
#   6. the mechanical rewrite catalog and the no-rewrite index agree with the
#      roster's gap classes, in both directions
#
# Checks 4 and 6 read the construct roster section of the page, so a table
# elsewhere on the page that repeats the same column names cannot influence
# either of them.
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

# section_of prints the body of the level-two Markdown section titled $2 in $1,
# up to the next level-two heading, skipping fenced code blocks. A table in
# another section is therefore not part of what a caller reads.
section_of() {
    awk -v want="$2" '
        /^```/ { fenced = !fenced }
        !fenced && /^## [^#]/ {
            name = $0
            sub(/^## /, "", name)
            sub(/[ \t]+#*$/, "", name)
            if (inside) exit
            inside = (name == want)
            next
        }
        inside { print }
    ' "$1"
}

gap_rows=""
roster_slice=""
catalog_rows=""
trap 'rm -f "$gap_rows" "$roster_slice" "$catalog_rows"' EXIT

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

# The construct roster is the authoritative slice of the page: the replacement
# summary and the inventory table repeat its column names, so only this section
# may decide what a family is classified as.
roster_slice="$(mktemp)"
section_of "$DOC" "Construct roster" >"$roster_slice"

before_classes=$failures
declare -A family_class=()
declare -A roster_probe=()
while IFS=$'\t' read -r fam cls probes; do
    [[ -n "$fam" ]] || continue
    family_class["$fam"]="$cls"
    roster_probe["$fam"]="$probes"
done < <(awk -F'|' '
function trim(s) { gsub(/^[ \t]+/, "", s); gsub(/[ \t]+$/, "", s); return s }
function links(cell,   out, parts, n, i) {
    n = split(cell, parts, "/probes/")
    for (i = 2; i <= n; i++) {
        sub(/\/.*$/, "", parts[i])
        out = out (out == "" ? "" : ",") parts[i]
    }
    return out
}
/^\| Construct/ && /Gap class/ {
    for (i = 2; i < NF; i++) {
        h = trim($i)
        if (h == "Gap class") gc = i
        if (h == "Construct") fc = i
        if (h == "Probe") pc = i
    }
    in_table = 1
    next
}
in_table && /^\|/ {
    if ($0 ~ /^\| *-+/) next
    fam = trim($fc); gsub(/`/, "", fam)
    if (fam != "") print fam "\t" trim($(gc)) "\t" links($(pc))
    next
}
{ in_table = 0 }' "$roster_slice")

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

# --- 6. mechanical rewrite catalog ------------------------------------------

# The catalog is derived from the roster's gap classes, so it is gated against
# the same slice check 4 read: a row may not name a family the roster does not
# classify as substituted, a substitutable family may not be left without a
# row, every named probe must exist and must be one the roster records for a
# family the row covers, and the no-rewrite index must list exactly the
# families that have no substitute.

before_catalog=$failures
catalog_rows="$(mktemp)"
section_of "$DOC" "Mechanical rewrites" | awk -F'|' '
function trim(s) { gsub(/^[ \t]+/, "", s); gsub(/[ \t]+$/, "", s); return s }
function names(cell,   out, parts, n, i, t) {
    n = split(cell, parts, "`")
    for (i = 2; i <= n; i += 2) {
        t = parts[i]
        gsub(/^[ \t]+/, "", t); gsub(/[ \t]+$/, "", t)
        if (t != "") out = out (out == "" ? "" : ",") t
    }
    return out
}
function links(cell,   out, parts, n, i) {
    n = split(cell, parts, "/probes/")
    for (i = 2; i <= n; i++) {
        sub(/\/.*$/, "", parts[i])
        out = out (out == "" ? "" : ",") parts[i]
    }
    return out
}
/^\| Go spelling/ {
    for (i = 2; i < NF; i++) {
        h = trim($i)
        if (h == "Construct families") fc = i
        if (h == "Probe") pc = i
    }
    kind = "rule"
    in_table = 1
    next
}
/^\| Gap class/ && /no mechanical rewrite/ {
    for (i = 2; i < NF; i++) {
        h = trim($i)
        if (h == "Gap class") gc = i
        if (h == "Construct families with no mechanical rewrite") fc = i
    }
    kind = "index"
    in_table = 1
    next
}
in_table && /^\|/ {
    if ($0 ~ /^\| *-+/) next
    if (kind == "rule") {
        print "rule\x1f" names($(fc)) "\x1f" links($(pc))
    } else {
        cls = trim($(gc)); gsub(/`/, "", cls)
        print "index\x1f" names($(fc)) "\x1f" cls
    }
    next
}
{ in_table = 0 }' >"$catalog_rows"

declare -A catalog_family=()
declare -A indexed_family=()
rule_rows=0
index_rows=0
# The second field is always the construct families; the third is the probe
# list for a rule row and the gap class for an index row.
while IFS=$'\x1f' read -r kind named extra; do
    case "$kind" in
    rule)
        rule_rows=$((rule_rows + 1))
        IFS=',' read -r -a fams <<<"$named"
        IFS=',' read -r -a probes <<<"$extra"
        for fam in "${fams[@]}"; do
            [[ -n "$fam" ]] || continue
            catalog_family["$fam"]=1
            case "${family_class[$fam]:-}" in
            analog | workaround) ;;
            "") fail "catalog row $rule_rows: $fam is not a construct family in the roster" ;;
            *) fail "catalog row $rule_rows: $fam is classified '${family_class[$fam]}', which has no substitute" ;;
            esac
        done
        for probe in "${probes[@]}"; do
            [[ -n "$probe" ]] || continue
            if [[ ! -d "probes/$probe" ]]; then
                fail "catalog row $rule_rows: probe $probe does not exist"
                continue
            fi
            pinned=0
            for fam in "${fams[@]}"; do
                [[ -n "$fam" ]] || continue
                if [[ ",${roster_probe[$fam]:-}," == *",$probe,"* ]]; then
                    pinned=1
                fi
            done
            if [[ $pinned -eq 0 ]]; then
                fail "catalog row $rule_rows: probe $probe is not recorded in the roster for ${fams[*]}"
            fi
        done
        ;;
    index)
        index_rows=$((index_rows + 1))
        IFS=',' read -r -a fams <<<"$named"
        for fam in "${fams[@]}"; do
            [[ -n "$fam" ]] || continue
            indexed_family["$fam"]="$extra"
            case "${family_class[$fam]:-}" in
            "$extra") ;;
            "") fail "no-rewrite index row $index_rows: $fam is not a construct family in the roster" ;;
            *) fail "no-rewrite index row $index_rows: $fam is classified '${family_class[$fam]}', not '$extra'" ;;
            esac
            if [[ -n "${catalog_family[$fam]:-}" ]]; then
                fail "no-rewrite index row $index_rows: $fam also has a mechanical rewrite row"
            fi
        done
        ;;
    esac
done <"$catalog_rows"

if [[ $rule_rows -eq 0 ]]; then
    fail "no mechanical rewrite rows found in the Mechanical rewrites section of $DOC"
fi
if [[ $index_rows -eq 0 ]]; then
    fail "no no-rewrite index rows found in the Mechanical rewrites section of $DOC"
fi
for fam in "${!family_class[@]}"; do
    case "${family_class[$fam]}" in
    analog | workaround)
        if [[ -z "${catalog_family[$fam]:-}" ]]; then
            fail "$fam has a substitute but no mechanical rewrite row"
        fi
        ;;
    *)
        if [[ -z "${indexed_family[$fam]:-}" ]]; then
            fail "$fam has no mechanical rewrite and is not listed in the no-rewrite index"
        fi
        ;;
    esac
done
if [[ $failures -eq $before_catalog ]]; then
    ok "$rule_rows mechanical rewrite rows and $index_rows no-rewrite index rows agree with the roster"
fi

# --- report -----------------------------------------------------------------

if [[ $failures -gt 0 ]]; then
    printf '\naudit failed: %d problem(s)\n' "$failures"
    exit 1
fi
printf '\naudit passed\n'
