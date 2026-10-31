---
id: TASK-0345
title: Request that upstream publish the transpiled GALA stdlib as a release asset
status: Done
assignee: []
created_date: '2026-10-04 19:57'
updated_date: '2026-10-04 19:59'
labels: []
dependencies: []
priority: medium
type: task
ordinal: 349005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
server/third_party/gala is a committed flattened copy of the transpiled GALA standard library, re-synced manually on every compiler pin bump. Upstream commits only .gala sources and generates the Go with Bazel/Nix, so a plain-Go build cannot consume the upstream module: `go get github.com/martianoff/gala` fails on the declared module path (`module martianoff/gala`), and a module replace reaches a tree with zero .gen.go files. This task asks upstream to publish the already-built transpiled stdlib as a per-release asset and records the request in the GALA ledger. Reporting/documentation only; no runtime, test, or build changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A feature request is filed on github.com/martianoff/gala asking for a per-release transpiled stdlib artifact that a plain-Go consumer can use with a single `replace` directive.
- [x] #2 The request carries the concrete evidence: upstream has no committed .gen.go files, the direct `go get` module-path failure, and the replace-resolves-but-build-fails result.
- [x] #3 server/GALA.md's Upstream report index gains a row for the filed issue, and the Compiler and runtime section names the request as the path off the committed vendored tree.
- [x] #4 No runtime code, tests, or build configuration change.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Plan and evidence

1. Search active Backlog tasks and the upstream martianoff/gala issue/PR lists to confirm no existing request covers publishing the generated stdlib. Closest is #530 (Nix-vs-Bazel parity, closed); no packaging request exists.
2. Verify the friction on the pinned rev 9beea5dc: upstream module has 0 `*.gen.go` files while `server/third_party/gala` has 53; `go get github.com/martianoff/gala@9beea5dc...` fails with "module declares its path as: martianoff/gala"; `go mod edit -replace=martianoff/gala=github.com/martianoff/gala@...` resolves and hash-pins in go.sum, but `go build ./...` reports 10 undefined symbols in `std`.
3. Confirm the extraction shape: `~/.gala/stdlib/v0.85.0/` is one Go module per package with relative `replace`s, so a single main-module replace cannot compose it; a flattened root module is required.
4. File the feature request with the template's sections and concrete commands/errors.
5. Record it in `server/GALA.md`: one sentence in the Compiler and runtime section and one Upstream report index row.

## Result

- Issue: https://github.com/martianoff/gala/issues/698 (open, label `enhancement`)
- `server/GALA.md`: Compiler and runtime section names #698 as the path off the committed vendored tree; report index row added.
- Validation: `npm run format:markdown` (no further edits) and `npm run format:markdown:check` (exit 0). No unit/e2e coverage applies to a reporting-only change.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Filed upstream feature request https://github.com/martianoff/gala/issues/698 ("[Feature]: publish the transpiled standard library as a release asset for Go consumers", label enhancement). The request asks for a per-release `gala-stdlib-v<version>.tar.gz` plus checksum shaped as one flattened Go module (`module martianoff/gala`, per-package `.gen.go` + handwritten `.go` helpers, no nested `go.mod`, LICENSE) so a plain-Go consumer can use a single `replace` with no GALA or Nix in the build. Evidence in the body: the 0.85.0 tree has zero `*.gen.go` vs 53 in the committed copy; `go get github.com/martianoff/gala` fails on the declared module path; a module `replace` resolves but `go build` reports 10 undefined errors in `std`; the CLI extraction's per-package modules cannot compose under one main-module replace. `server/GALA.md` now records the request in the Compiler and runtime section and the Upstream report index (#698, open). Reporting-only change; no runtime, test, or build configuration touched. Validation: `npm run format:markdown` produced no further edits and `npm run format:markdown:check` exits 0.
<!-- SECTION:FINAL_SUMMARY:END -->
