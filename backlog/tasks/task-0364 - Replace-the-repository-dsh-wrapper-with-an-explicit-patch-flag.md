---
id: TASK-0364
title: Replace the repository dsh wrapper with an explicit --patch flag
status: Done
assignee: []
created_date: '2026-10-25 11:08'
updated_date: '2026-10-25 11:08'
labels: []
dependencies: []
references:
  - .dsh/cordis.patch.yml
  - .dsh/README.md
  - AGENTS.md
  - docs/codebase-memory.md
modified_files:
  - .dsh/cordis.patch.yml
  - .dsh/README.md
  - AGENTS.md
  - docs/codebase-memory.md
priority: medium
type: chore
ordinal: 368300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The repository-local `dsh` overlay `.dsh/cordis.patch.yml` registers the Backlog and codebase-memory MCP servers, but DSH does not discover project-local patch files: only bundle patches, the profile patch, the home patch, and explicit `--patch` overlays are applied. Remove the `.dsh/bin/dsh` wrapper and its `.envrc` `PATH` export, and make the overlay opt-in by passing `--patch .dsh/cordis.patch.yml` on the launcher command line. Document the canonical `dsh web --patch .dsh/cordis.patch.yml` form, argument ordering, restart-on-edit behavior, and the `dsh plugin` / `--dump-default-config` exceptions in the overlay header, `.dsh/README.md`, `AGENTS.md`, and `docs/codebase-memory.md`. This is a development-tooling and documentation change: no runtime code, tests, generated artifacts, build configuration, or deployment configuration changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The `.dsh/bin/dsh` wrapper and the `.envrc` `PATH` export are removed, `dsh` resolves to the installed launcher, and no tracked file references `.dsh/bin`.
- [x] #2 `dsh web --patch .dsh/cordis.patch.yml --dump-config` composes the `mcp-backlog` and `mcp-codebase-memory` rows, while `dsh web --dump-config` composes neither; the `dsh headless --patch ...` form also parses.
- [x] #3 The overlay header, `.dsh/README.md`, `AGENTS.md`, and `docs/codebase-memory.md` state the explicit `--patch` invocation, its ordering before app arguments, the restart-after-edit rule, and the `dsh plugin` / `--dump-default-config` exceptions.
- [x] #4 Changed Markdown passes the repository sentences-per-line formatting check, and the verification evidence is recorded in the task notes and final summary.
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
## Plan

1. Confirm DSH's patch layering from the installed implementation: bundle layers, the profile patch, the home patch, and explicit `--patch` overlays only.
2. Delete the untracked `.dsh/bin/dsh` wrapper and restore `.envrc` to its committed content so no `PATH` injection remains.
3. Rewrite the overlay header, `.dsh/README.md`, `AGENTS.md`, and `docs/codebase-memory.md` around the explicit `--patch` invocation.
4. Verify with `--dump-config` that the flag composes the MCP rows and its absence does not, then run the Markdown formatting check.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-10-25)

Removed `.dsh/bin/dsh` and the `.envrc` block that put `.dsh/bin` first on `PATH`; `.envrc` now matches its committed content. `command -v dsh` resolves to `~/.nix-profile/bin/dsh`, which reports 0.2.1-rc.3.

Rewrote the header of `.dsh/cordis.patch.yml`, `.dsh/README.md`, the DSH MCP row in `AGENTS.md`, and the dsh sentence in `docs/codebase-memory.md` to describe the explicit `--patch` invocation, flag ordering, restart semantics, and the `dsh plugin` / `--dump-default-config` exceptions.

## Verification evidence

- Auto-discovery negative case: with a throwaway `DSH_HOME`, `dsh web --dump-config` composes zero `mcp-backlog` rows even though `.dsh/cordis.patch.yml` exists in the checkout.
- Documented command: `dsh web --patch .dsh/cordis.patch.yml --dump-config` composes one `mcp-backlog` row and one `mcp-codebase-memory` row; `dsh headless --patch .dsh/cordis.patch.yml --dump-config` also composes the row, so the profile-then-flag form parses for both apps.
- `bash -n .envrc` passes.
- `npm run format:markdown:check` exits 0 for tracked files; the untracked `.dsh/README.md` was checked with the same prettier plus sentences-per-line config and is already formatted.

## Checks

- Documentation-only and development-tooling change: unit tests, e2e tests, Swagger regeneration, the code index, and `fmt:check` do not apply.
- No tracked file references `.dsh/bin`.
- Temporary `DSH_HOME` probe directories and dumps under `.local/` were removed after verification.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The repository dsh overlay is now applied only by an explicit launcher flag; the `.dsh/bin/dsh` wrapper and its `.envrc` `PATH` export are gone.

## Verification

- Removed `.dsh/bin/dsh` (and the empty `bin/` directory) and restored `.envrc` to its committed content; `command -v dsh` now resolves to `~/.nix-profile/bin/dsh`.
- With a throwaway `DSH_HOME`, `dsh web --patch .dsh/cordis.patch.yml --dump-config` composes one `mcp-backlog` row and one `mcp-codebase-memory` row; `dsh web --dump-config` composes neither, and `dsh headless --patch .dsh/cordis.patch.yml --dump-config` also composes the row, so the profile-then-flag form parses.
- `bash -n .envrc` passes; `npm run format:markdown:check` exits 0; the untracked `.dsh/README.md` passes the same sentences-per-line pipeline when checked directly.
- Documentation-only and development-tooling change: unit and e2e tests, Swagger regeneration, the code index, and `fmt:check` do not apply.
<!-- SECTION:FINAL_SUMMARY:END -->
