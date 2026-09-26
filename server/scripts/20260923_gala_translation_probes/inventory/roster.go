// Roster reading for the single-file preflight mode.
//
// docs/gala-translation.md carries several tables that share column names: the
// construct roster, the replacement summary, the gap table, and the inventory
// table all have a construct-family-shaped first column, and the replacement
// summary repeats the Construct and Gap class headers of the roster. The reader
// is therefore bounded to the construct roster section, maps columns by header
// name rather than by position, and reads the decision ladder labels from the
// page so that no verdict is restated in this tool.

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	rosterSection = "Construct roster"
	ladderSection = "Decision ladder"
)

// rosterColumns are the columns a table must name to be the construct roster.
var rosterColumns = []string{"Construct", "Gap class", "Verdict"}

// rungRe matches a numbered decision-ladder rung and captures its bold label.
var rungRe = regexp.MustCompile(`^[0-9]+\.\s+\*\*(.+?)\*\*`)

type rosterRow struct {
	family   string
	gapClass string
	verdict  string
}

// ladder holds the roster's own decision ladder. A rung number of zero means
// the page no longer names that rung.
type ladder struct {
	labels      []string
	rewrite     int
	split       int
	handwritten int
}

// rung maps a decision onto its ladder rung number.
func (l ladder) rung(d decision) int {
	switch d {
	case decisionRewrite:
		return l.rewrite
	case decisionSplit:
		return l.split
	case decisionHandwritten:
		return l.handwritten
	}
	return 0
}

// label returns the roster's verbatim text for a rung number.
func (l ladder) label(n int) string {
	if n < 1 || n > len(l.labels) {
		return ""
	}
	return l.labels[n-1]
}

type roster struct {
	rows   map[string]rosterRow
	ladder ladder
}

type rosterError struct{ msg string }

func (e *rosterError) Error() string { return e.msg }

func rosterErrf(format string, args ...any) error {
	return &rosterError{msg: fmt.Sprintf(format, args...)}
}

// loadRoster reads the roster page from disk.
func loadRoster(path string) (roster, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return roster{}, rosterErrf("read %s: %v", path, err)
	}
	return parseRoster(string(data))
}

// parseRoster reads the construct roster tables and the decision ladder out of
// the roster page. Every failure is a roster error, because the page is either
// missing or no longer shaped the way the preflight mode reads it.
func parseRoster(text string) (roster, error) {
	body, err := section(text, rosterSection)
	if err != nil {
		return roster{}, err
	}
	rows, tables, err := readRosterTables(body)
	if err != nil {
		return roster{}, err
	}
	if tables == 0 {
		return roster{}, rosterErrf("the %q section has no table with the %s columns", rosterSection, strings.Join(rosterColumns, ", "))
	}
	body, err = section(text, ladderSection)
	if err != nil {
		return roster{}, err
	}
	lad, err := readLadder(body)
	if err != nil {
		return roster{}, err
	}
	return roster{rows: rows, ladder: lad}, nil
}

// section returns the lines of the level-two Markdown section titled title.
// Headings inside a fenced code block are ignored, and the section ends at the
// next level-two heading, so a table in another section cannot contribute rows.
func section(text, title string) ([]string, error) {
	var body []string
	found, fenced := false, false
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
		}
		if !fenced {
			if name, ok := heading2(line); ok {
				if found {
					return body, nil
				}
				if name == title {
					found = true
				}
				continue
			}
		}
		if found {
			body = append(body, line)
		}
	}
	if found {
		return body, nil
	}
	return nil, rosterErrf("no %q section", title)
}

// heading2 reports whether line is a level-two Markdown heading and returns its
// title.
func heading2(line string) (string, bool) {
	if !strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimRight(strings.TrimSpace(line[3:]), "#")), true
}

// readRosterTables returns every construct-roster row in the section body,
// keyed by family, and the number of tables it read.
func readRosterTables(body []string) (map[string]rosterRow, int, error) {
	rows := map[string]rosterRow{}
	tables := 0
	for i := 0; i < len(body); i++ {
		cells, ok := tableCells(body[i])
		if !ok {
			continue
		}
		cols, ok := columnIndex(cells, rosterColumns)
		if !ok {
			continue
		}
		tables++
		for i+1 < len(body) {
			cells, ok = tableCells(body[i+1])
			if !ok {
				break
			}
			i++
			if isSeparator(cells) {
				continue
			}
			row := rosterRow{
				family:   strings.Trim(cells[cols["Construct"]], "`"),
				gapClass: cells[cols["Gap class"]],
				verdict:  cells[cols["Verdict"]],
			}
			if row.family == "" {
				continue
			}
			if row.gapClass == "" || row.verdict == "" {
				return nil, tables, rosterErrf("roster row %s has an empty Gap class or Verdict cell", row.family)
			}
			rows[row.family] = row
		}
	}
	return rows, tables, nil
}

// columnIndex maps the named header cells onto their positions, so that adding
// or reordering a column cannot move a verdict into the wrong field.
func columnIndex(cells, want []string) (map[string]int, bool) {
	cols := map[string]int{}
	for i, c := range cells {
		cols[strings.Trim(c, "`")] = i
	}
	for _, name := range want {
		if _, ok := cols[name]; !ok {
			return nil, false
		}
	}
	return cols, true
}

// tableCells splits a Markdown table row into its trimmed cells.
func tableCells(line string) ([]string, bool) {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "|") {
		return nil, false
	}
	parts := strings.Split(t, "|")[1:]
	if n := len(parts); n > 0 && strings.TrimSpace(parts[n-1]) == "" {
		parts = parts[:n-1]
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts, true
}

// isSeparator reports whether every cell of a table row is a Markdown column
// divider, which is how the syntax-only second row of a table is recognised.
func isSeparator(row []string) bool {
	for _, c := range row {
		if strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}

// readLadder reads the ordered rungs of the decision ladder. The rung labels
// stay the roster's text, so a reworded ladder changes the tool's output
// instead of silently going stale.
func readLadder(body []string) (ladder, error) {
	var labels []string
	for _, line := range body {
		if m := rungRe.FindStringSubmatch(line); m != nil {
			labels = append(labels, m[1])
		}
	}
	if len(labels) < 3 {
		return ladder{}, rosterErrf("the %q section has %d numbered rungs, want at least 3", ladderSection, len(labels))
	}
	lad := ladder{labels: labels}
	for i, label := range labels {
		l := strings.ToLower(label)
		switch {
		case strings.Contains(l, "split"):
			lad.split = i + 1
		case strings.Contains(l, "keep handwritten"):
			lad.handwritten = i + 1
		case i == 0 && strings.Contains(l, "rewrite"):
			lad.rewrite = 1
		}
	}
	if lad.rewrite == 0 || lad.split == 0 || lad.handwritten == 0 {
		return ladder{}, rosterErrf("the %q section does not name a rewrite rung, a split rung, and a keep-handwritten rung", ladderSection)
	}
	return lad, nil
}

// decision is the file-strategy reading of one roster verdict cell.
type decision int

const (
	decisionNone decision = iota
	decisionRewrite
	decisionSplit
	decisionHandwritten
)

func (d decision) String() string {
	switch d {
	case decisionRewrite:
		return "rewrite"
	case decisionSplit:
		return "split"
	case decisionHandwritten:
		return "handwritten"
	}
	return "unclassified"
}

// Markers from the roster's own vocabulary rather than per-family verdicts: the
// Gap classification section defines a direct form, an analog, and a
// workaround as the substitute outcomes, and the verdict cells name a
// handwritten outcome in prose.
var (
	splitMarker         = "split"
	handwrittenMarkers  = []string{"keep handwritten", "stay handwritten", "stays handwritten", "handwritten sibling"}
	handwrittenNegation = "no handwritten sibling"
	substituteMarkers   = []string{"direct form", "workaround", "analog"}
	gapMarker           = "gap"
)

// classifyVerdict reads a roster verdict cell as a file strategy.
//
// The rule is mechanical and conservative:
//
//   - a cell that asks for a split is a split;
//   - a cell that offers a substitute and also names a handwritten outcome is a
//     split, because only part of the family rewrites;
//   - a cell whose only outcome is handwritten, or that names a gap, keeps the
//     file handwritten;
//   - any other cell is a rewrite.
//
// A negated handwritten clause such as "no handwritten sibling needed" is not a
// handwritten outcome.
func classifyVerdict(verdict string) decision {
	v := strings.ToLower(verdict)
	if strings.Contains(v, splitMarker) {
		return decisionSplit
	}
	v = strings.ReplaceAll(v, handwrittenNegation, "")
	hand, sub := containsAny(v, handwrittenMarkers), containsAny(v, substituteMarkers)
	switch {
	case hand && sub:
		return decisionSplit
	case hand, strings.Contains(v, gapMarker):
		return decisionHandwritten
	}
	return decisionRewrite
}

func containsAny(s string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

// familyReport is one construct family found in the preflighted file, with the
// roster's reading of it.
type familyReport struct {
	family   string
	count    int
	rep      string
	row      rosterRow
	decision decision
	rung     int
}

// fileVerdict reduces the per-family decisions to one file-level verdict by
// taking the most conservative rung any family reaches, which is the roster's
// own "choose the first rung that fits" rule read in reverse: a file needs the
// highest rung its construct families need. It returns the rung number, zero
// when no family is classified, and the families that reached it.
func fileVerdict(reports []familyReport) (int, []string) {
	best := 0
	var driving []string
	for _, r := range reports {
		switch {
		case r.rung > best:
			best, driving = r.rung, []string{r.family}
		case r.rung == best && r.rung > 0:
			driving = append(driving, r.family)
		}
	}
	return best, driving
}

// summarize attaches the roster's row, decision, and rung to one detected
// family. A family the tool detects but the roster does not classify keeps a
// zero rung, which is a roster defect to surface rather than a row to invent.
func summarize(family string, count int, rep string, r roster) familyReport {
	out := familyReport{family: family, count: count, rep: rep}
	row, ok := r.rows[family]
	if !ok {
		return out
	}
	out.row = row
	out.decision = classifyVerdict(row.verdict)
	out.rung = r.ladder.rung(out.decision)
	return out
}
