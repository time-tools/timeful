# Repository Layout

This repo contains:

- a frontend in `./frontend`
- a backend in `./server`

## Agent Notes

- Store durable agent memory and handoff notes in the repository, not a home-directory auto-memory location.

## Session Handoffs

- Session handoffs are an append-only custom archive in `backlog/handoffs/`, not Backlog-managed task or document records.
- Create handoff templates with `scripts/handoff/create-handoff.sh` or the `/handoff` skill.
  Do not overwrite or edit older archive entries.
- A superseded entry may be removed only when a newer reference preserves its durable facts, the removal is explicitly approved and recorded there, and retained handoffs contain no links to it.
  Keep entries that provide irreplaceable external validation evidence or unresolved operational context.

## Backlog Workflow

<CRITICAL_INSTRUCTION>

Before taking non-trivial implementation action requested by the user, read `BACKLOG_WORKFLOW.md` and use it to decide whether the work requires a Backlog task.
Questions, exploration, and obvious mechanical changes do not require this review.
`BACKLOG_WORKFLOW.md` is this repository's authoritative Backlog policy.

</CRITICAL_INSTRUCTION>

## Documentation

- Canonical product requirements are in `docs/requirements/`; read `docs/requirements/AGENTS.md` and `docs/requirements/README.md` before creating, changing, or migrating a requirement.
- Markdown authoring rules, including the sentence-per-line and table-row rules, live in `docs/AGENTS.md`.
- Controlled terminology and linking rules live in `docs/terminology/README.md`; follow them when prose refers to a glossary concept.

## Working Defaults

Unless the user explicitly asks for server changes:

- treat `frontend/` as the primary working directory
- keep repo-tracked frontend browser checks under `e2e`

## Bug Fix Protocol

For any bug fix, regardless of layer:

- First add or update the regression check and observe it fail for the reported reason before changing the implementation.
- After the fix, run the identical check and confirm it passes.
- Prefer the narrowest layer that can capture the behavior: a unit test, then a browser E2E spec, then a scripted or documented manual reproduction.
- If no test layer can capture the behavior, record the manual reproduction steps and why automated coverage is not practical.
- Record the fail-before and pass-after evidence in the Backlog task notes and final summary.

The protocol is scoped to bug fixes and does not require a failing check before feature work, refactoring, or documentation-only changes.

## Backend Conventions

- The Go module path is `timeful/server`; use that prefix for internal imports.
- Keep PostgreSQL access in `server/postgres/`; route handlers and services should not access PostgreSQL directly.
- Put one-off PostgreSQL data migrations in dated `server/scripts/YYYYMMDD_description/` directories.
  Run them manually; do not import them into runtime code.
- Add Swag annotations to API handlers.
  When route annotations change, regenerate Swagger and the generated API types with the commands in `server/routes/README.md`.
- Do not change browser-plugin `window.postMessage` payload shapes without also updating `PLUGIN_API_README.md`.
- Backend tests use the isolated Compose test stack; the canonical commands live in `server/README.md`, and the isolation semantics live in the test isolation section of `docs/environments.md`.

## Frontend

- Follow the frontend ADRs and `frontend/AGENTS.md`; they own the frontend architecture, styling, Vuetify migration, browser verification, and required-check rules.

## Formatting

- Code formatting uses oxfmt (configs in `.oxfmtrc.json` and `frontend/.oxfmtrc.json`); root Markdown stays formatted by the Prettier sentences-per-line pipeline and must not be formatted by oxfmt.
- For changes to root JS files under `scripts/` or `prettier/`, run `npm run fmt:check` from the repo root.

## Local Development

- Local frontend debugging follows `frontend/README.md`; the canonical env-file contract lives in `docs/environments.md`.
- Browser E2E follows `e2e/AGENTS.md`; it owns project selection, commands, isolated-stack rules, and failure diagnosis.

## Rewrite Safety

- when cleaning the worktree for rebases, amends, or other history rewrites, prefer explicitly moving or copying tracked and untracked files aside and restoring them afterward
- do not use `rm` as the primary cleanup mechanism when a non-destructive move or backup approach is practical

## MCP Usage

- VS Code MCP: only use `search_symbols_code`, `get_symbol_definition_code`, and `get_diagnostics_code`.
- codebase-memory-mcp: follow `docs/codebase-memory.md` before grepping or reading files for codebase questions.
