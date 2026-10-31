---
id: TASK-0366
title: Keep the Go module cache inside the checkout under the DSH sandbox
status: Done
assignee: []
created_date: '2026-10-25 11:47'
updated_date: '2026-10-25 11:50'
labels:
  - tooling
dependencies: []
references:
  - .envrc
  - >-
    backlog/tasks/task-0362 -
    Keep-local-tool-state-inside-the-checkout-under-the-DSH-sandbox.md
priority: medium
type: chore
ordinal: 370300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0362 relocated nix, the Go build cache, npm, GALA, and codebase-memory-mcp state into `.local/`, but deliberately left the Go module cache in `~/go` on the grounds that it was a read-only dependency source. Bumping `go.gala.fyi/stdlib` to a version absent from that cache fails with `read-only file system` on `~/go/pkg/mod/cache/download`, and requesting elevated access was rejected, so the dependency bump cannot complete under the sandbox. Move `GOPATH` (module cache and checksum database) into the checkout the same way the other tools were moved, and verify that a new module version downloads and builds while `$HOME` stays read-only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 .envrc exports GOPATH=$PWD/.local/go and GOMODCACHE=$GOPATH/pkg/mod before use flake.
- [x] #2 A Go command that needs a module version absent from ~/go/pkg/mod succeeds with the checkout cache, and no write reaches ~/go (evidence: `go get go.gala.fyi/stdlib@v0.93.0` and a clean `go build ./...`).
- [x] #3 GOPATH relocation does not affect the Compose backend suite, whose containers keep their own module cache volume.
- [x] #4 .local/ stays gitignored and the `direnv allow` step for the edited .envrc is recorded.
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Reproduce the read-only failure while fetching go.gala.fyi/stdlib v0.93.0.
2. Export GOPATH and GOMODCACHE under $PWD/.local in .envrc before use flake.
3. Fetch the new module and run go build with the relocated cache; confirm ~/go stays read-only.
4. Re-run the canonical Compose backend sequence and confirm the containers keep using their own module-cache volume.
5. Refresh the code graph, check formatting, record the direnv allow step, and commit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Failure reproduced: `nix develop -c go get go.gala.fyi/stdlib@v0.93.0` under the default cache reported `open /home/…/go/pkg/mod/cache/download/go.gala.fyi/stdlib/@v/v0.93.0.info….tmp: read-only file system` and the v0.93.0 lock file was unwritable; a `touch` probe on `~/go/pkg/mod` also failed.

Tried seeding a checkout cache with `cp -al ~/go/pkg/mod .local/go/pkg/mod`; the workspace and home cache are different mounts, so every hardlink failed with `Invalid cross-device link`, and the partially created read-only directory tree was cleaned up (directories only; no regular files were linked).

Final setup: empty `.local/go/pkg/mod`, with `GOPATH=$PWD/.local/go` and `GOMODCACHE=$GOPATH/pkg/mod` exported for the bump commands. `go get` downloaded only the modules it needed, then `go build ./...`, `verify.sh --write` (go build phase), and the Compose suite passed.

`direnv exec` reports `.envrc` blocked until `direnv allow`, which is expected after editing it; the current session's shell keeps the earlier environment, so the bump commands passed GOPATH/GOMODCACHE explicitly.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added `GOPATH="$PWD/.local/go"` and `GOMODCACHE="$GOPATH/pkg/mod"` to `.envrc`, before `use flake`, next to the nix, Go-build, npm, GALA, and codebase-memory relocations from TASK-0362.

Why: bumping `go.gala.fyi/stdlib` to v0.93.0 needs a module version absent from `~/go/pkg/mod`, and the DSH workspace-write sandbox mounts `$HOME` read-only. The first `go get` failed writing `~/go/pkg/mod/cache/download/...` and `~/go/pkg/sumdb/...`; escalating to danger-full-access was rejected. With `GOPATH`/`GOMODCACHE` pointing into the checkout, `go get go.gala.fyi/stdlib@v0.93.0` downloaded the module and updated `go.mod`/`go.sum`, `go mod tidy` pruned the old sums, and `go build ./...` passed. A `touch ~/go/pkg/mod/.dsh-write-test` still fails, so nothing wrote through to the home cache.

The canonical Compose backend sequence is unaffected because its containers mount their own `timeful-test-go-mod-cache` volume at `/go/pkg/mod`. `npm run format:markdown:check` passed and `codebase-memory-mcp cli index_repository --repo-path .` refreshed the graph (9102 nodes / 29467 edges). `.local/` is already gitignored.

Manual step: editing `.envrc` makes direnv block it until `direnv allow` is run in the checkout; the current shell must be re-entered for the exports to take effect.
<!-- SECTION:FINAL_SUMMARY:END -->
