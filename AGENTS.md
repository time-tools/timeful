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
Use Backlog MCP tools for managed task, milestone, document, and Definition of Done records.
Do not edit their generated Markdown files directly.

</CRITICAL_INSTRUCTION>

## Requirements Documentation

- Canonical product requirements are in `docs/requirements/`.
- Before creating, changing, or migrating a requirement, read
  `docs/requirements/AGENTS.md` and `docs/requirements/README.md`.

## Documentation Authoring

- Write each Markdown sentence on one physical source line, and never split a sentence across lines.
- Keep each Markdown table row on one physical source line, even when its cells contain multiple sentences.
- When prose refers to a concept in `docs/terminology/glossary.md`, use the term's canonical form exactly as recorded there.
  Read `docs/terminology/README.md` for the full canonicalization and linking rules.

## Working Defaults

Unless the user explicitly asks for server changes:

- treat `frontend/` as the primary working directory
- use clean layout-based fixes, not hacks
- prefer adding regression tests before fixing frontend bugs
- keep repo-tracked frontend browser checks under `e2e`
- newly added regression tests may fail when they are meant to expose an existing bug

## Server Test Workflow

For backend work that touches PostgreSQL-backed route tests:

- use the isolated test overlay in `compose.test.yaml` as the default path
- create the external Go cache volumes first (`docker volume create timeful-test-go-build-cache timeful-test-go-mod-cache`); compose never creates external volumes, `docker volume create` is idempotent, and `docker volume rm timeful-test-go-build-cache timeful-test-go-mod-cache` resets them
- start test PostgreSQL with `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test`
- run the backend test suite, including the account repository packages, with `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test`
- retain test state by default; remove it only with `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml down -v`
- prefer the isolated Compose stack over host PostgreSQL for repeatable local and CI-friendly runs
- if PostgreSQL-backed tests are run directly on the host, require an explicit `POSTGRES_APPLICATION_URI`; the database must be `timeful-test` or have a `timeful-test-` prefix

## Backend Conventions

- The Go module path is `timeful/server`; use that prefix for internal imports.
- Keep PostgreSQL access in `server/postgres/`; route handlers and services should not access PostgreSQL directly.
- Put one-off PostgreSQL data migrations in dated `server/scripts/YYYYMMDD_description/` directories.
  Run them manually; do not import them into runtime code.
- Add Swag annotations to API handlers.
  When route annotations change, from `server/` run `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --parseDependency`, then from `frontend/` run `npm run gen:api`.
- Do not change browser-plugin `window.postMessage` payload shapes without also updating `PLUGIN_API_README.md`.

## Cross-Cutting Frontend Rules

Follow the frontend ADRs and `./frontend/AGENTS.md` for implementation details.
In particular:

- keep boundary and transport types separate from internal types
- preserve one canonical internal shape per concept
- keep compatibility coercions and transport decoding or encoding at explicit boundaries, not inside views, composables, or submit paths
- many Temporal regressions are runtime issues, so passing typecheck or build is not sufficient
- keep shared timezone decoding centralized and avoid rebuilding it at call sites
- treat Temporal values with value semantics, not identity semantics
- keep civil-date, end-of-day, and working-hours semantics explicit at domain boundaries

## Required Checks

For frontend work, run:

- `cd frontend && npm run lint`
- `cd frontend && npm run fmt:check`
- `cd frontend && npm run typecheck`
- `cd frontend && npm run build`
- `cd frontend && npm run test:unit`

For changes to root JS files under `scripts/` or `prettier/`, run `npm run fmt:check` from the repo root.
Code formatting uses oxfmt (configs in `.oxfmtrc.json` and `frontend/.oxfmtrc.json`); root Markdown stays formatted by the Prettier sentences-per-line pipeline and must not be formatted by oxfmt.

## Local Frontend Debug

For local frontend debugging, keep the production-oriented `compose.yaml` and layer the repo-local override on top of it.

Local frontend tooling expects these variables in the repo-root `.env.development`:

- `VITE_DEV_HOST`
- `VITE_DEV_PORT`
- `VITE_API_PROXY_TARGET`

If your local backend is on a different host or port, point `VITE_API_PROXY_TARGET` there instead.

Useful local entry points:

- fast UI debug: `http://127.0.0.1:4173/test`
- real integrated flow: sign in, open `http://127.0.0.1:4173/home`, then click create event

The Vite dev server proxies `/api` and `/swagger` to `VITE_API_PROXY_TARGET`, so frontend requests stay same-origin and avoid browser CORS issues.
The canonical env-file contract lives in `docs/environments.md`.

## Local Firefox E2E Verification

Browser E2E always uses the isolated test stack and must never target either development database:

- run Playwright from `e2e/` with `npm run test:e2e -- --project=firefox-desktop`; it starts `postgres-test` and `server-test` on `3003`, then Vite on `4174`
- run E2E so its full output streams: never pipe a run through `tail` or `head`; when a persistent full log is needed, append `2>&1 | tee /tmp/opencode/<name>.log`
- `TEST_DB_PERSIST` defaults to `false`, removing the test stack and database volumes; set it to `true` to retain database state after successful or failed E2E setup
- Playwright owns the isolated test stack and Vite process; do not use an existing server for browser E2E.
- the test stack keeps persistent Go caches in the external `timeful-test-go-build-cache` and `timeful-test-go-mod-cache` volumes, so `go run .` inside `server-test` compiles incrementally across runs; `down -v` retains them, and `docker volume rm timeful-test-go-build-cache timeful-test-go-mod-cache` resets them

## Rewrite Safety

- when cleaning the worktree for rebases, amends, or other history rewrites, prefer explicitly moving or copying tracked and untracked files aside and restoring them afterward
- do not use `rm` as the primary cleanup mechanism when a non-destructive move or backup approach is practical

## VS Code MCP Usage

Only use:

- `search_symbols_code`
- `get_symbol_definition_code`
- `get_diagnostics_code`

## codebase-memory-mcp

This project's code knowledge graph is served by codebase-memory-mcp over MCP, and the devShell provides the `codebase-memory-mcp` binary from the pinned nixpkgs flake input.

Rules:

- For codebase questions, use the codebase-memory-mcp MCP tools (`search_graph`, `trace_path`, `get_architecture`, `query_graph`, and the rest) before grepping or reading files.
  They return scoped graph evidence that is much smaller than raw source browsing.
- When the graph is missing or stale, refresh it with `codebase-memory-mcp cli index_repository --repo-path .`.
  The project Definition of Done requires this command after code changes.
- `.cbmignore` at the repository root keeps excluded paths, including inactive ADRs, out of the index across re-indexing.
- If the MCP server is unavailable, confirm the session shell came from `nix develop` and that `codebase-memory-mcp --version` resolves.
