---
id: TASK-0338
title: Re-derive the import-resolution rows that GALA 0.84.1 contradicts
status: To Do
assignee:
  - Danila Danko
created_date: '2026-09-30 17:16'
labels: []
dependencies: []
references:
  - >-
    backlog/tasks/task-0337.01 -
    Repair-the-local-GALA-documents-left-inconsistent-by-the-TASK-0337-filing-pass.md
  - 'https://github.com/martianoff/gala/blob/master/docs/GALA.MD'
documentation:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/GALA.md
  - docs/gala-translation.md
  - server/scripts/20260923_gala_translation_probes/run.sh
  - server/scripts/20260923_gala_translation_probes/audit.sh
priority: medium
type: docs
ordinal: 343200
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Two documents under `.agents/skills/gala-from-go/` claim a GALA failure mode that `gala` 0.84.1 does not have, and the claim is the kind that costs an afternoon if a reader believes it.

The claim is that a qualified name resolves against the imports of *every* file in the package, so a `.gala` file that omits an import its sibling already has transpiles cleanly and then fails `go build` with `undefined: <pkg>`, blamed on the `.gala` file through its `//line` directive.
On 0.84.1 the transpile is refused instead, with a hint that names the package which declares the name, whether the sibling that imports it is a `.gala` file or a handwritten `.go` file.

This was found while TASK-0337.01 was writing a verified import bullet for `server/GALA.md`, and it is recorded there rather than fixed, because it predates TASK-0337 and is outside that task's criteria.

## What was reproduced, on 0.84.1 / go1.26.7

A package whose sibling imports `strings`, and a second file that calls `strings.ToLower` without importing it:

```
error[GALA-E0023]: undefined: strings
  --> b.gala:3:31
    = hint: check the spelling, add the import that introduces this name, or
      declare it — every identifier must resolve to a binding, a declaration in
      this package, or an imported symbol
```

The same diagnostic appears when the importing sibling is a handwritten `helper.go` instead of a `.gala` file, so the sibling's language is not what decides it.

An unqualified GALA-runtime name is the other shape: `Future(2)` in a file that does not import `martianoff/gala/concurrent` is refused with `GALA-E0023` and a hint that names that package and suggests `import . "martianoff/gala/concurrent"`.
`gala explain GALA-E0025` documents the same per-file rule for the unqualified case, gives a two-file repro, and says outright that sibling files' imports do not propagate.
`server/GALA.md` already quotes an `GALA-E0025` diagnostic from a concurrency probe, so the two documents disagree about the same rule.

## What is wrong where

- `SKILL.md`, the mixed-package bullet: "A qualified name resolves against the imports of *any* file in the package ... When no file in the package imports that package at all, the transpile is refused with `GALA-E0023` instead, so the diagnostic itself tells you which of the two cases you are in." The second branch is the only one that exists on 0.84.1.
- `SKILL.md`, the trap "A sibling file's import hides a missing import from the transpiler": the same claim, stated as a trap, so a reader is told to watch for a build failure the transpiler prevents.
- `constructs.md`, the "Packages And Files" row "an import the file does not declare": `Status` `direct`, `Code` `—`, and a `Check` that reads "omit it from one file of a pair, read the generated import block, then build".
- `docs/gala-translation.md`, the roster's `named-imports` row: `—` and "Works", with no per-file caveat; TASK-0337.01 added the caveat to `server/GALA.md` and named it as the limit that row does not carry, so the two documents do not agree until this is done.

## Constraints

- `gala` 0.84.1 and the go1.26 series stay the pinned toolchains, and no probe expectation is re-baselined beyond what this change moves.
- `server/third_party/gala/` is not modified.
- Filing is out of scope: a documentation defect about the transpiler's own diagnostic is not a new issue, and no existing issue is touched.
- The change follows `docs/AGENTS.md`: one sentence per line, and the Prettier sentences-per-line pipeline rather than oxfmt.
- The new corpus probe and every document edit must leave `run.sh` and `audit.sh` green, including `audit.sh`'s seventh check, which compares every Markdown table row in the five GALA documents against its own header.

## Where to look

The three rows above, the probe corpus under `server/scripts/20260923_gala_translation_probes/`, and `gala explain GALA-E0023` and `gala explain GALA-E0025`, which between them state the rule the documents contradict.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `SKILL.md` states what 0.84.1 does when a `.gala` file omits an import a sibling has, and no sentence claims a qualified name resolves against another file's imports
- [ ] #2 `constructs.md`'s two import rows carry the Status and Code the pinned compiler produces, and each row's Check re-derives that behaviour rather than reading the generated import block
- [ ] #3 A corpus probe pins the refusal, so neither claim can drift back without a failing expectation
- [ ] #4 `server/GALA.md` and `docs/gala-translation.md` say the same thing about import resolution, including the per-file check the roster's `named-imports` row does not carry
- [ ] #5 `run.sh`, `audit.sh` with all seven checks, `go run ./inventory -check`, the inventory unit tests, `npm run format:markdown:check`, `npm run test:markdown-rules` and root `npm run fmt:check` all pass
- [ ] #6 No upstream issue is filed or edited, and `server/third_party/gala/` is unchanged
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
