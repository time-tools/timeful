package main

import (
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleSource = `package sample

import (
	_ "embed"
	pgstore "timeful/server/postgres"
)

type Kind string

type Person struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

func (k Kind) String() string {
	return string(k)
}

func read(n int) (string, error) {
	var out []string
	people := map[string]Person{"a": {Name: "a"}}
	body, err := load()
	if err != nil {
		return "", err
	}
	defer body.Close()
	for i := 0; i < n; i++ {
		if len(out) > 0 && cap(out) > 0 {
			out = append(out, people["a"].Name)
		}
	}
	if s, ok := body.(stringer); ok {
		_ = s
	}
	return strings.Join(out, "\n"), nil
}
`

const containerSource = `package sample

import "strings"

type Handler func(args []string) error

func each(names []string) {
	out := []string{strings.Join(names, ",")}
	for _, n := range names {
		out = append(out, n)
	}
	f := func(x int) int { return x }
	_ = f(len(out))
}
`

// inspectSource runs the shared detection pass over one Go source and returns
// the count and first representative line of every construct family found.
func inspectSource(t *testing.T, path, src string, dns map[fileKey]bool) (map[string]int, map[string]int) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	counts, first := map[string]int{}, map[string]int{}
	in := &inspector{fset: fset, dns: dns}
	in.hit = func(family string, line int) {
		if !famSet[family] {
			return
		}
		counts[family]++
		if first[family] == 0 {
			first[family] = line
		}
	}
	in.file(parsedFile{path: path, f: f, rel: path})
	return counts, first
}

// TestInspectSingleFileCountsFamilies pins the single-file detection: every
// family in the source, its occurrence count, and the line of the
// file:line representative the preflight mode prints.
func TestInspectSingleFileCountsFamilies(t *testing.T) {
	cases := []struct {
		name     string
		src      string
		dns      map[fileKey]bool
		counts   map[string]int
		firstHit map[string]int
	}{
		{
			name: "a mixed source",
			src:  sampleSource,
			dns:  map[fileKey]bool{{".", "Kind"}: true},
			counts: map[string]int{
				"append-calls":                  1,
				"blank-imports":                 1,
				"cap-calls":                     1,
				"composite-literals":            2,
				"defer":                         1,
				"defined-non-struct-types":      1,
				"if-initializers":               1,
				"len-calls":                     1,
				"map-literals":                  1,
				"map-types":                     1,
				"methods":                       1,
				"methods-on-defined-non-struct": 1,
				"multi-return-signature":        1,
				"multi-value-define":            2,
				"named-imports":                 1,
				"struct-declarations":           1,
				"struct-tags":                   1,
				"type-assertions":               1,
			},
			// A struct tag is reported on its struct's line, the same
			// representative the whole-module walk records.
			firstHit: map[string]int{"struct-tags": 10, "len-calls": 27, "type-assertions": 31, "defer": 25},
		},
		{
			name: "container and callback shapes",
			src:  containerSource,
			counts: map[string]int{
				"function-literals":        1,
				"function-types":           1,
				"range-loops":              1,
				"slice-literals":           1,
				"append-calls":             1,
				"len-calls":                1,
				"composite-literals":       1,
				"defined-non-struct-types": 1,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			counts, first := inspectSource(t, "sample.go", c.src, c.dns)
			for family, want := range c.counts {
				if got := counts[family]; got != want {
					t.Errorf("%s: counted %d, want %d", family, got, want)
				}
			}
			for family := range counts {
				if _, ok := c.counts[family]; !ok {
					t.Errorf("%s: counted %d, want the family to be absent from this expectation", family, counts[family])
				}
			}
			for family, want := range c.firstHit {
				if got := first[family]; got != want {
					t.Errorf("%s: first hit on line %d, want %d", family, got, want)
				}
			}
		})
	}
}

// TestAddDefinedNonStruct pins the scope that decides whether a receiver is a
// defined non-struct type: the type declarations of the file's own directory.
func TestAddDefinedNonStruct(t *testing.T) {
	const types = `package sample

type Kind string

type Alias = Kind

type Person struct{ Name string }
`
	dns := map[fileKey]bool{}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "types.go", types, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	addDefinedNonStruct(fset, "types.go", f, dns)
	if !dns[fileKey{".", "Kind"}] {
		t.Errorf("Kind is not recorded as a defined non-struct type: %v", dns)
	}
	for _, name := range []string{"Person", "Alias", "Missing"} {
		if dns[fileKey{".", name}] {
			t.Errorf("%s must not be recorded as a defined non-struct type", name)
		}
	}

	// The sample's method on Kind is only recognized with that set present,
	// which is why the preflight mode reads the sibling files of its directory.
	withDNS, _ := inspectSource(t, "sample.go", sampleSource, dns)
	if withDNS["methods-on-defined-non-struct"] != 1 {
		t.Errorf("methods-on-defined-non-struct: counted %d, want 1", withDNS["methods-on-defined-non-struct"])
	}
	withoutDNS, _ := inspectSource(t, "sample.go", sampleSource, nil)
	if withoutDNS["methods-on-defined-non-struct"] != 0 {
		t.Errorf("without the type set the receiver counts as a plain method, want no methods-on-defined-non-struct hit")
	}
}

// TestPreflightExitStatuses pins that the three failure modes the mode must
// tell apart have distinct messages and distinct exit codes, and that the
// advisory verdict itself never fails.
func TestPreflightExitStatuses(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "roster.md")
	write(t, dir, "roster.md", fixturePage)

	rewritable := write(t, dir, "rewritable.go", "package p\n\nfunc f(s []int) int { return len(s) }\n")
	handwritten := write(t, dir, "handwritten.go", "package p\n\ntype P struct{ Name string `json:\"name\"` }\n")
	unclassified := write(t, dir, "unclassified.go", "package p\n\nfunc f(s []int) int { return cap(s) }\n")
	empty := write(t, dir, "empty.go", "package p\n\n// nothing recognizable here\n")
	broken := write(t, dir, "broken.go", "package p\n\nfunc f( {\n")
	noVerdictColumn := write(t, dir, "no-verdict-column.md", strings.ReplaceAll(fixturePage, "| Verdict |", "| Outcome |"))

	cases := []struct {
		name     string
		file     string
		doc      string
		wantCode int
		wantOut  []string
		wantErr  []string
	}{
		{
			name:     "a rewritable file reports the first rung",
			file:     rewritable,
			doc:      doc,
			wantCode: 0,
			wantOut:  []string{"`len-calls`", "| 1 |", `file-level verdict: rung 1, "Rewrite runtime-free."`, "every construct family in this file rewrites"},
		},
		{
			name:     "a handwritten family is advisory, not a failure",
			file:     handwritten,
			doc:      doc,
			wantCode: 0,
			wantOut:  []string{"`struct-tags`", "| language | 3 | Keep handwritten |", `file-level verdict: rung 3, "Keep handwritten."`, "driving families: `struct-tags`"},
		},
		{
			name:     "a family the roster does not classify surfaces",
			file:     unclassified,
			doc:      doc,
			wantCode: 0,
			wantOut:  []string{"| `cap-calls` | 1 |", "not classified in the roster", "undecided"},
		},
		{
			name:     "a missing file",
			file:     filepath.Join(dir, "absent.go"),
			doc:      doc,
			wantCode: 2,
			wantErr:  []string{"read "},
		},
		{
			name:     "an unparseable file",
			file:     broken,
			doc:      doc,
			wantCode: 2,
			wantErr:  []string{"parse "},
		},
		{
			name:     "a missing roster",
			file:     rewritable,
			doc:      filepath.Join(dir, "absent.md"),
			wantCode: 3,
			wantErr:  []string{"read "},
		},
		{
			name:     "a restructured roster",
			file:     rewritable,
			doc:      noVerdictColumn,
			wantCode: 3,
			wantErr:  []string{"no table with the Construct, Gap class, Verdict columns"},
		},
		{
			name:     "a file with no recognized construct family",
			file:     empty,
			doc:      doc,
			wantCode: 4,
			wantOut:  []string{"has no recognized construct family"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, out, errOut := runPreflight(t, c.file, c.doc)
			if code != c.wantCode {
				t.Errorf("exit code %d, want %d; output was:\n%s%s", code, c.wantCode, out, errOut)
			}
			for _, want := range c.wantOut {
				if !strings.Contains(out, want) {
					t.Errorf("stdout does not contain %q; stdout was:\n%s", want, out)
				}
			}
			for _, want := range c.wantErr {
				if !strings.Contains(errOut, want) {
					t.Errorf("stderr does not contain %q; stderr was:\n%s", want, errOut)
				}
			}
		})
	}
}

// runPreflight calls preflight with its streams captured and returns the exit
// code it would have exited the process with.
func runPreflight(t *testing.T, file, doc string) (int, string, string) {
	t.Helper()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	code := preflight(file, doc)
	os.Stdout, os.Stderr = origOut, origErr
	outW.Close()
	errW.Close()
	out, readErr := io.ReadAll(outR)
	if readErr != nil {
		t.Fatalf("read captured stdout: %v", readErr)
	}
	errOut, readErr := io.ReadAll(errR)
	if readErr != nil {
		t.Fatalf("read captured stderr: %v", readErr)
	}
	outR.Close()
	errR.Close()
	return code, string(out), string(errOut)
}

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}
