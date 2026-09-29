---
id: TASK-0334
title: Add an emitted-Go expectation kind to the GALA probe corpus
status: To Do
assignee: []
created_date: '2026-09-29 19:46'
labels: []
dependencies: []
priority: medium
type: task
ordinal: 339000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Give the GALA probe corpus an expectation kind that asserts on the text of the generated Go, so a claim about emitted shape can be checked rather than asserted in prose.

## Why

The corpus has three expectation kinds, and none of them can observe the shape of the emitted Go. `KIND=pass` with `RUN=yes` asserts that the program builds and prints the expected output, which cannot distinguish a plain Go binding from a wrapped one, because a wrapper the program never observes is invisible to an output assertion. `KIND=transpile_fail` and `KIND=build_fail` assert that something is rejected, and a rejection is not the claim: a construct can be accepted with the wrong shape.

The consequence is already paid for. A wrong claim about the multi-value `:=` receive survived in three documents and a 54-probe corpus, because the probe that covers it pins the program's recovered value and not the emitted shape. The probe's note now says so, which records the gap without closing it.

## What an emitted-text assertion buys

It closes claims of three kinds, each of which is currently unchecked:

- A binding that wraps versus one that does not, checked by asserting a wrapper marker is present in one spelling and absent in the other. The two spellings are the same program, so the absence assertion is the one that carries the claim.
- A declaration that gains synthesized members versus one that does not, checked across a generic and a non-generic declaration of the same shape.
- A field that is wrapped versus one that is plain, checked by asserting the wrapper's type name in the generated field list.

## Design constraints to resolve before implementing

The runner reads at most the first value of each `KIND`, `CODE`, `ERR`, `RUN`, and `NOTE` key, so a single-valued new key gives one assertion per probe. A decision is needed on whether that is enough, or whether multiple assertions per probe are wanted and the reader has to change. One probe per assertion is the cheaper option and keeps the reader unchanged.

An absence assertion is the interesting half, and it is the half that interacts with the audit. The audit requires a gap's blocking probe not to be `KIND=pass`, on the reasoning that a blocker has to fail for the construct it names. An absence assertion fits that reasoning, because a construct is blocked exactly when the thing it needs is not emitted; a presence assertion does not fit it. Whatever kind is chosen, the audit's check has to classify it correctly rather than inherit the pass rule.

Assertions must target semantic markers rather than layout, because the generated Go is formatted and carries source-map directives that name the probe's own path. A marker like a wrapper constructor call or a synthesized interface name is stable across formatting; an indentation or a line break is not.

These assertions are coupled to the compiler release more tightly than any existing kind, because they pin codegen rather than behaviour. The `build_fail` expectations are already coupled to the Go toolchain, and the roster documents re-baselining for that; a new kind needs the same treatment stated up front, and the corpus page's re-baseline guidance has to cover it. A probe that fails on a release bump for a reason nobody recorded is worse than no probe.

## What this does not cover

The mixed-package defect, where a method call is emitted on a field whose type the transpiler does not know, is already covered: the generated Go does not compile, so `KIND=build_fail` with the compiler's diagnostic is the right kind for it. An emitted-text assertion adds nothing there and should not be used for it.

The upstream transpiler defects found in TASK-0332 belong to the upstream tracker and are not in scope.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `run.sh` supports an expectation that asserts on the text of the generated Go, with both a presence assertion and an absence assertion, and a probe using it passes when the generated text matches and fails with a readable diff when it does not
- [ ] #2 `audit.sh` accepts the new expectation kind rather than rejecting it as an unknown kind
- [ ] #3 At least one probe pins that a multi-value `:=` receive wraps its bindings and one pins that the `var` spelling does not, so the two spellings can no longer be documented as equivalent
- [ ] #4 At least one probe pins that a generic struct declaration gains an `Instance` interface and an `Is<T>()` method while a non-generic declaration does not, so the qualified claim is checked in both directions
- [ ] #5 The roster's notes state that these assertions are coupled to the compiler release by construction, that a compiler bump is expected to fail them, and what the re-baseline procedure is
- [ ] #6 The runner's documentation and the corpus page state which of the existing probe kinds assert on program behaviour, which assert on a build failure, and which assert on emitted text, so a future author knows which kind a new claim needs
- [ ] #7 `audit.sh` and `run.sh` both pass, and the full probe count is reported
- [ ] #8 Any probe that is added rather than extended uses a name that does not collide with an existing probe, and the roster's probe links resolve to it
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
