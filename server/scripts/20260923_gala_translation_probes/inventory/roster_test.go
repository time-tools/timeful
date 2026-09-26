package main

import (
	"strings"
	"testing"
)

// fixturePage is a miniature roster page. It carries the two sections the
// preflight mode reads, and three decoys that it must ignore: a replacement
// table outside the construct roster that repeats the Construct, Gap class, and
// Verdict headers, a gap table that names a Verdict column, and a fenced block
// that looks like a second construct roster.
var fixturePage = strings.Join([]string{
	"# Roster fixture",
	"",
	"## Decision ladder",
	"",
	"Choose the first rung that fits.",
	"",
	"1. **Rewrite runtime-free.**",
	"2. **Rewrite runtime-enabled, or split with a handwritten sibling.**",
	"3. **Keep handwritten.**",
	"",
	"## Construct roster",
	"",
	"| Construct | Gap class | Minimal Go | Verdict |",
	"| --- | --- | --- | --- |",
	"| `len-calls` | analog | `len(s)` | Workaround |",
	"| `defer` | analog | `defer f.Close()` | Workaround (`use`); keep handwritten when cleanup is not a resource release |",
	"| `struct-tags` | language | a tagged struct field | Keep handwritten |",
	"",
	"### Statements and control flow",
	"",
	"| Construct | Gap class | Minimal Go | Verdict |",
	"| --- | --- | --- | --- |",
	"| `select-statements` | language | `select { ... }` | Keep handwritten |",
	"| `multi-value-define` | workaround | `n, err := f()` | Workaround (`var`) until #529 ships |",
	"",
	"## Replaced by analog or workaround",
	"",
	"| Construct | Gap class | Replacement | Verdict |",
	"| --- | --- | --- | --- |",
	"| `len-calls` | analog | `.ByteSize()` | decoy replacement verdict |",
	"| `cap-calls` | — | nothing | decoy verdict for a family the roster does not classify |",
	"",
	"## Genuine gaps",
	"",
	"| Gap | Construct families | Verdict |",
	"| --- | --- | --- |",
	"| GAP-1 | `struct-tags` | decoy gap verdict |",
	"",
	"## When to revisit",
	"",
	"```",
	"## Construct roster",
	"",
	"| Construct | Gap class | Verdict |",
	"| --- | --- | --- |",
	"| `code-fence-decoy` | language | decoy verdict inside a fence |",
	"```",
	"",
}, "\n")

// TestParseRosterBoundedToConstructSection pins the section bounding. The
// replacement table on the fixture page carries the same Construct, Gap class,
// and Verdict headers as the roster, so only the section bound keeps its rows
// and its values out of the parse.
func TestParseRosterBoundedToConstructSection(t *testing.T) {
	r, err := parseRoster(fixturePage)
	if err != nil {
		t.Fatalf("parseRoster: %v", err)
	}
	want := map[string]rosterRow{
		"len-calls":          {family: "len-calls", gapClass: "analog", verdict: "Workaround"},
		"defer":              {family: "defer", gapClass: "analog", verdict: "Workaround (`use`); keep handwritten when cleanup is not a resource release"},
		"struct-tags":        {family: "struct-tags", gapClass: "language", verdict: "Keep handwritten"},
		"select-statements":  {family: "select-statements", gapClass: "language", verdict: "Keep handwritten"},
		"multi-value-define": {family: "multi-value-define", gapClass: "workaround", verdict: "Workaround (`var`) until #529 ships"},
	}
	if len(r.rows) != len(want) {
		t.Errorf("parsed %d families, want %d: %v", len(r.rows), len(want), keys(r.rows))
	}
	for family, wantRow := range want {
		got, ok := r.rows[family]
		if !ok {
			t.Errorf("%s: missing from the parse", family)
			continue
		}
		if got != wantRow {
			t.Errorf("%s: parsed %+v, want %+v", family, got, wantRow)
		}
	}
	for _, decoy := range []string{"cap-calls", "GAP-1", "code-fence-decoy"} {
		if _, ok := r.rows[decoy]; ok {
			t.Errorf("%s: a table outside the construct roster section leaked into the parse", decoy)
		}
	}
}

// TestParseRosterColumnsByHeaderName pins the header-driven column mapping: the
// same rows in a different column order must produce the same parse.
func TestParseRosterColumnsByHeaderName(t *testing.T) {
	page := strings.Join([]string{
		"## Decision ladder",
		"",
		"1. **Rewrite runtime-free.**",
		"2. **Rewrite runtime-enabled, or split with a handwritten sibling.**",
		"3. **Keep handwritten.**",
		"",
		"## Construct roster",
		"",
		"| Verdict | Reference | Gap class | Construct |",
		"| --- | --- | --- | --- |",
		"| Keep handwritten | a reference | language | `struct-tags` |",
		"| Workaround | another reference | analog | `len-calls` |",
		"",
	}, "\n")
	r, err := parseRoster(page)
	if err != nil {
		t.Fatalf("parseRoster: %v", err)
	}
	want := rosterRow{family: "struct-tags", gapClass: "language", verdict: "Keep handwritten"}
	if got := r.rows["struct-tags"]; got != want {
		t.Errorf("struct-tags: parsed %+v, want %+v", got, want)
	}
	if got := r.rows["len-calls"]; got.gapClass != "analog" || got.verdict != "Workaround" {
		t.Errorf("len-calls: parsed %+v, want gap class analog and verdict Workaround", got)
	}
}

func TestParseRosterRejectsRestructuredPages(t *testing.T) {
	cases := []struct {
		name string
		page string
		want string
	}{
		{
			name: "no construct roster section",
			page: "# Roster\n\n## Decision ladder\n\n1. **Rewrite runtime-free.**\n2. **Split with a handwritten sibling.**\n3. **Keep handwritten.**\n",
			want: `no "Construct roster" section`,
		},
		{
			name: "no verdict column",
			page: "## Decision ladder\n\n1. **Rewrite runtime-free.**\n2. **Split with a handwritten sibling.**\n3. **Keep handwritten.**\n\n## Construct roster\n\n| Construct | Gap class | Outcome |\n| --- | --- | --- |\n| `len-calls` | analog | Workaround |\n",
			want: "no table with the Construct, Gap class, Verdict columns",
		},
		{
			name: "empty verdict cell",
			page: "## Decision ladder\n\n1. **Rewrite runtime-free.**\n2. **Split with a handwritten sibling.**\n3. **Keep handwritten.**\n\n## Construct roster\n\n| Construct | Gap class | Verdict |\n| --- | --- | --- |\n| `len-calls` | analog | |\n",
			want: "empty Gap class or Verdict cell",
		},
		{
			name: "no decision ladder section",
			page: "## Construct roster\n\n| Construct | Gap class | Verdict |\n| --- | --- | --- |\n| `len-calls` | analog | Workaround |\n",
			want: `no "Decision ladder" section`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseRoster(c.page)
			if err == nil {
				t.Fatalf("parseRoster succeeded, want an error containing %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q, want it to contain %q", err, c.want)
			}
		})
	}
}

func TestReadLadder(t *testing.T) {
	r, err := parseRoster(fixturePage)
	if err != nil {
		t.Fatalf("parseRoster: %v", err)
	}
	lad := r.ladder
	if len(lad.labels) != 3 {
		t.Fatalf("read %d rungs, want 3: %v", len(lad.labels), lad.labels)
	}
	if lad.rewrite != 1 || lad.split != 2 || lad.handwritten != 3 {
		t.Errorf("rungs rewrite=%d split=%d handwritten=%d, want 1, 2, 3", lad.rewrite, lad.split, lad.handwritten)
	}
	if got, want := lad.label(2), "Rewrite runtime-enabled, or split with a handwritten sibling."; got != want {
		t.Errorf("label(2) = %q, want the roster's own %q", got, want)
	}
	if got := lad.label(4); got != "" {
		t.Errorf("label(4) = %q, want an empty string", got)
	}
}

func TestReadLadderRejectsUnnamedRungs(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "no split rung",
			body: "1. **Rewrite runtime-free.**\n2. **Rewrite runtime-enabled.**\n3. **Keep handwritten.**\n",
			want: "does not name a rewrite rung, a split rung, and a keep-handwritten rung",
		},
		{
			name: "too few rungs",
			body: "1. **Rewrite runtime-free.**\n2. **Split with a handwritten sibling.**\n",
			want: "has 2 numbered rungs, want at least 3",
		},
		{
			name: "no rungs at all",
			body: "Choose the first rung that fits.\n",
			want: "has 0 numbered rungs, want at least 3",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := readLadder(strings.Split(c.body, "\n"))
			if err == nil {
				t.Fatalf("readLadder succeeded, want an error containing %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q, want it to contain %q", err, c.want)
			}
		})
	}
}

// TestClassifyVerdict covers the verdict vocabulary the roster uses, including
// the negated handwritten clause that is not a handwritten outcome. An empty
// cell is not in the table because readRosterTables rejects one.
func TestClassifyVerdict(t *testing.T) {
	cases := []struct {
		verdict string
		want    decision
	}{
		{"Direct form", decisionRewrite},
		{"Direct form with `var`", decisionRewrite},
		{"Direct form in signatures; workaround for literal construction", decisionRewrite},
		{"Workaround", decisionRewrite},
		{"Workaround (`var`) until #529 ships", decisionRewrite},
		{"Analog (type-pattern `match`, runtime-enabled)", decisionRewrite},
		{"Workaround (`go_interop` helpers); no handwritten sibling needed", decisionRewrite},
		{"Split with a handwritten sibling", decisionSplit},
		{"Workaround (`use` for single-value acquires); keep handwritten when cleanup is not a resource release", decisionSplit},
		{"Direct form for `any`/`comparable`; keep handwritten otherwise", decisionSplit},
		{"Keep handwritten when a channel type crosses the boundary; workaround with `concurrent`/`go_interop` otherwise", decisionSplit},
		{"Keep handwritten", decisionHandwritten},
		{"Keep handwritten, or use `any` in new GALA-first files", decisionHandwritten},
		{"Boundary gap (named empty struct)", decisionHandwritten},
	}
	for _, c := range cases {
		t.Run(c.verdict, func(t *testing.T) {
			if got := classifyVerdict(c.verdict); got != c.want {
				t.Errorf("classifyVerdict(%q) = %s, want %s", c.verdict, got, c.want)
			}
		})
	}
}

// TestFileVerdict pins the file-level rule: the roster's own ladder says to
// choose the first rung that fits, so a file needs the highest rung any of its
// construct families needs.
func TestFileVerdict(t *testing.T) {
	rewrite := familyReport{family: "len-calls", decision: decisionRewrite, rung: 1}
	split := familyReport{family: "defer", decision: decisionSplit, rung: 2}
	hand := familyReport{family: "struct-tags", decision: decisionHandwritten, rung: 3}
	unknown := familyReport{family: "cap-calls", decision: decisionNone, rung: 0}

	cases := []struct {
		name        string
		reports     []familyReport
		wantRung    int
		wantDriving []string
	}{
		{"all rewrite", []familyReport{rewrite, rewrite}, 1, []string{"len-calls", "len-calls"}},
		{"a split family", []familyReport{rewrite, split}, 2, []string{"defer"}},
		{"a handwritten family", []familyReport{rewrite, split, hand}, 3, []string{"struct-tags"}},
		{"two handwritten families", []familyReport{hand, split, hand}, 3, []string{"struct-tags", "struct-tags"}},
		{"only unclassified families", []familyReport{unknown}, 0, nil},
		{"unclassified families do not decide", []familyReport{unknown, rewrite}, 1, []string{"len-calls"}},
		{"no families", nil, 0, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rung, driving := fileVerdict(c.reports)
			if rung != c.wantRung {
				t.Errorf("fileVerdict rung = %d, want %d", rung, c.wantRung)
			}
			if strings.Join(driving, ",") != strings.Join(c.wantDriving, ",") {
				t.Errorf("fileVerdict driving = %v, want %v", driving, c.wantDriving)
			}
		})
	}
}

// TestSummarizeUnknownFamily pins that a family the tool detects but the roster
// does not classify surfaces as unclassified rather than borrowing a verdict.
func TestSummarizeUnknownFamily(t *testing.T) {
	r, err := parseRoster(fixturePage)
	if err != nil {
		t.Fatalf("parseRoster: %v", err)
	}
	known := summarize("len-calls", 3, "x.go:9", r)
	if known.decision != decisionRewrite || known.rung != 1 || known.row.gapClass != "analog" {
		t.Errorf("len-calls: %+v, want a rewrite on rung 1 with gap class analog", known)
	}
	if known.count != 3 || known.rep != "x.go:9" {
		t.Errorf("len-calls: count %d rep %q, want 3 and x.go:9", known.count, known.rep)
	}
	unknown := summarize("cap-calls", 1, "x.go:4", r)
	if unknown.decision != decisionNone || unknown.rung != 0 {
		t.Errorf("cap-calls: %+v, want an unclassified family with no rung", unknown)
	}
	if unknown.row.gapClass != "" || unknown.row.verdict != "" {
		t.Errorf("cap-calls: row %+v, want no roster values", unknown.row)
	}
}

func keys(m map[string]rosterRow) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
