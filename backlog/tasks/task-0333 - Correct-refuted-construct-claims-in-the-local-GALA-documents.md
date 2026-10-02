---
id: TASK-0333
title: Correct refuted construct claims in the local GALA documents
status: Done
assignee: []
created_date: '2026-09-29 19:24'
updated_date: '2026-09-29 19:30'
labels: []
dependencies: []
modified_files:
  - docs/gala-translation.md
  - docs/gala-rewrite-playbook.md
  - server/GALA.md
  - >-
    server/scripts/20260923_gala_translation_probes/probes/delta_multi_value_define/expect
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
priority: high
type: docs
ordinal: 338000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Correct the statements in this repository's own GALA documents that measurement on the pinned compiler has refuted.

The corrections are derived in TASK-0332 by transpiling a minimal repro for each claim on the pinned compiler and, for a shape claim, reading the emitted Go and then building and running it. The evidence for each correction is recorded in that task's notes and is not repeated here.

## Claims to correct

A multi-value `:=` receive does not lower like `var`. It is a val-style binding: each result is bound through an immutable wrapper and the transpiler inserts the unwrap at every read. The `var` spelling emits a plain Go binding. Three places in the translation roster state the old claim, and the mechanical-rewrite catalog prescribes the `:=` spelling in a GALA column whose own note recommends the `var` spelling.

A GALA struct does not synthesize an `Instance` interface or an `Is<Type>` method. The emitted Go carries three synthesized members. Two roster rows and one spike-record paragraph claim the other set.

A multi-value `:=` receive is described as a transpiler panic in the translation roster's contested section, the spike record's not-working list, and a playbook trap. The panic was fixed in the pinned release, so all three are stale, and one of them is the reason a mechanical rewrite exists.

The `if` initializer is explained by a historical build failure in the spike record; the compiler now rejects it at transpile time with its own diagnostic.

The `GAP-2` revisit trigger asks for a fix to the receive panic that has already landed.

## Scope

Documents and probe notes only.

Extending the probe corpus to pin the corrected claims is deliberately not in scope: a claim about emitted shape has no assertion in the current expectation format, so pinning it means extending the runner, which is a separate change and should be decided on its own merits rather than folded into a correction.

The three upstream transpiler defects found alongside these corrections are reported in TASK-0332's notes and belong to the upstream tracker, not to this repository.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every occurrence of the claim that a multi-value `:=` receive lowers like `var a, b = f()` is corrected to the measured behaviour, and the correction states that the transpiler inserts the unwrap at each read
- [x] #2 Every occurrence of the claim that a GALA struct synthesizes an `Instance` interface or an `Is<Type>` method is removed and replaced with the three members the emitted Go actually contains
- [x] #3 The multi-value receive's stale description as a transpiler panic is removed from every document that repeats it, and each is updated to the release in which the panic was fixed
- [x] #4 The reason given for the `if` initializer restriction matches the diagnostic the compiler currently emits rather than a historical build failure
- [x] #5 The `if`-initializer mechanical-rewrite row's GALA column is consistent with its own note, so the column and the note name the same binding spelling
- [x] #6 The `GAP-2` revisit trigger no longer requests a fix for an already-landed fix
- [x] #7 `audit.sh` passes after the edits, so the roster, the corpus, and the spike record still agree
- [x] #8 `run.sh` passes, so the probe that pins the multi-value `:=` receive still transpiles, builds, and runs
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Scope decision taken at implementation time

Extending the probe corpus to pin the corrected claims is not in scope, and the task description records why: the `expect` format supports `KIND`, `CODE`, `ERR`, `RUN`, and `NOTE` and has no assertion about emitted text, so pinning a shape claim means extending `run.sh`. That is corpus tooling rather than a documentation correction, and folding it in would have made the correction's blast radius unclear.

The probe note is therefore corrected to say what it actually pins: the program's recovered value, not the emitted shape. Previously it claimed the fix "lowers it like `var a, b = f()`", which was both wrong and the reason a reader would believe the shape was pinned.

## What the `Instance` sweep cost

The first pass at this task removed the `Instance` claim from four places on the strength of TASK-0332's finding. A final sweep for the string across the repository, run after those edits, matched a `main.gen.go` the corpus had just written, containing a `PairInstance` interface. The claim is real for a generic struct; only the non-generic case had been tested.

Re-measured side by side on the pinned compiler: `struct P(var A int)` emits `Copy`, `Equal`, `Unapply`; `struct Pair[T any](var First T, var Second T)` emits those three plus `PairInstance` and `IsPair() bool`. All four removals were reverted to a qualified form.

The lesson is recorded in TASK-0332 as well: a single negative result does not refute a conditional claim, and a sweep that reads generated output catches what a source-level reading misses.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Corrected the statements in this repository's GALA documents that measurement on the pinned compiler had refuted, and qualified two that were right but stated without their condition. Tracked as TASK-0333; the measurements are in TASK-0332.

## Corrections

**The multi-value `:=` receive is a val-style binding, not a `var`.** It transpiles, but binds each result through an immutable wrapper and the transpiler inserts the unwrap at every read, while `var` emits a plain Go binding. Corrected in the roster's compiler history, in the `multi-value-define` roster row, and in the roster's multi-value-bindings section, which now frames the rewrite as a shape choice for a Go-visible binding rather than a repair.

**The mechanical-rewrite catalog contradicted its own note.** The `if`-initializer row prescribed `err := f()` in its GALA column while its note recommended `var`; the column now says `var err = f()` and the note explains why the `:=` spelling would wrap the value.

**The multi-value receive is no longer described as a transpiler panic** in the roster's contested section, the spike record's not-working list, or a playbook trap. All three now say the panic was fixed by the upstream pull request and state the val-style binding as the reason the `var` spelling is still the right one. The playbook trap was the most misleading, because it promised a diagnostic that does not come: the `:=` spelling transpiles, builds, and runs, so nothing reports the problem.

**The `if` initializer is explained by its current diagnostic.** The spike record explained it by a historical build failure; it now shows the transpile-time rejection with its own code and hint.

**`GAP-2`'s revisit trigger no longer asks for the already-landed fix** to the receive panic.

**The `Instance` claim was qualified rather than removed.** A repository-wide sweep found a freshly generated probe file containing a `PairInstance` interface, which refuted TASK-0332's finding: the interface and the `Is<T>()` method are synthesized for a generic struct and not for a non-generic one. Four places stated the rule without the generic qualifier and now carry it. This was my error, generalized from one non-generic test case, and TASK-0332's record carries the withdrawal.

## Files

`docs/gala-translation.md`, `docs/gala-rewrite-playbook.md`, `server/GALA.md`, and one probe expectation note.

## Checks

`audit.sh` passes, including the check that the mechanical-rewrite catalog and the no-rewrite index agree with the roster's gap classes in both directions, which is what the catalog-column edit could most plausibly have broken. `run.sh` passes 54 of 54. `npm run format:markdown` was run and re-checked, and the audit was re-run after it reformatted the roster. A structural check confirms every table in the edited documents still has a consistent cell count.

Swagger regeneration and the code-graph refresh are not applicable: no annotations, runtime code, or generated artifact changed. Backend and e2e tests are not required for a documentation-only change and were not run.

## Deliberately not done

Extending the probe corpus to pin the corrected shape claim is not in scope, and the task description says why: the expectation format has no assertion about emitted text, so pinning a shape claim means extending the runner. That is a change to the corpus tooling and deserves its own decision rather than being folded into a correction. The probe note now says explicitly that it pins the program's value and not the emitted shape, so the gap is recorded rather than implied.

The three upstream transpiler defects remain in TASK-0332's notes for the upstream tracker.
<!-- SECTION:FINAL_SUMMARY:END -->
