# Environment Files

Timeful uses one root app env file per environment:

- `.env.development` for local development
- `.env.test` for isolated browser and database-backed tests
- `.env.staging` for staging deployments and staging-style runs
- `.env.production` for production builds and production-style runs

Caddy uses a separate root edge env file:

- `.env.edge` for staging and production hostnames

The shared observability instance uses its own root env file:

- `.env.observability` for the shared staging and production OpenObserve instance

The application-level deployment environment is defined separately from toolchain mode:

- backend runtime: `APP_ENV`
- frontend build-time/browser boundary: `VITE_APP_ENV`

Allowed values for both are `development`, `test`, `staging`, and `production`.
When unset, blank, or invalid, both sides default to `development`.
Normal deployments should keep `APP_ENV` and `VITE_APP_ENV` aligned.
`staging` is preserved as a distinct label, but it uses production-like defaults unless overridden explicitly.

Shareable defaults live in:

- `.env.development.example`
- `.env.test.example`
- `.env.staging.example`
- `.env.production.example`
- `.env.edge.example`
- `.env.observability.example`

## How the env files are used

- Frontend dev tooling reads `.env.development` through `frontend/config/tooling.ts`.
- Frontend browser tests read `.env.test` through `e2e/config/tooling.ts` and run Vite in test mode.
- Frontend staging-style builds read `.env.staging`.
- Frontend production builds and `vite preview` read `.env.production`.
- Vite client env loading uses the repo root as `envDir`, so `import.meta.env.VITE_*` also comes from the same root file for the active mode.
- Docker Compose reads the selected root env file through `--env-file`.
- `frontend-artifacts` receives frontend build-time values from that same Compose env file.
- `server` receives backend runtime variables from Compose interpolation based on that same file.
- The edge Caddy Compose project reads `.env.edge`; it passes only its `CADDY_*` values into Caddy.
- The observability Compose project reads `.env.observability`; it passes only its `OPENOBSERVE_*` values into the shared OpenObserve instance.
- The Go server runs through Docker Compose.
  Compose injects its complete runtime environment; direct `go run` is unsupported.
- `SESSION_SECRET` is required and must contain at least 32 characters.
- The server uses `FRONTEND_DIST` when set.
  Otherwise, it looks for frontend artifacts at `./frontend/dist`, then `../frontend/dist`.

## Variable ownership

Frontend tooling variables:

- `VITE_DEV_HOST`
- `VITE_DEV_PORT`
- `VITE_API_PROXY_TARGET`
- `VITE_PREVIEW_HOST`
- `VITE_PREVIEW_PORT`

Isolated browser E2E network variables:

- `E2E_VITE_HOST`
- `E2E_VITE_PORT`
- `E2E_API_HOST`
- `E2E_API_PORT`
- `E2E_API_INTERNAL_PORT`

Frontend build-time variables:

- `VITE_APP_ENV`
- `VITE_POSTHOG_API_KEY`
- `VITE_POSTHOG_API_HOST`
- `VITE_ENABLE_SIGN_IN`
- `VITE_ENABLE_RICH_LANDING`
- `VITE_FEEDBACK_URL`
- `VITE_SUPPORT_EMAIL`
- `VITE_GITHUB_REPO_URL`

Compose-to-frontend build arg mappings:

- `CLIENT_ID` -> `VITE_GOOGLE_CLIENT_ID`
- `MICROSOFT_CLIENT_ID` -> `VITE_MICROSOFT_CLIENT_ID`
- the `VITE_*` build-time flags above are passed through directly

## Frontend build-time flag semantics

- **`VITE_ENABLE_SIGN_IN`** — Controls sign-in and sign-up availability in the frontend.
  Defaults to `true` when unset or blank.
  Set to `false` to hide sign-in buttons, redirect
  sign-in/sign-up routes away, and replace sign-in-gated feature prompts with
  "Requires sign-in, which is disabled in this build." Existing auth sessions still
  work, so previously signed-in users retain access to auth-protected routes.
  This is a frontend-only gate; backend auth endpoints remain live regardless.
- **`VITE_ENABLE_RICH_LANDING`** — Controls whether the full landing page is shown.
  Defaults to `true` when unset or blank.
  Set to `false` to keep only the Timeful
  brand, the header "How it works" action, the GitHub icon, the "Find a time to meet"
  heading, the primary create-event CTA, and the hero preview card.
  This minimal mode
  hides landing sign-in affordances, the in-page how-it-works section, testimonials, the FAQ,
  and the footer.
- **`VITE_FEEDBACK_URL`** — Controls where frontend “Give feedback” links point.
  Defaults to `https://github.com/deemp/timeful/issues` when unset or blank.
- **`VITE_SUPPORT_EMAIL`** — Controls the support email address shown in the frontend.
  Support
  affordances are hidden when unset or blank.
- **`VITE_POSTHOG_API_HOST`** — Optional PostHog API host.
  Set this when analytics uses a
  self-hosted or reverse-proxied PostHog endpoint; otherwise the PostHog SDK default is used.
- **`VITE_GITHUB_REPO_URL`** — Controls where frontend GitHub links point.
  This value
  is required for Docker-built frontend artifacts.

Backend runtime variables:

- `APP_ENV`
- `APP_PORT`
- `APP_BASE_URL`
- `SERVER_BIND_HOST` (Compose host binding only)
- `CLIENT_ID`
- `CLIENT_SECRET`
- `ANDROID_CLIENT_ID`
- `IOS_CLIENT_ID`
- `MICROSOFT_CLIENT_ID`
- `MICROSOFT_CLIENT_SECRET`
- `POSTGRES_DATABASE`
- `POSTGRES_TEST_DATABASE` (test only)
- `POSTGRES_BIND_HOST`
- `POSTGRES_PORT`
- `POSTGRES_BOOTSTRAP_USERNAME`
- `POSTGRES_BOOTSTRAP_PASSWORD`
- `POSTGRES_MIGRATOR_USERNAME`
- `POSTGRES_MIGRATOR_PASSWORD`
- `POSTGRES_APPLICATION_USERNAME`
- `POSTGRES_APPLICATION_PASSWORD`
- `POSTGRES_BACKUP_USERNAME`
- `POSTGRES_BACKUP_PASSWORD`
- `POSTGRES_MIGRATOR_URI`
- `POSTGRES_APPLICATION_URI`
- `POSTGRES_CONNECT_TIMEOUT_SECONDS`
- `POSTGRES_MAX_CONNS`
- `TEST_DB_PERSIST` (test only)
- `ENCRYPTION_KEY`
- `SESSION_SECRET`
- `CORS_ORIGINS`
- `SERVICE_ACCOUNT_KEY_PATH`
- `ANALYTICS_USERNAME`
- `ANALYTICS_PASSWORD`
- `DISCORD_BOT_TOKEN`
- `DISCORD_BOT_CHANNEL`
- `GUILD_ID`
- `GOOGLE_CLOUD_PROJECT_ID`
- `GOOGLE_CLOUD_TASKS_LOCATION`
- `GOOGLE_CLOUD_TASKS_QUEUE`
- `SLACK_DEV_WEBHOOK_URL`
- `SLACK_PROD_WEBHOOK_URL`
- `MAILCHIMP_API_KEY`
- `MAILJET_API_KEY`
- `MAILJET_API_SECRET`
- `MAILJET_LIST_ID`
- `LISTMONK_ENABLED`
- `LISTMONK_URL`
- `LISTMONK_USERNAME`
- `LISTMONK_PASSWORD`
- `LISTMONK_LIST_ID`
- `LISTMONK_INITIAL_EMAIL_REMINDER_ID`
- `LISTMONK_SECOND_EMAIL_REMINDER_ID`
- `LISTMONK_FINAL_EMAIL_REMINDER_ID`
- `LISTMONK_OTP_EMAIL_TEMPLATE_ID`
- `LISTMONK_OTP_FROM_ADDRESS`
- `GMAIL_APP_PASSWORD`
- `TIMEFUL_EMAIL_ADDRESS`
- `GIN_MODE`

Observability variables for the shared OpenObserve instance (`.env.observability`):

- `OPENOBSERVE_PORT`
- `OPENOBSERVE_ROOT_USER_EMAIL`
- `OPENOBSERVE_ROOT_USER_PASSWORD`

Observability variables for the environment-scoped server ingest contract (development, staging, and production app env files):

- `OPENOBSERVE_ENDPOINT`
- `OPENOBSERVE_ORGANIZATION_ID`
- `OPENOBSERVE_INGEST_USERNAME`
- `OPENOBSERVE_INGEST_PASSWORD`

Test-only calendar provider override variables (isolated stack only):

- `TEST_GOOGLE_OAUTH_TOKEN_ENDPOINT`
- `TEST_GOOGLE_CALENDAR_API_BASE_URL`
- `TEST_MICROSOFT_OAUTH_TOKEN_ENDPOINT`
- `TEST_MICROSOFT_GRAPH_API_BASE_URL`

Manual CalDAV debug script variables:

- `APPLE_CALDAV_EMAIL`
- `APPLE_CALDAV_APP_PASSWORD`

`server/scripts/20240721_apple_calendar_test/` is a manual CalDAV debugging tool that runs outside Compose.
It reads the account email from `APPLE_CALDAV_EMAIL` and the app-specific password from `APPLE_CALDAV_APP_PASSWORD`, and exits when either is unset or blank.
Provide both values in the shell environment only; never store them in a tracked file.

Deployment environment semantics:

- `APP_ENV=development` defaults the Go server to port `3002` and defaults Gin to debug unless `GIN_MODE` overrides it.
- `APP_ENV=test` defaults the Go server to port `3003` and defaults Gin to debug unless `GIN_MODE` overrides it.
- `APP_ENV=staging` defaults the Go server to port `3004` and defaults Gin to release unless `GIN_MODE` overrides it.
- `APP_ENV=production` defaults the Go server to port `3005`, and defaults Gin to release unless `GIN_MODE` overrides it.
- `VITE_APP_ENV` is the frontend-facing mirror for browser-exposed environment-dependent behavior and should normally match `APP_ENV`.
- `APP_BASE_URL` is required and must be an absolute HTTP(S) origin without a path.
  The backend uses it for generated email links, Cloud Tasks payloads, and Slack messages.
- `CORS_ORIGINS` is an optional comma-separated list of additional browser origins.
  The normalized
  `APP_BASE_URL` is always allowed, so use this for `www`, localhost, preview, or alternate-client
  origins only.
- `APP_PORT` is required by Compose and selects both the server listener and its container port.
  `SERVER_BIND_HOST` is required by Compose and selects the host interface for that binding.
- `LISTMONK_OTP_FROM_ADDRESS` is the sender used for OTP emails.
  It must be a valid mailbox or
  RFC 5322 display-name address when an OTP email is sent.

## Precedence

- For frontend tooling, shell variables override values from the selected root env file.
- For Compose commands, shell variables passed into `docker compose --env-file ...` override values from the selected env file during interpolation.

## Commands

Development:

```sh
cp .env.development.example .env.development
docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml up --build postgres postgres-migrate server openobserve
cd frontend
npm run dev
```

Compose `up --build` builds images only for the services named on the command line, so `postgres-migrate` stays in the list to build its migration image on a cold image cache.
The development `openobserve` service is optional; omit it from the service list when the observability instance is not needed.

Staging Docker Compose:

```sh
cp .env.staging.example .env.staging
docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml up -d --build
```

Production-style local build/preview:

```sh
cp .env.production.example .env.production
cd frontend
npm run build
npm run preview
```

Production Docker Compose:

```sh
cp .env.production.example .env.production
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml up -d --build
```

Route and browser tests:

```sh
cp .env.test.example .env.test
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test
```

## Ports and isolation

| Environment        | Frontend                      | Backend host binding        | Backend container port  | PostgreSQL host binding | PostgreSQL database            |
| ------------------ | ----------------------------- | --------------------------- | ----------------------- | ----------------------- | ------------------------------ |
| Development        | `127.0.0.1:4173`              | `127.0.0.1:3002`            | `3002`                  | `127.0.0.1:5432`        | `timeful-postgres-development` |
| Test / browser E2E | `E2E_VITE_HOST:E2E_VITE_PORT` | `E2E_API_HOST:E2E_API_PORT` | `E2E_API_INTERNAL_PORT` | `127.0.0.1:5433`        | `timeful-test-*`               |
| Staging            | Caddy                         | `127.0.0.1:3004`            | `3004`                  | `127.0.0.1:5434`        | `timeful-postgres-staging`     |
| Production         | Caddy                         | `127.0.0.1:3005`            | `3005`                  | `127.0.0.1:5435`        | `timeful-postgres-production`  |

The shared Caddy edge owns public TCP ports `80` and `443` and UDP port `443`. `VITE_PREVIEW_PORT=4173` in the staging and production app env files only configures local `vite preview`; Docker deployments serve frontend artifacts through Caddy.
PostgreSQL is published only to `POSTGRES_BIND_HOST`, which defaults to `127.0.0.1` in every environment; do not change it to a public interface.
OpenObserve publishes only on `127.0.0.1` (`OPENOBSERVE_PORT`), and the shared instance is reached over the `timeful-edge` network as `openobserve:5080`; do not publish it on a public interface.
Development, test, staging, and production use distinct Compose projects, networks, and database volumes.
Browser E2E always targets the isolated test server; it must not target the development server or `timeful-postgres-development` database.

Compose has no application-value fallbacks.
Every variable it interpolates must be declared in the selected env file.
Variables with intentionally optional values may be declared blank; deployment configuration, database credentials, ports, and session secrets must be non-blank.

## Shared HTTPS edge

Local development does not run Caddy.
It uses the Vite server and its same-origin API proxy.
When staging and production share a host, a single Docker Caddy service owns public ports 80 and 443, issues certificates, and routes requests by hostname to each stack over the `timeful-edge` Docker network.

Set these values in `.env.edge` to DNS names whose `A` and, if applicable, `AAAA` records point to the host:

- `CADDY_PRODUCTION_DOMAIN`
- `CADDY_PRODUCTION_WWW_DOMAIN`
- `CADDY_PRODUCTION_UPSTREAM`
- `CADDY_STAGING_DOMAIN`
- `CADDY_STAGING_WWW_DOMAIN`
- `CADDY_STAGING_UPSTREAM`

Each canonical Caddy hostname must match the hostname in that environment's `APP_BASE_URL`.
`CADDY_STAGING_UPSTREAM` must be `staging-server:3004` and `CADDY_PRODUCTION_UPSTREAM` must be `production-server:3005`, matching the port selected by each app stack's `APP_ENV`.

Provision the shared network and artifact volumes once.
They are external so tearing down one Compose project cannot remove resources used by another:

```sh
docker network create timeful-edge
docker volume create timeful-production-frontend-dist
docker volume create timeful-staging-frontend-dist
```

Then create the edge configuration and start the edge:

```sh
cp .env.edge.example .env.edge
docker compose --env-file .env.edge -f compose.edge.yaml up -d
```

Then start each app as a separate project:

```sh
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml up -d --build
docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml up -d --build
```

The edge configuration is split into `caddy/Caddyfile`, shared handlers in `caddy/snippets/timeful.caddy`, and one site file per environment.
Keep shared routing in the snippet; site files should only provide hostnames, upstreams, and frontend roots.

Open inbound TCP ports 80 and 443 and UDP port 443.
Caddy automatically redirects HTTP to HTTPS and obtains certificates after DNS points to the host.
Update OAuth redirect URIs and allowed origins to use the configured HTTPS canonical hostnames.

## OpenObserve Observability

Staging and production share one OpenObserve instance that runs as its own `timeful-observability` Compose project on the deployment host.
Development runs its own OpenObserve instance in the development stack, and development [Observability Data](terminology/glossary.md#observability-data) is never sent to the shared instance.
Both instances use the open-source edition in single-node mode with local-disk storage, and the image is pinned by digest to `openobserve/openobserve:v1.0.1@sha256:089921e511490ef5124fe1331dfd82b475326592a1a1dad6164289e88a8ace91`.

The shared instance joins the external `timeful-edge` network, so the staging and production servers reach it at `http://openobserve:5080`, and it persists its data to the `timeful-observability-data` volume.
The development instance runs on the development network with its own project-scoped `openobserve_data` volume.
The UI and the HTTP API are published on `127.0.0.1:5080` (`OPENOBSERVE_PORT`) only; no public interface and no Caddy site exposes either instance.
Operators reach the shared UI through an SSH tunnel:

```sh
ssh -L 5080:127.0.0.1:5080 <deploy-user>@<deployment-host>
```

Start the shared instance once on the deployment host:

```sh
cp .env.observability.example .env.observability
docker compose --env-file .env.observability -f compose.observability.yaml up -d
```

The instance runs with explicit caps instead of OpenObserve's memory-consuming defaults: `ZO_MEMORY_CACHE_MAX_SIZE=128`, `ZO_MEMORY_CACHE_DATAFUSION_MAX_SIZE=256`, `ZO_MEM_TABLE_MAX_SIZE=128`, and `ZO_DISK_CACHE_MAX_SIZE=1024`, all in megabytes (QR-015).
Every [Stream](terminology/glossary.md#stream) deletes its records after the configured 14-day [Retention Window](terminology/glossary.md#retention-window), set through `ZO_COMPACT_DATA_RETENTION_DAYS=14`, and anonymous telemetry is disabled through `ZO_TELEMETRY=false` (QR-018).

### Provisioning Organizations and Credentials

Each environment ingests into its own OpenObserve organization with an [Environment-Scoped Observability Credential](terminology/glossary.md#environment-scoped-observability-credential), named `timeful_development`, `timeful_staging`, or `timeful_production`.
OpenObserve organization names allow only letters, digits, spaces, and underscores, and OpenObserve generates each organization's identifier when the organization is created.
Environment credentials authenticate only against that generated identifier, so `OPENOBSERVE_ORGANIZATION_ID` must hold the identifier rather than the organization name.
OpenObserve authorizes each **Environment-Scoped Observability Credential** only for its own organization and rejects cross-organization reads and ingests (QR-019).

Provision each instance after its first start:

1. Sign in to the UI with the root credentials from the instance's env file.
2. Create the environment's organization under the matching name.
3. Copy the organization identifier that OpenObserve generated at creation.
4. In that organization, create a service account for ingestion and copy its token, which OpenObserve shows only at creation.
5. Set `OPENOBSERVE_ORGANIZATION_ID` to the identifier, `OPENOBSERVE_INGEST_USERNAME` to the service account email, and `OPENOBSERVE_INGEST_PASSWORD` to the token in the environment's app env file.
6. Restart the environment's `server` service so it receives the updated contract.

The isolated test stack runs no OpenObserve instance and needs none of these variables.

### Server-Side Structured Diagnostics

The server consumes the environment's ingest contract on startup and ships [Observability Data](terminology/glossary.md#observability-data) over OTLP/HTTP for every [Signal](terminology/glossary.md#signal): logs to `<OPENOBSERVE_ENDPOINT>/api/<OPENOBSERVE_ORGANIZATION_ID>/v1/logs`, metrics to `/v1/metrics`, and traces to `/v1/traces`.
It authenticates every **Signal** with the same HTTP Basic credentials built from `OPENOBSERVE_INGEST_USERNAME` and `OPENOBSERVE_INGEST_PASSWORD`, so metrics and traces add no environment variables beyond the four `OPENOBSERVE_*` values and `APP_ENV` that log shipping established.
The shared resource marks every **Signal** with `service.name=timeful-server` and `deployment.environment`, so logs, metrics, and traces join on the deployment environment.

Log [Structured Log Records](terminology/glossary.md#structured-log-records) are written to the `timeful_server_logs` [Stream](terminology/glossary.md#stream).
Each record carries the request correlation identifier, the matched route template, the HTTP method and status code, the outcome, the latency in milliseconds, redacted error context, and the service's readiness state, so an operator can correlate a failed request with its records and [Service Health Status](terminology/glossary.md#service-health-status) (QR-010).
The server returns the correlation identifier to the caller as the `X-Request-ID` response header, so an operator can start from the failed response and query the matching records.
Each record also carries the native `trace_id` and `span_id` of its request span, so an operator can pivot between a log record and its trace.

Traces are written to the `timeful_server_traces` [Stream](terminology/glossary.md#stream), every request is sampled, and each request produces one server span named `METHOD /route/template`.
A request span carries the same correlation identifier, route, status, outcome, latency, readiness, and redacted error context as its log record.
PostgreSQL queries join the request trace as client spans carrying only the operation, database, host, and SQLSTATE, and outbound HTTP client spans carry only the method, scheme, host, port, status, and a coarse error class, with any 4xx or 5xx response marking the outbound span failed as otelhttp does; a call made without a request context appears as its own trace.
Dependency spans never record SQL text, query parameters, URLs, query strings, headers, bodies, or transport error text (QR-004).

Metrics are exported on a periodic reader and cover request rate, errors, and latency through the `http.server.request.duration` histogram, service readiness through the `timeful.service.readiness` gauge (1 ready, 0 unavailable), PostgreSQL pool use through `db.client.connection.count` and `db.client.connection.max`, and Go runtime health through the standard runtime instrumentation.
OpenObserve stores each metric in its own [Stream](terminology/glossary.md#stream) named after the metric with dots replaced by underscores, for example `timeful_service_readiness` and `db_client_connection_count`, and a histogram also produces its bucket, count, sum, min, and max **Streams**.
The readiness gauge carries the PostgreSQL-dependent readiness that explains dependency-related request failures, and the `service.readiness` label on request metrics lets an operator slice request outcomes by the dependency state observed at completion (QR-010, QR-017).

Export runs on bounded background pipelines for every [Signal](terminology/glossary.md#signal): request handling never waits on an exporter, a full trace queue drops spans instead of blocking, and metric recording is in-memory with a periodic background export (QR-017).
Shipped telemetry for every **Signal** excludes credentials, secrets, [Event Owner Edit Tokens](terminology/glossary.md#event-owner-edit-token), and token-bearing URLs and headers; error context passes through redaction, and the outbound HTTP transport records no URL path, query, or headers (QR-004).
When any of the four variables is missing or blank, the server disables export for all **Signals**, records a local warning, and keeps serving with its file and standard-stream [Diagnostic Output](terminology/glossary.md#diagnostic-output) only.
The GIN access log passes its request path and handler error messages through the same redaction, so token-bearing query parameters do not reach the file or standard-stream **Diagnostic Output** either.
The development stack uses the same contract against its own instance, and the isolated test stack sets no OpenObserve variables and exports nothing.
Every **Signal** inherits the instance's 14-day [Retention Window](terminology/glossary.md#retention-window) (QR-018).

## External Service Names

`GOOGLE_CLOUD_PROJECT_ID`, `GOOGLE_CLOUD_TASKS_LOCATION`, and `GOOGLE_CLOUD_TASKS_QUEUE` form the Cloud Tasks parent used for reminder jobs.
The defaults name the Timeful project and existing `us-central1` / `SendReminderEmail` resources.
Create the target project and queue, grant the configured service account access, and update these values before retiring the old Google Cloud project; Google Cloud project IDs cannot be renamed in place.

`DISCORD_BOT_CHANNEL` selects the channel used by the Discord bot.
Set it explicitly for each environment after creating the replacement channel; the Timeful defaults are only used when the variable is unset.

## PostgreSQL roles and migrations

PostgreSQL uses `postgres:18.6-bookworm` pinned to its OCI index digest.
The standard `POSTGRES_*` container bootstrap account owns initialization only.
`POSTGRES_MIGRATOR_URI` is used by the one-shot Goose migration service; `POSTGRES_APPLICATION_URI` is the server's least-privilege connection.
The `POSTGRES_BACKUP_*` role is a least-privilege read-only role used by the custom-format `pg_dump` backup procedure; restores run as the bootstrap superuser because they create and drop objects.
The isolated rehearsal reconciles the restored schema and representative migrated records, while off-host replication, automated scheduling, and recovery objectives remain later operational work.
See [PostgreSQL Operations Runbook](postgres-operations.md) for the operator procedure.

Use the selected environment's `POSTGRES_BIND_HOST` and `POSTGRES_PORT` with a local PostgreSQL client.
For example, development can be accessed with:

```sh
psql --host 127.0.0.1 --port 5432 --username timeful_postgres_admin --dbname timeful-postgres-development
```

For staging and production, connect through an SSH tunnel to the deployment host rather than exposing PostgreSQL on a public interface.

Compose starts `postgres-migrate` after PostgreSQL is healthy and starts the server only when the migration service exits successfully. `/api/health/live` reports process liveness; `/api/health` is readiness and requires PostgreSQL.
`server/migrations/` starts from a single baseline migration that creates the complete current schema on a fresh database; later schema changes are added as incremental Goose migrations on top.
Incremental migrations are forward-only and must remain compatible with the prior PostgreSQL-aware server release.
A database created before the baseline migration is recreated from it instead of upgraded in place.

## Test-only calendar provider overrides

The server resolves OAuth token endpoints and calendar API base URLs through the test-only override variables listed under backend runtime variables.
Each variable defaults to the real provider URL when it is unset or blank, so production and staging behavior is unchanged without them.
`compose.test.yaml` sets the overrides only on `server-test` and points them at the `calendar-mock` service, which exists only in the isolated test overlay and has no published ports.
Production and staging must never set any `TEST_`-prefixed provider variable.
The isolated stack also requires a real 32-byte `ENCRYPTION_KEY`, because the add-calendar flows encrypt provider credentials with AES-256-GCM and fail closed when the key is missing or the wrong length.

## Test isolation

Pure Go unit tests can run either on the host or in a container.

PostgreSQL-backed route tests and browser E2E use the isolated Compose overlay.
It runs `postgres-test` in the `timeful-test` project and uses a test-only volume, never a development database volume. `.env.test` supplies the complete server and PostgreSQL role configuration.
E2E creates a fresh `timeful-test-*` PostgreSQL database for each run.

Backend tests:

```sh
cp .env.test.example .env.test
docker volume create timeful-test-go-build-cache timeful-test-go-mod-cache
POSTGRES_TEST_DATABASE=timeful-test-postgres docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test postgres-test-bootstrap postgres-test-migrate
POSTGRES_TEST_DATABASE=timeful-test-postgres docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test
```

`server-route-test` runs `go test ./... -count=1`, so the route, account repository, and PostgreSQL store suites all run against the isolated stack.

Do not gate this sequence on `docker compose wait postgres-test-migrate`.
Compose `wait` only lists running containers, so the one-shot migrate container has usually already exited by the time the command runs and the command fails with `no containers for project`.
`server-route-test` declares `postgres-test-migrate` with the `service_completed_successfully` dependency condition, and `docker compose run` enforces it before starting the tests.

Browser E2E starts its own isolated `postgres-test` and `server-test` services, waits for `http://E2E_API_HOST:E2E_API_PORT/api/health`, and launches a fresh Vite process at `http://E2E_VITE_HOST:E2E_VITE_PORT`. `server-test` listens on `E2E_API_INTERNAL_PORT`; Compose publishes it at `E2E_API_HOST:E2E_API_PORT`.
It inherits the complete `.env.test` server environment contract.
The E2E harness overrides only the generated PostgreSQL database name; `.env.test` clears external integration secrets to prevent side effects:

```sh
cd e2e
npm run test:e2e
```

Browser E2E creates supported events in PostgreSQL by default; no creation flag is required.

Local browser runs default to two workers, while CI defaults to one; `--workers=1`, `--workers=2`, and `--workers=4` override the shared worker budget, including Firefox.
Keep concurrent tests within one Playwright invocation because separate invocations share the test-stack project and ports.
Existing serial test groups preserve their internal ordering.
Ordinary projects start Vite without building production assets.
The `chromium-production-desktop` and `chromium-production-mobile` projects depend on a shared fresh production build, and unfiltered `npm run test:e2e` includes them automatically.
Set `E2E_FRONTEND=bundled` to have the webServer build a fresh test-mode frontend and serve it from a Playwright-owned preview instead of the dev server's unbundled modules; this is an opt-in speedup for the heavier recorded journeys.
See [fast local runs](../e2e/AGENTS.md#fast-local-runs) for focused commands, the bundled mode, and the production-asset verification workflow.

`TEST_DB_PERSIST` defaults to `false`, removing the test stack and its database volume after E2E for repeatable runs.
Set it to `true` to stop only the test server and retain the database state after successful or failed E2E setup for inspection.

`server-test` and `server-route-test` share a persistent Go build cache in the external `timeful-test-go-build-cache` volume (`GOCACHE=/go-build-cache`), so `go run` and `go test` compile incrementally instead of from cold on every container start.
They also share the external `timeful-test-go-mod-cache` volume mounted at `/go/pkg/mod` (`GOMODCACHE=/go/pkg/mod`), which carries the downloaded Go modules.
Compose does not create external volumes, so the route-test snippet creates both first; `docker volume create` is idempotent when the volume already exists.
Backend CI runs the same step before its `compose run` because a fresh runner has no volumes.
Compose `down -v` does not remove them (they are external); delete them with `docker volume rm timeful-test-go-build-cache timeful-test-go-mod-cache` to force a clean compile and module re-download.

Remove persistent test state explicitly:

```sh
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml down -v
```
