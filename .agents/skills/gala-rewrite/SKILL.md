---
name: gala-rewrite
description: Rewrite a Go file under server/ as a committed GALA twin by following the GALA rewrite playbook and running the inventory preflight: use when the user asks to rewrite, port, translate, or convert a server file to GALA, or asks whether a server file is a GALA rewrite candidate.
---

# GALA Rewrite

A GALA twin is a committed `.gala` source beside the committed Go file the transpiler generates from it.
Both halves are committed because the Go build never invokes GALA, so neither half is useful alone.

This skill routes rather than restates.
The ordered procedure, the decisions, and the reasons live in `docs/gala-rewrite-playbook.md`, and the verdict for every Go construct lives in `docs/gala-translation.md`.
Follow the playbook and read a verdict from the roster; do not work from this file or from recalled knowledge of a spike session.
Where this file names a page for a judgement, that page owns the judgement and this file deliberately carries no copy of it.

## Where The Knowledge Lives

Ask each page only the question it owns.

| Question                                                                                                               | Page that owns it                                                                                                             |
| ---------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| What is the order of operations, and what does each step decide?                                                       | `docs/gala-rewrite-playbook.md`                                                                                               |
| What does GALA do with this Go construct, and how is each construct classified?                                        | `docs/gala-translation.md`, the translation roster                                                                            |
| What did the transpiler actually do to real server files, including what failed and which output styles are committed? | `server/GALA.md`, the spike record                                                                                            |
| Is this specific file a candidate, and what is its file-level verdict?                                                 | the preflight command, in the next section                                                                                    |
| What has actually been proven about each construct?                                                                    | the probe corpus under `server/scripts/20260923_gala_translation_probes/`, which keeps the roster and the spike record honest |

The roster is the source of truth for a verdict, and the probe corpus is the check that the roster has not drifted.
Read the roster before writing a line of GALA.

## Preflight The Candidate First

Run this before forming an opinion about any file, because a construct the roster blocks and one it treats as a direct form are both plain Go until the roster is consulted.

```sh
cd server/scripts/20260923_gala_translation_probes
go run ./inventory -preflight ../../<package>/<file>.go
```

It prints the decision ladder in the roster's own words, then one row per construct family with an occurrence count, a `file:line` representative, that family's gap class, the rung its verdict reaches, and one file-level verdict.
Read the rung labels from that output, not from the playbook and not from this skill, because the labels are the roster's and the roster can change them.

These facts about the command are the reason it is safe to rely on at all:

- The verdict is advisory, so the shell status is `0` for every verdict including the one that says stop, and the verdict line rather than the status decides the outcome.
- A tool failure does not surface as its own number through `go run`, which prints `exit status <n>` on stderr and exits `1`, so a non-zero status means the tool could not answer rather than that the file is blocked.
- The playbook's abort-rules section maps each status to what to do about it.

Treat any construct family the roster does not classify as blocking until the roster classifies it, because the file-level verdict ignores such a family and a verdict line that reads like a decision is not one.

## The Steps

Seven steps, in this order, each owning one decision.
The playbook section named for each step is the authority for that step; read it before acting rather than after failing.

1. Select a candidate, and read the spike record and the roster's decision ladder for its package before forming an opinion.
   Playbook: step 1.
2. Pre-triage it with the preflight command above, and act on the file-level verdict.
   Playbook: step 2.
3. Choose the output style, which is a decision rather than a preference because two shape rules constrain it.
   Playbook: step 3.
4. Write the `.gala` source, one construct at a time, from the roster's mechanical rewrites, and transpile to a scratch path first so a parse error costs nothing.
   Playbook: step 4.
5. Lay out the twin, which fixes where each artifact lives and which comments move to `doc.go`.
   Playbook: step 5.
6. Verify the twin, which is the step a rewrite is most often skipped past.
   Playbook: step 6.
7. Register the twin in every document and count that lists what exists, or the next agent reads stale tables as the truth.
   Playbook: step 7.

### Commands

These are the commands the steps act on, each owned by the file named beside it.
Every command is written for the repository root unless the step says otherwise, and the transpile must be run from the package directory.

```sh
# Step 2: pre-triage, from the corpus directory
cd server/scripts/20260923_gala_translation_probes && go run ./inventory -preflight ../../<package>/<file>.go

# Step 4: transpile to a scratch path, from the package directory
cd server/<package> && gala transpile -i <name>.gala -o /tmp/<name>.go

# Step 6: regenerate, build, format, prove regeneration is byte-identical
cd server/<package> && gala transpile -i <name>.gala -o <name>.go
cd server && go build ./...
cd server && gofmt -l <package>/<name>.go
git diff --exit-code -- server/<package>/<name>.go

# Step 6: corpus, roster audit, and inventory drift, after registration
server/scripts/20260923_gala_translation_probes/run.sh
server/scripts/20260923_gala_translation_probes/audit.sh
cd server/scripts/20260923_gala_translation_probes && go run ./inventory -check
```

The backend test sequence in step 6 is not reproduced here.
`server/README.md` declares the canonical sequence for backend tests and `docs/environments.md` documents its isolation semantics, so a third copy of those commands is a place they can drift; run the sequence from `server/README.md`.

One expected failure belongs in step 6: `go run ./inventory -check` is meant to fail there, because adding a twin changes the module and step 7 has not yet re-derived the roster's inventory counts.
That failure is the signal to register the twin, not a defect in the twin.

### Preconditions

Four things silently invalidate a rewrite rather than failing loudly, so establish them before step 1.
The playbook's preconditions section states each one and why it matters.

- The `gala` CLI on `PATH` is the version the committed twins were produced with, because regeneration on any other version is not comparable and the corpus runner refuses to start on it.
- The Go toolchain is in the series the corpus runner pins, because the corpus' build-failure expectations embed Go compiler diagnostics.
- The vendored runtime at `server/third_party/gala/` is present and unmodified, and `server/go.mod` replaces the module with it, so a generated file resolves with no network fetch.
- Both halves of the twin are committed.

If a precondition does not hold, say so and stop rather than proceeding on a different version, because a rewrite produced under the wrong toolchain is worse than no rewrite.

## Stop And Report Instead

Leave the file handwritten and report the reason whenever any of these holds.
Aborting is a correct outcome of this procedure, not a failure of it, and the playbook's abort-rules section gives the reasoning for each condition.

- The preflight file-level verdict is the top rung, the one whose roster wording is to keep the file handwritten, and re-deriving it is how a wire-facing Go API gets silently reshaped.
- The rewrite would change a Go API or a wire format that callers outside the package depend on.
- A member's construct family has no rewrite in the roster, and moving it to a handwritten sibling would leave less in the transpiled file than the split is worth, so the file is mostly such members.
- A transpile, build, or test failure survives applying the roster's documented rewrite for that construct, which is evidence the roster is out of date for this compiler and not a licence to invent a rewrite.
- The only way to make the file work is a transpiler version other than the pinned one, which re-baselines the corpus and every committed twin and is a separate exercise.

When you stop, report the preflight verdict, the construct families that drove it, and the condition above that applies.
Do not proceed by hand-editing a generated file, by reshaping a Go-facing signature to make the transpiler happy, or by relaxing a test.

## Rewrites That Transpile And Then Fail

Several rewrites transpile cleanly and fail afterwards, at build time, at the Go call boundary, or on the next regeneration, and none of them produces an error at the point of the mistake.
The playbook's traps table is the authority: read it before debugging any failure in a rewritten file, because the answer is usually a documented transpiler behaviour rather than a bug in the twin.

Two of those behaviours are actions to take or avoid rather than diagnoses to make, so they are stated here:

- Never hand-edit a generated twin, because it carries a `DO NOT EDIT` header and the next transpile discards the edit without a trace.
- Never transpile from the repository root, because the `//line` directives the transpiler emits name the path it was given, so a root transpile writes directives that do not resolve from the generated file and then show up as a diff on every regeneration.

## Definition Of Done For One File

The playbook's definition of done for one file is the authoritative checklist.
In short, before reporting the rewrite as finished:

- The preflight verdict is recorded, with the rung and the construct families that drove it.
- The `.gala` source and the generated Go file are both committed side by side, and the generated file was not hand-edited.
- Transpiling twice from the package directory produces no diff, the emitted `//line` directives are relative, and `gofmt -l` on the twin is empty.
- `go build ./...` in `server/` passes, the canonical backend test sequence passes, and `run.sh`, `audit.sh`, and `go run ./inventory -check` all pass.
- Every document and count that lists committed twins has gained its row, and the roster's inventory counts were re-derived by the inventory tool rather than hand-edited.
- Anything the procedure could not answer is recorded as a follow-up rather than worked around silently.

## Working Practices

- If the rewrite is not confined to a single file and its twin, read `BACKLOG_WORKFLOW.md` first and route the work through Backlog as that policy requires.
- A file under `server/` is Go code with callers, not a standalone artifact, so preserve its exported signatures, its Go-facing behaviour, and its wire shapes exactly.
- Prefer a package that already carries a `.gala` source: it has a committed style to copy, a sibling convention to follow, and a regeneration command to extend.
- Record what the procedure failed to say, because a step that was ambiguous, skipped, or worked around is a finding about the playbook, and correcting it in place is part of finishing the rewrite.
