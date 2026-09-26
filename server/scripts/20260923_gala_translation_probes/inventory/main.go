// Command inventory counts the Go construct families in the server module and
// checks them against the Inventory section of docs/gala-translation.md.
//
// Run it from this directory:
//
//	go run ./inventory            # print the computed inventory table
//	go run ./inventory -check     # compare against the roster's Inventory section
//	go run ./inventory -preflight accounts/accounts.go
//	                              # roster triage for one Go file
//
// The walk covers the server Go module and skips nested modules, so third_party/
// and this probe corpus are out of scope.
//
// -preflight reports every construct family in one file with its occurrence
// count, a file:line representative, and the roster's gap class, verdict, and
// decision-ladder rung, then one advisory file-level verdict. Its exit statuses
// are 0 for a reported verdict, 2 for a file it cannot read or parse, 3 for a
// missing or restructured roster, and 4 for a file with no recognized construct
// family; a file the roster judges untranslatable still exits 0, because the
// verdict is advisory and gates nothing. -check keeps its own 0, 1, and 2.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type class int

const (
	handNonTest class = iota
	handTest
	galaGen
	swagGen
	numClasses
)

// Process exit codes. -check keeps returning 1 for drift and 2 for an
// unreadable roster page; the preflight statuses are its own.
const (
	exitVerdict   = 0 // the file-level verdict is advisory
	exitBadFile   = 2 // the Go file, or the flag combination, cannot be honored
	exitBadRoster = 3 // the roster page is missing or no longer shaped as expected
	exitNoFamily  = 4 // the file has no recognized construct family
)

var classNames = [numClasses]string{"hand-nontest", "hand-test", "gala-generated", "swag-generated"}

var fams = []string{
	// functions and signatures
	"multi-return-signature",
	"blank-parameters",
	"variadic-parameters",
	"generic-declarations",
	"function-literals",
	"func-literals-in-composite-literals",
	"methods",
	"methods-on-defined-non-struct",
	"function-types",

	// declarations
	"multi-value-define",
	"multi-value-var",
	"multi-value-assign",
	"const-declarations",
	"package-level-vars",
	"struct-declarations",
	"interface-declarations",
	"defined-non-struct-types",
	"type-aliases",

	// statements and control flow
	"defer",
	"go-statements",
	"switch-statements",
	"select-statements",
	"if-initializers",
	"range-loops",
	"for-loops-with-omitted-init",
	"goto",
	"fallthrough",
	"labeled-statements",

	// expressions and builtins
	"len-calls",
	"cap-calls",
	"make-calls",
	"new-calls",
	"append-calls",
	"delete-calls",
	"close-calls",
	"panic-calls",
	"recover-calls",
	"copy-calls",
	"type-assertions",
	"slice-expressions",
	"map-literals",
	"slice-literals",
	"fixed-size-array-types",
	"fixed-size-array-literals",
	"byte-slice-conversions",
	"composite-literals",
	"struct-tags",
	"interface{} type",
	"empty-struct-type",
	"anonymous-struct-types",
	"channel-types",
	"map-types",
	"embedded-struct-fields",

	// packages and files
	"init-functions",
	"main-functions",
	"blank-imports",
	"dot-imports",
	"named-imports",
	"go-embed-directives",
}

var famSet = map[string]bool{}

func init() {
	for _, f := range fams {
		famSet[f] = true
	}
}

type stat struct {
	count  [numClasses]int
	rep    [numClasses]string
	total  int
	anyRep string
}

var stats = map[string]*stat{}

func hit(family string, cl class, file string, line int) {
	st := stats[family]
	if st == nil {
		st = &stat{}
		stats[family] = st
	}
	st.count[cl]++
	st.total++
	loc := fmt.Sprintf("%s:%d", file, line)
	if st.rep[cl] == "" {
		st.rep[cl] = loc
	}
	if st.anyRep == "" {
		st.anyRep = loc
	}
}

type parsedFile struct {
	path string
	f    *ast.File
	cl   class
	rel  string
}

// fileKey identifies a declared type by directory and name, which is the scope
// the walk uses to tell a method on a defined non-struct type from a method on
// a struct.
type fileKey struct{ dir, name string }

func main() {
	server := flag.String("server", "../..", "server module root to walk")
	doc := flag.String("doc", "../../../docs/gala-translation.md", "roster page with the Inventory section")
	check := flag.Bool("check", false, "compare the computed inventory against the roster and exit non-zero on drift")
	file := flag.String("preflight", "", "report the roster's gap class, verdict, and decision-ladder rung for one Go file")
	flag.Parse()

	if *check && *file != "" {
		fmt.Fprintln(os.Stderr, "-check and -preflight are mutually exclusive")
		flag.Usage()
		os.Exit(exitBadFile)
	}
	if *file != "" {
		os.Exit(preflight(*file, *doc))
	}

	clCount, fileCount, stats := compute(*server)

	if *check {
		os.Exit(checkDoc(*doc, fileCount, clCount, stats))
	}
	printInventory(fileCount, clCount, stats)
}

// compute walks the server module, skipping nested modules, and returns the
// per-class file counts and the construct-family statistics.
func compute(server string) ([numClasses]int, int, map[string]*stat) {
	stats = map[string]*stat{}
	fset := token.NewFileSet()
	var paths []string
	filepath.Walk(server, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != server {
				if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.HasSuffix(info.Name(), ".go") {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)

	definedNonStruct := map[fileKey]bool{}
	var files []parsedFile
	for _, path := range paths {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err)
			continue
		}
		rel, _ := filepath.Rel(server, path)
		p := parsedFile{path: path, f: f, cl: classify(f, rel), rel: rel}
		files = append(files, p)
		addDefinedNonStruct(fset, path, f, definedNonStruct)
	}

	in := &inspector{fset: fset, dns: definedNonStruct}
	for _, p := range files {
		cl, file := p.cl, p.rel
		in.hit = func(family string, line int) {
			if famSet[family] {
				hit(family, cl, file, line)
			}
		}
		in.file(p)
	}

	clCount := [numClasses]int{}
	for _, p := range files {
		clCount[p.cl]++
	}
	return clCount, len(files), stats
}

// addDefinedNonStruct records the defined non-struct types a file declares, so
// that a later method declaration can be classified by the same scope the walk
// uses.
func addDefinedNonStruct(fset *token.FileSet, path string, f *ast.File, out map[fileKey]bool) {
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, sp := range gd.Specs {
			ts := sp.(*ast.TypeSpec)
			switch ts.Type.(type) {
			case *ast.StructType, *ast.InterfaceType:
			default:
				if !ts.Assign.IsValid() {
					out[fileKey{filepath.Dir(path), ts.Name.Name}] = true
				}
			}
		}
	}
}

// inspector walks one parsed file and reports every construct family in it. The
// caller supplies the hit callback, so the whole-module walk and the
// single-file preflight share one detection pass.
type inspector struct {
	fset *token.FileSet
	dns  map[fileKey]bool
	hit  func(family string, line int)
}

func (in *inspector) file(p parsedFile) {
	var stack []ast.Node
	ast.Inspect(p.f, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, n)
		line := in.fset.Position(n.Pos()).Line
		h := func(id string) {
			in.hit(id, line)
		}
		ancestor := func(level int) ast.Node {
			i := len(stack) - 1 - level
			if i < 0 {
				return nil
			}
			return stack[i]
		}
		parent := ancestor(1)
		switch v := n.(type) {
		case *ast.FuncDecl:
			if v.Recv == nil {
				if v.Name.Name == "init" {
					h("init-functions")
				}
				if v.Name.Name == "main" {
					h("main-functions")
				}
			}
			if v.Recv != nil && len(v.Recv.List) > 0 {
				h("methods")
				t := v.Recv.List[0].Type
				if se, ok := t.(*ast.StarExpr); ok {
					t = se.X
				}
				if id, ok := t.(*ast.Ident); ok && in.dns[fileKey{filepath.Dir(p.path), id.Name}] {
					h("methods-on-defined-non-struct")
				}
			}
			if v.Type.TypeParams != nil {
				h("generic-declarations")
			}
			if resultCount(v.Type) > 1 {
				h("multi-return-signature")
			}
			countParams(v.Type, h)
		case *ast.FuncLit:
			h("function-literals")
			if insideCompositeLiteral(stack) {
				h("func-literals-in-composite-literals")
			}
			if resultCount(v.Type) > 1 {
				h("multi-return-signature")
			}
			countParams(v.Type, h)
		case *ast.TypeSpec:
			if v.TypeParams != nil {
				h("generic-declarations")
			}
			switch v.Type.(type) {
			case *ast.StructType:
				h("struct-declarations")
			case *ast.InterfaceType:
				h("interface-declarations")
			default:
				if v.Assign.IsValid() {
					h("type-aliases")
				} else {
					h("defined-non-struct-types")
				}
			}
		case *ast.AssignStmt:
			if len(v.Lhs) > 1 {
				if v.Tok == token.DEFINE {
					h("multi-value-define")
				} else {
					h("multi-value-assign")
				}
			}
		case *ast.GenDecl:
			switch v.Tok {
			case token.VAR:
				if _, ok := parent.(*ast.File); ok {
					h("package-level-vars")
				}
			case token.CONST:
				h("const-declarations")
			}
		case *ast.ValueSpec:
			if len(v.Names) > 1 {
				if gd, ok := parent.(*ast.GenDecl); ok && gd.Tok == token.VAR {
					h("multi-value-var")
				}
			}
		case *ast.DeferStmt:
			h("defer")
		case *ast.GoStmt:
			h("go-statements")
		case *ast.SwitchStmt:
			h("switch-statements")
		case *ast.TypeSwitchStmt:
			h("switch-statements")
		case *ast.SelectStmt:
			h("select-statements")
		case *ast.BranchStmt:
			if v.Tok == token.GOTO {
				h("goto")
			}
			if v.Tok == token.FALLTHROUGH {
				h("fallthrough")
			}
		case *ast.IfStmt:
			if v.Init != nil {
				h("if-initializers")
			}
		case *ast.RangeStmt:
			h("range-loops")
		case *ast.ForStmt:
			if v.Init == nil && v.Post != nil && v.Cond != nil {
				h("for-loops-with-omitted-init")
			}
		case *ast.LabeledStmt:
			h("labeled-statements")
		case *ast.CallExpr:
			if id, ok := v.Fun.(*ast.Ident); ok {
				switch id.Name {
				case "len":
					h("len-calls")
				case "cap":
					h("cap-calls")
				case "make":
					h("make-calls")
				case "new":
					h("new-calls")
				case "append":
					h("append-calls")
				case "delete":
					h("delete-calls")
				case "close":
					h("close-calls")
				case "panic":
					h("panic-calls")
				case "recover":
					h("recover-calls")
				case "copy":
					h("copy-calls")
				}
			}
			if at, ok := v.Fun.(*ast.ArrayType); ok && at.Len == nil {
				if id, ok := at.Elt.(*ast.Ident); ok && id.Name == "byte" {
					h("byte-slice-conversions")
				}
			}
		case *ast.TypeAssertExpr:
			h("type-assertions")
		case *ast.SliceExpr:
			h("slice-expressions")
		case *ast.CompositeLit:
			if _, ok := v.Type.(*ast.MapType); ok {
				h("map-literals")
			}
			if at, ok := v.Type.(*ast.ArrayType); ok {
				if at.Len == nil {
					h("slice-literals")
				} else {
					h("fixed-size-array-literals")
				}
			}
			h("composite-literals")
		case *ast.ArrayType:
			if v.Len != nil {
				h("fixed-size-array-types")
			}
		case *ast.StructType:
			if _, ok := parent.(*ast.TypeSpec); !ok {
				h("anonymous-struct-types")
				if v.Fields == nil || len(v.Fields.List) == 0 {
					h("empty-struct-type")
				}
			}
			if v.Fields != nil {
				for _, fl := range v.Fields.List {
					if fl.Tag != nil {
						h("struct-tags")
					}
					if len(fl.Names) == 0 {
						h("embedded-struct-fields")
					}
				}
			}
		case *ast.InterfaceType:
			if v.Methods == nil || len(v.Methods.List) == 0 {
				h("interface{} type")
			}
		case *ast.ChanType:
			h("channel-types")
		case *ast.MapType:
			h("map-types")
		case *ast.FuncType:
			if _, ok := parent.(*ast.FuncDecl); !ok {
				if _, ok := parent.(*ast.FuncLit); !ok {
					h("function-types")
				}
			}
		case *ast.ImportSpec:
			if v.Name != nil {
				switch v.Name.Name {
				case "_":
					h("blank-imports")
				case ".":
					h("dot-imports")
				default:
					h("named-imports")
				}
			}
		}
		return true
	})
	for _, cg := range p.f.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "//go:embed") {
				in.hit("go-embed-directives", in.fset.Position(c.Pos()).Line)
			}
		}
	}
}

// preflight reports the construct families in one Go file, the roster's reading
// of each, and one advisory file-level verdict derived from the roster's
// decision ladder. It returns a process exit code.
func preflight(path, docPath string) int {
	rost, err := loadRoster(docPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "preflight: %v\n", err)
		return exitBadRoster
	}
	fset := token.NewFileSet()
	f, err := parseOne(fset, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "preflight: %v\n", err)
		return exitBadFile
	}

	// A method on a defined non-struct type is only recognizable against the
	// type declarations of its own directory, so the single-file mode reads the
	// sibling files for the same set the whole-module walk builds.
	dns := map[fileKey]bool{}
	for _, sibling := range dirSiblings(path) {
		sf, err := parseOne(fset, sibling)
		if err != nil {
			fmt.Fprintf(os.Stderr, "preflight: %v; methods on its types may be misclassified\n", err)
			continue
		}
		addDefinedNonStruct(fset, sibling, sf, dns)
	}
	addDefinedNonStruct(fset, path, f, dns)

	counts := map[string]int{}
	reps := map[string]string{}
	in := &inspector{fset: fset, dns: dns}
	in.hit = func(family string, line int) {
		if !famSet[family] {
			return
		}
		counts[family]++
		if reps[family] == "" {
			reps[family] = fmt.Sprintf("%s:%d", path, line)
		}
	}
	in.file(parsedFile{path: path, f: f, rel: path})
	if len(counts) == 0 {
		fmt.Printf("preflight: %s has no recognized construct family\n", path)
		return exitNoFamily
	}

	var reports []familyReport
	for _, family := range fams {
		count, ok := counts[family]
		if !ok {
			continue
		}
		reports = append(reports, summarize(family, count, reps[family], rost))
	}

	rung, driving := fileVerdict(reports)
	printPreflight(path, rost, reports, rung, driving)
	return exitVerdict
}

func parseOne(fset *token.FileSet, path string) (*ast.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %v", path, err)
	}
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %v", path, err)
	}
	return f, nil
}

// dirSiblings returns the other Go files in the same directory as path.
func dirSiblings(path string) []string {
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "preflight: read %s: %v\n", dir, err)
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		sibling := filepath.Join(dir, e.Name())
		if sibling != path {
			out = append(out, sibling)
		}
	}
	sort.Strings(out)
	return out
}

func printPreflight(path string, rost roster, reports []familyReport, rung int, driving []string) {
	fmt.Printf("preflight: %s\n", path)
	fmt.Printf("roster: %d construct families\n\n", len(rost.rows))
	fmt.Println("decision ladder:")
	for i, label := range rost.ladder.labels {
		fmt.Printf("  %d %s\n", i+1, label)
	}
	fmt.Println()
	fmt.Println("| Construct family | Count | Representative | Gap class | Rung | Verdict |")
	fmt.Println("| --- | --- | --- | --- | --- | --- |")
	for _, r := range reports {
		if r.rung == 0 {
			fmt.Printf("| `%s` | %d | `%s` | not classified | — | not classified in the roster |\n", r.family, r.count, r.rep)
			continue
		}
		fmt.Printf("| `%s` | %d | `%s` | %s | %d | %s |\n", r.family, r.count, r.rep, r.row.gapClass, r.rung, r.row.verdict)
	}
	fmt.Println()
	switch {
	case rung == 0:
		fmt.Printf("file-level verdict: undecided; the roster classifies none of the %d construct families in this file\n", len(reports))
	case len(driving) == len(reports):
		fmt.Printf("file-level verdict: rung %d, %q; every construct family in this file rewrites\n", rung, rost.ladder.label(rung))
	default:
		fmt.Printf("file-level verdict: rung %d, %q\n", rung, rost.ladder.label(rung))
		fmt.Printf("driving families: %s\n", codeList(driving))
	}
}

func codeList(names []string) string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = "`" + n + "`"
	}
	return strings.Join(out, ", ")
}

func printInventory(fileCount int, clCount [numClasses]int, stats map[string]*stat) {
	fmt.Printf("files parsed: %d\n", fileCount)
	fmt.Printf("files by class: %s=%d %s=%d %s=%d %s=%d\n\n",
		classNames[0], clCount[0], classNames[1], clCount[1],
		classNames[2], clCount[2], classNames[3], clCount[3])
	printTable(stats)
}

func printTable(stats map[string]*stat) {
	fmt.Println("| Construct family | Hand non-test | Hand test | GALA-generated | Swag-generated | Total | Representative |")
	fmt.Println("| --- | --- | --- | --- | --- | --- | --- |")
	for _, fm := range fams {
		st := stats[fm]
		if st == nil {
			st = &stat{}
		}
		if st.total == 0 {
			continue
		}
		rep := st.rep[handNonTest]
		if rep == "" {
			rep = st.anyRep
		}
		if rep == "" {
			rep = "—"
		}
		fmt.Printf("| `%s` | %d | %d | %d | %d | %d | `%s` |\n",
			fm, st.count[0], st.count[1], st.count[2], st.count[3], st.total, rep)
	}
}

type docRow struct {
	classCounts [numClasses]int
	total       int
	rep         string
}

var summaryRe = regexp.MustCompile(`covers (\d+) Go files: (\d+) handwritten non-test, (\d+) handwritten test, (\d+) GALA-generated, and (\d+) swag-generated`)

// checkDoc compares the computed inventory with the roster page and returns a
// process exit code.
func checkDoc(path string, fileCount int, clCount [numClasses]int, stats map[string]*stat) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		return 2
	}
	text := string(data)
	failures := 0

	summary := summaryRe.FindStringSubmatch(text)
	if summary == nil {
		fmt.Printf("FAIL summary: no inventory summary sentence found in %s\n", path)
		failures++
	} else {
		docFiles, _ := strconv.Atoi(summary[1])
		if docFiles != fileCount {
			fmt.Printf("FAIL files: doc says %d, computed %d\n", docFiles, fileCount)
			failures++
		}
		want := summary[2:]
		for i, c := range []class{handNonTest, handTest, galaGen, swagGen} {
			got, _ := strconv.Atoi(want[i])
			if got != clCount[c] {
				fmt.Printf("FAIL class %s: doc says %d, computed %d\n", classNames[c], got, clCount[c])
				failures++
			}
		}
	}

	docRows, err := parseInventoryTable(text)
	if err != nil {
		fmt.Printf("FAIL table: %v\n", err)
		return 1
	}
	for _, fm := range fams {
		st := stats[fm]
		if st == nil {
			st = &stat{}
		}
		row, inDoc := docRows[fm]
		switch {
		case st.total == 0 && inDoc:
			fmt.Printf("FAIL %s: doc lists it but the server walk found zero occurrences\n", fm)
			failures++
		case st.total > 0 && !inDoc:
			fmt.Printf("FAIL %s: missing from the doc table (computed total %d)\n", fm, st.total)
			failures++
		case st.total > 0:
			for i, c := range []class{handNonTest, handTest, galaGen, swagGen} {
				if row.classCounts[i] != st.count[c] {
					fmt.Printf("FAIL %s %s: doc says %d, computed %d\n", fm, classNames[c], row.classCounts[i], st.count[c])
					failures++
				}
			}
			if row.total != st.total {
				fmt.Printf("FAIL %s total: doc says %d, computed %d\n", fm, row.total, st.total)
				failures++
			}
			rep := st.rep[handNonTest]
			if rep == "" {
				rep = st.anyRep
			}
			if row.rep != rep {
				fmt.Printf("FAIL %s representative: doc says %s, computed %s\n", fm, row.rep, rep)
				failures++
			}
		}
	}
	for fm := range docRows {
		if !famSet[fm] {
			fmt.Printf("FAIL %s: doc lists an unknown family\n", fm)
			failures++
		}
	}

	if failures > 0 {
		fmt.Printf("\ninventory check: %d mismatch(es)\n", failures)
		return 1
	}
	fmt.Printf("inventory check: doc matches (%d files)\n", fileCount)
	return 0
}

// parseInventoryTable locates the Inventory table and returns it keyed by
// construct family.
func parseInventoryTable(text string) (map[string]docRow, error) {
	rows := map[string]docRow{}
	lines := strings.Split(text, "\n")
	inTable := false
	for _, line := range lines {
		if !strings.HasPrefix(line, "|") {
			if inTable && strings.TrimSpace(line) == "" {
				break
			}
			continue
		}
		cells := splitRow(line)
		if !inTable {
			if len(cells) > 1 && cells[0] == "Construct family" {
				inTable = true
			}
			continue
		}
		if len(cells) < 7 || strings.Trim(cells[0], "- ") == "" {
			continue
		}
		fam := strings.Trim(cells[0], "`")
		if fam == "" || fam == "Construct family" {
			continue
		}
		row := docRow{rep: strings.Trim(cells[6], "`")}
		for i := range []int{0, 1, 2, 3} {
			n, err := strconv.Atoi(strings.TrimSpace(cells[i+1]))
			if err != nil {
				return nil, fmt.Errorf("row %s has a non-numeric count %q", fam, cells[i+1])
			}
			row.classCounts[i] = n
		}
		total, err := strconv.Atoi(strings.TrimSpace(cells[5]))
		if err != nil {
			return nil, fmt.Errorf("row %s has a non-numeric total %q", fam, cells[5])
		}
		row.total = total
		rows[fam] = row
	}
	if !inTable {
		return nil, fmt.Errorf("no table with a Construct family header")
	}
	return rows, nil
}

func splitRow(line string) []string {
	row, _ := tableCells(line)
	return row
}

func classify(f *ast.File, rel string) class {
	generated := false
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "DO NOT EDIT") {
				generated = true
				break
			}
		}
		if generated {
			break
		}
	}
	test := strings.HasSuffix(rel, "_test.go")
	switch {
	case !generated && !test:
		return handNonTest
	case !generated && test:
		return handTest
	case generated && strings.HasPrefix(rel, "docs/"):
		return swagGen
	default:
		return galaGen
	}
}

func resultCount(ft *ast.FuncType) int {
	if ft.Results == nil {
		return 0
	}
	n := 0
	for _, f := range ft.Results.List {
		if len(f.Names) == 0 {
			n++
		} else {
			n += len(f.Names)
		}
	}
	return n
}

func countParams(ft *ast.FuncType, h func(string)) {
	if ft.Params == nil {
		return
	}
	for _, f := range ft.Params.List {
		if _, ok := f.Type.(*ast.Ellipsis); ok {
			h("variadic-parameters")
		}
		for _, n := range f.Names {
			if n.Name == "_" {
				h("blank-parameters")
			}
		}
	}
}

func insideCompositeLiteral(stack []ast.Node) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		if _, ok := stack[i].(*ast.CompositeLit); ok {
			return true
		}
		if _, ok := stack[i].(*ast.FuncDecl); ok {
			return false
		}
	}
	return false
}
