---
id: TASK-0362
title: Keep local tool state inside the checkout under the DSH sandbox
status: Done
assignee: []
created_date: '2026-10-24 07:49'
updated_date: '2026-10-24 08:03'
labels: []
dependencies: []
ordinal: 366300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
DSH's workspace-write sandbox mounts $HOME read-only, so the local tools this repository relies on fail when they write their state there: `nix develop` cannot open `~/.cache/nix/fetcher-locks/*.lock`; `gala build` extracts its stdlib and keeps build workspaces under `~/.gala` and writes its Go work to `~/.cache/go-build`; `npm` fails on `~/.npm/_cacache`; `codebase-memory-mcp` fails to create its private cache under `~/.cache/codebase-memory-mcp`; and a `BACKLOG_CWD` inherited from a shell started in another checkout can point `backlog` at the wrong project. Relocate that state into the checkout the way session a65e4b1d did for that repository's tools: `.envrc` exports XDG_CACHE_HOME, GALA_HOME, npm_config_cache, and CBM_CACHE_DIR under `.local/` before `use flake`; `flake.nix` stops exporting an invocation-dependent BACKLOG_CWD; and `.envrc` pins the correct one after the shell loads. Scope: local development environment only; no runtime, test, or deployment change. Out of scope: relocating the Go module cache (`~/go/pkg/mod`), which is a read-only dependency source here; adding a dependency would still need a writable module cache.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 From a workspace-write sandboxed shell with a read-only $HOME, `nix develop`, `gala transpile`, `gala build`, `go build ./...` in `server/`, and `backlog` operate without writing outside the checkout.
- [x] #2 `.envrc` exports XDG_CACHE_HOME and GALA_HOME under `$PWD/.local/` before `use flake`, and pins BACKLOG_CWD to the checkout after it.
- [x] #3 `flake.nix`'s shellHook no longer exports an invocation-dependent BACKLOG_CWD and clears any value inherited from the launching shell, so backlog resolves the project from its cwd inside `nix develop`.
- [x] #4 `.local/` is ignored, no probe artifacts or unintended untracked files remain, and the verification evidence is recorded in the task notes.
- [x] #5 `.envrc` also relocates npm's cache and codebase-memory-mcp's private cache, so `npm cache add` and `codebase-memory-mcp cli index_repository` succeed in the sandbox.
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
Plan: probe each tool for writes outside the checkout under the sandbox; relocate nix/Go caches with XDG_CACHE_HOME and GALA state with GALA_HOME in .envrc; move BACKLOG_CWD out of shellHook; verify with a read-only HOME and a deliberately stale BACKLOG_CWD; leave direnv allow and the DSH relaunch as the documented manual steps.

Verified under the workspace-write sandbox ($HOME read-only):
- `XDG_CACHE_HOME=$PWD/.local/cache nix develop --command true` exited 0 and created `.local/cache/nix/fetcher-locks`; the same invocation from `server/` resolved the flake.
- In that dev shell, `gala version` reports 0.92.1. With an ambient `BACKLOG_CWD=/tmp`, the shellHook leaves it unset, and `backlog task list` reports the repository's 150 tasks from both the root and `server/`; from `/tmp` it reports "No Backlog.md project found".
- `GALA_HOME=$PWD/.local/gala server/scripts/gala/verify.sh` reported OK (30 twins), and its `go build ./...` passed; `gala build` in a scratch project built with 0.92.1 and extracted the stdlib to `.local/gala/stdlib/v0.92.1`.
- `npm cache add left-pad` fails with EROFS on `~/.npm`; with `npm_config_cache` set it exits 0.
- `codebase-memory-mcp cli index_repository --repo-path .` indexed 9,043 nodes and 28,718 edges with `CBM_CACHE_DIR` under `.local/cache/codebase-memory-mcp`.
- `npm run format:markdown` made no changes and `format:markdown:check` passes; the task file is under the pipeline's `backlog/**` exclusion.
- Temporary `.local/` caches (511 MB) and probe files were removed after verification; the directories are recreated on demand.
Manual follow-up: run `direnv allow` in a normal terminal, then restart `dsh web` from that shell so tool calls inherit the exports; a bare `backlog` outside the dev shell still uses whatever `BACKLOG_CWD` DSH was launched with.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Local tool state now lives inside the checkout: .envrc relocates the nix/Go/npm caches, GALA's home, and codebase-memory-mcp's private cache under .local/, and pins BACKLOG_CWD after use flake, while flake.nix clears an inherited BACKLOG_CWD instead of exporting an invocation-dependent one. Verified in the read-only-home sandbox: nix develop, gala transpile/build, verify.sh (30 twins), go build, backlog, npm cache, and codebase-memory indexing all work; .local/ is ignored. direnv allow and a DSH relaunch remain manual.
<!-- SECTION:FINAL_SUMMARY:END -->
