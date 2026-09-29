# GALA rewrite playbook

This page is the ordered procedure for turning one Go file under `server/` into a GALA twin: a committed `.gala` source beside the committed Go file the transpiler generates from it.
It is written for an agent that has to choose a file, act, verify, and stop without inferring the order of operations.
It routes to the knowledge rather than copying it.

Three documents own that knowledge, and this page adds nothing to any of them.
[`gala-translation.md`](gala-translation.md) is the source of truth for what GALA does with each Go construct, how each construct is classified, and how each substitution is spelled.
[`../server/GALA.md`](../server/GALA.md) records what the transpiler actually did to server files during the spikes, including the failures and the two output styles.
The probe corpus under `server/scripts/20260923_gala_translation_probes/` is what keeps both of them honest.
Read the roster for a verdict, never this page, and read it before writing a line of GALA.

Every command below is written for the repository root unless the step says otherwise, and each one is the command the file that owns it documents.

## Preconditions

Establish these four before touching a candidate, because each of them silently invalidates the result rather than failing loudly.

- The `gala` CLI on `PATH` reports the version the committed twins were produced with.
  Regeneration with that same version is deterministic, and the corpus runner refuses to start on any other version unless `GALA_PROBES_ALLOW_ANY_VERSION=1` is set.
- The Go toolchain is in the series the corpus runner pins, because the corpus' build-failure expectations embed Go compiler diagnostics, and `GALA_PROBES_ALLOW_ANY_GO=1` is the only supported override.
  The server module declares its own language version in `server/go.mod`, and the installed toolchain has to satisfy both.
- The vendored runtime at `server/third_party/gala/` is present and unmodified.
  `server/go.mod` requires `martianoff/gala` and replaces it with that directory, so a generated file that imports a vendored subpackage resolves with no network fetch and no `go.sum` entry.
- Both halves of a twin get committed: the `.gala` source and the generated `.go` file.
  The Go build never invokes GALA, so a generated file with no committed source is unreadable to the next maintainer, and a source with no committed twin is not in the build at all.

The transpiler writes an analysis cache under `server/.gala/`, which is gitignored, and a `DO NOT EDIT` header on every file it generates.

## Step 1: Select a candidate

Choose the file, then read the two places that describe the package it lives in before forming an opinion about it.

- `server/GALA.md` states the committed style for that package, whether a twin there is runtime-free or runtime-enabled, and which members already live in a handwritten sibling.
- The [decision ladder](gala-translation.md#decision-ladder) in the roster is the rule that decides the file's fate, and the packages it already names are the ones with a proven rewrite.

A package that already carries a `.gala` source is the cheap case: it has a style to copy, a sibling convention to follow, and a regeneration command to extend.
A file whose exported Go API is consumed far outside its package is the expensive case, whatever the file looks like internally.

Do not decide the outcome from the file's own appearance.
A construct the roster classifies as blocked and one it classifies as a direct form are both plain Go until the roster is consulted, and consulting it is the next step rather than a judgement call made here.

## Step 2: Pre-triage the candidate

Run the preflight from the corpus directory, which is where the tool lives and where its default path to the roster page resolves:

```sh
cd server/scripts/20260923_gala_translation_probes
go run ./inventory -preflight ../../<package>/<file>.go
```

It prints the roster's decision ladder in the roster's own words, then one row per construct family the file contains with an occurrence count, a `file:line` representative, that family's gap class, the ladder rung its verdict reaches, and finally one file-level verdict.
Read the rung from that output rather than from this page, because the labels are the roster's and the roster can change them.

Act on the file-level verdict:

| File-level verdict                                             | What to do                                                                                                                     |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| rung 1                                                         | Rewrite the whole file, and continue to step 3.                                                                                |
| rung 2                                                         | Transpile the part the roster can express and move the rest into a handwritten sibling, cutting between the members at step 4. |
| rung 3                                                         | Stop, and leave the file handwritten. The roster's own wording for that rung is the reason.                                    |
| no verdict, because the roster classifies none of the families | Stop, and treat the missing classification as a roster gap to fill rather than a file to attempt.                              |

The verdict is advisory, so the exit status is `0` for every row in that table, including the one that says stop; the exit status never decides the outcome.
A row in the family table that reads `not classified in the roster` is a construct family the roster does not cover.
The file-level verdict ignores such a family, so its presence is blocking until the roster classifies it: add the row to the roster instead of guessing a verdict.

The other exit statuses mean the tool could not answer, not that the file is blocked:

| Exit status | Meaning                                                            | What to do                                                                                                      |
| ----------- | ------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------- |
| `2`         | the path could not be read, or the file did not parse as Go        | check the path and confirm the argument is the Go file rather than its `.gala` source, which is a parse failure |
| `3`         | the roster page is missing or no longer shaped as the tool expects | stop, because the tool cannot gate anything it cannot read, and repair the roster before retrying               |
| `4`         | the file contains no construct family the roster classifies        | check the path, because this status means the tool walked the file and found nothing it recognises              |

## Step 3: Choose the output style

Two styles are committable, and `server/GALA.md` describes both in full with the committed examples.

- Runtime-free output suits a leaf twin whose exported Go API has to stay exactly Go-shaped: import Go packages directly, bind raw Go values with `var`, leave the GALA standard library alone, and take the builtin's substitute from the roster's mechanical rewrites.
- Runtime-enabled output suits a file that benefits from GALA-native features and whose Go callers tolerate the shapes those features emit; the vendored runtime already resolves through the single local replace, so nothing extra is needed to compile it.

Two shape rules constrain the choice, and they are the reason the style is a decision rather than a preference.
The roster states both at the end of its [decision ladder](gala-translation.md#decision-ladder).

- An exported package-level `val` emits a Go variable that wraps the value in the runtime's immutable type, so a Go caller sees that wrapper instead of the value and has to unwrap it.
  Keep `var` for anything that crosses into Go, and reserve `val` for a local, GALA-internal binding.
- A GALA struct declaration is not a drop-in twin: the transpiler synthesizes exported `Copy`, `Equal`, `Unapply`, and `Is<Type>` members on it, and it wraps a field in the runtime's immutable type unless that field is declared `var`.
  A struct whose Go API has to stay exactly as its callers expect therefore stays handwritten, and the file's functions can still be transpiled around it.

Decide the style before writing the source, because it determines which rows of the roster's mechanical rewrites apply and which Go packages the generated file will import.

## Step 4: Write the `.gala` source

Copy the Go file to `<name>.gala` in the same directory as the Go file, then rewrite it one construct at a time.
The roster's [mechanical rewrites](gala-translation.md#mechanical-rewrites) section is the work list: each row is a Go spelling, the GALA spelling that replaces it, the condition that selects that branch, the construct families it covers, and the pin that holds it in place.
Apply the rows as written rather than paraphrasing them, because a row is copyable and its prose is not.

Two things in that section govern the cut between transpiled and handwritten:

- A construct family in the no-rewrite index at the end of the section has no rule to apply, and the roster records that absence rather than leaving it to be inferred: either the Go shape is already what GALA wants, or GALA cannot express it and that member belongs in a handwritten sibling, which is what makes the file a rung-2 split.
- A decision rule whose branches are not purely mechanical is written up in the roster's [workarounds and contested verdicts](gala-translation.md#workarounds-and-contested-verdicts) section, which also names the trap attached to each one; read that section before choosing a branch.

Then transpile to a scratch path first, so a parse error costs nothing:

```sh
cd server/<package>
gala transpile -i <name>.gala -o /tmp/<name>.go
```

Every `.gala` file imports everything it uses.
Imports do not propagate between sibling `.gala` files in a package, so a name that another file imports still has to be imported here, and the transpiler says so when it is missing.

## Step 5: Lay out the twin

- The `.gala` source sits beside the generated `.go` file in the same package directory, and both are committed.
- The generated file is never hand-edited.
  It carries a `DO NOT EDIT` header, the next transpile overwrites it, and a hand edit is lost without a trace.
- The package comment lives in a handwritten `doc.go` in the same package, because the transpiler emits no comments at all; `eventid/doc.go` and `services/providerconfig/doc.go` are the committed examples.
  A comment attached to an exported declaration is lost the moment its file is transpiled, so move the prose rather than rewriting it.
- A member GALA cannot express stays in a handwritten `.go` sibling in the same package; `appenv/appenv_port.go`, `utils/array_utils_extra.go`, and `slackbot/commands/utils.go` are the committed examples.
  A sibling is not generated, so it has no regeneration command and no header.
- Every `.gala` file in the package imports what it uses, as above.

## Step 6: Verify the twin

Run these in order.
Each one catches a failure the others do not, and a step skipped here is a step that will fail later in someone else's checkout.

1. Transpile from the package directory.

   ```sh
   cd server/<package>
   gala transpile -i <name>.gala -o <name>.go
   ```

   The `//line` directives the transpiler emits name the source path it was given, so running this from anywhere else writes a directive that does not resolve from the generated file and that the next correct regeneration then changes.
   Run it from the package directory every time, and read the emitted directives before moving on.

2. Build the server module.

   ```sh
   cd server
   go build ./...
   ```

3. Confirm the generated file is formatted.

   ```sh
   cd server
   gofmt -l <package>/<name>.go
   ```

   It reports nothing; a path in its output means the transpiler emitted unformatted Go.

4. Run the tests with the canonical backend sequence, which [`../server/README.md`](../server/README.md) owns and `docs/environments.md` documents.

   ```sh
   cp .env.test.example .env.test
   docker volume create timeful-test-go-build-cache timeful-test-go-mod-cache
   docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test postgres-test-bootstrap postgres-test-migrate
   docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test
   ```

   A package with no test of its own still needs this, because the twin changes what its callers compile against.

5. Prove regeneration is byte-identical by running the transpile command from step one a second time and requiring no diff.

   ```sh
   git diff --exit-code -- server/<package>/<name>.go
   ```

6. Run the corpus and the inventory checks.
   These are not about the twin's behaviour; they are the checks that catch a twin whose shape contradicts the roster or the corpus.

   ```sh
   server/scripts/20260923_gala_translation_probes/run.sh
   server/scripts/20260923_gala_translation_probes/audit.sh
   cd server/scripts/20260923_gala_translation_probes && go run ./inventory -check
   ```

   `run.sh` is the version-gated probe corpus and needs the pinned toolchains; `audit.sh` is read-only and needs neither, because it checks the roster, the corpus, and `server/GALA.md` against each other; `-check` re-derives the roster's inventory counts from a fresh walk of the server module.
   Adding a twin changes that module, so `-check` is expected to fail here until step 7 updates the roster's inventory summary sentence and table, and it is the signal for that edit rather than a defect in the twin.

## Step 7: Register the twin

A twin that is not registered is not finished, because the next agent reads these tables as the list of what exists.

| Where                                        | What gains a row or a count                                                                                                                                                                                       |
| -------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [`../server/README.md`](../server/README.md) | one row in the transpiled GALA sources table, carrying the generated file, its `.gala` source, and its regeneration command; and the handwritten-siblings sentence, when the package gained one                   |
| [`../server/GALA.md`](../server/GALA.md)     | one row in the current-usage table, carrying the package, the `.gala` source, the generated Go, and the handwritten sibling; and the sentence above that table, which counts the packages and the `.gala` sources |
| [`gala-translation.md`](gala-translation.md) | the inventory summary sentence and the inventory table, both re-derived by `go run ./inventory` from the corpus directory and re-checked with `go run ./inventory -check`                                         |

The roster's construct verdicts, gap classes, and mechanical rewrite rows do not change when a twin is added, because they are derived from construct families and from probes rather than from the list of files.
Run `audit.sh` after the roster edit to confirm the page is still self-consistent.

## Abort rules

Stop and leave the file handwritten when any one of these holds.
Aborting is a correct outcome of this procedure, not a failure of it, and the reason belongs in the task that asked for the rewrite.

- The preflight file-level verdict is rung 3.
  The roster's own wording for that rung is the reason, and re-deriving it is how a wire-facing Go API gets silently reshaped.
- The rewrite would change a Go API or a wire format that callers outside the package depend on, which the roster's [gap classification](gala-translation.md#gap-classification) records as what a boundary gap costs.
- A member's construct family is in the roster's no-rewrite index and moving it to a handwritten sibling would leave less in the transpiled file than the split is worth; a file that is mostly such members is not a rewrite candidate.
- A transpile, build, or test failure survives applying the roster's documented rewrite for that construct.
  That is evidence the roster is out of date for this compiler, and the fix is a roster change with a probe behind it, not an invented rewrite.
- The only way to make the file work is a transpiler version other than the pinned one.
  Changing the version is a separate exercise that re-baselines the corpus and the twins, and it is recorded in the roster's revisit section.

## Traps

These transpile cleanly and fail afterwards, either at build time, at the Go call boundary, or on the next regeneration.
Each is a transpiler behaviour that `server/GALA.md` and the roster's [workarounds and contested verdicts](gala-translation.md#workarounds-and-contested-verdicts) section record; that section is where the pin and the probe behind each one live.

| Trap                                                                                           | How it fails                                                                                                                                                                                                                                                                                                                                                   | What to do                                                                                                                                                |
| ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.Size()` on a field whose type is declared in a handwritten `.go` sibling in the same package | the transpiler only knows a field's type from a GALA declaration, so with the struct declared in the sibling it emits the call through the field and `go build` fails on a method the Go type does not have; this is specific to a package that mixes a `.gala` file with a handwritten sibling, and the same call on a GALA-declared field is not the problem | compare the value against `""` instead, which is what the committed twins do, and take the byte-count and character-count spellings apart before choosing |
| A bare type or value name that an imported package also exports                                | the transpiler resolves the bare name to the import, so the twin calls the imported type instead of the local one and `go build` fails on a mismatched type                                                                                                                                                                                                    | never name the type in the `.gala` file; call a handwritten constructor in the sibling instead, as the Slack command twin does                            |
| An exported package-level `val`                                                                | it transpiles, and the Go caller's own source stops compiling because the exported Go type changed                                                                                                                                                                                                                                                             | keep `var` for anything that crosses into Go                                                                                                              |
| A GALA struct declared for a type Go callers already construct                                 | it transpiles, and Go callers see synthesized exported members and wrapped fields they did not ask for                                                                                                                                                                                                                                                         | keep the struct handwritten and transpile the functions around it                                                                                         |
| A multi-value receive left in its `:=` spelling                                                | the transpiler panics rather than emitting a diagnostic                                                                                                                                                                                                                                                                                                        | take the `var` spelling from the roster's mechanical rewrites                                                                                             |
| A documentation comment moved into a `.gala` source                                            | it transpiles, and the comment is silently absent from the twin, so a later reader sees an undocumented export                                                                                                                                                                                                                                                 | put the prose in `doc.go` or in a handwritten sibling                                                                                                     |
| A transpile run from the repository root                                                       | it transpiles, and the emitted `//line` directives name a path that does not resolve from the generated file, which then shows up as a diff on every regeneration                                                                                                                                                                                              | run it from the package directory                                                                                                                         |
| `.Size()` chosen where the Go code meant a byte count                                          | it transpiles, it builds, and the count is wrong, because the two spellings do not count the same thing                                                                                                                                                                                                                                                        | match the original intent; the roster's mechanical rewrites separate the byte count from the character count                                              |

## Definition of done for one file

- [ ] The preflight file-level verdict for the file is recorded, with the rung it returned and the construct families that drove it.
- [ ] The `.gala` source and the generated `.go` file are both committed, side by side, and the generated file was not hand-edited.
- [ ] Every construct the rewrite touched is covered by a row in the roster's mechanical rewrites, and every member with no substitute is in a handwritten sibling or in `doc.go`.
- [ ] The package comment is in a handwritten `doc.go`, and no documentation comment was left in a `.gala` source.
- [ ] The regeneration command is recorded next to the twin in `server/README.md`, and the package row, the sentence above it, and the roster's inventory summary sentence and table are all updated.
- [ ] Transpiling twice from the package directory produces no diff, the emitted `//line` directives are relative, and `gofmt -l` on the twin is empty.
- [ ] `go build ./...` in `server/` passes and the canonical backend test sequence passes.
- [ ] `run.sh`, `audit.sh`, and `go run ./inventory -check` all pass.
- [ ] Any step the procedure could not answer is recorded as a follow-up, rather than worked around silently.
