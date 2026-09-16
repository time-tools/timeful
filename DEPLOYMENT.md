# Timeful Deployment Guide

Production and staging deployment using Docker Compose behind one shared Docker Caddy edge.

## Prerequisites

- Docker and Docker Compose
- The deploy user can run Docker.
  Either add it to the `docker` group or prefix the commands below with `sudo`.
- A domain with DNS pointing to your server before Caddy starts
- Inbound TCP ports 80 and 443, plus UDP port 443

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/deemp/timeful
cd timeful

# 2. Create app environment files and the shared edge configuration.
cp .env.production.example .env.production
cp .env.staging.example .env.staging
cp .env.edge.example .env.edge

# Edit the app env files and .env.edge with their environment-specific values (see Configuration below).

# 3. Configure APP_BASE_URL in both app env files and matching CADDY_* values in .env.edge.
# Start the shared HTTPS edge once.
docker network create timeful-edge
docker volume create timeful-production-frontend-dist
docker volume create timeful-staging-frontend-dist
docker compose --env-file .env.edge -f compose.edge.yaml up -d

# 4. Configure and start the shared observability instance once.
cp .env.observability.example .env.observability
# Edit .env.observability and set a strong OPENOBSERVE_ROOT_USER_PASSWORD before the first start.
docker compose --env-file .env.observability -f compose.observability.yaml up -d

# 5. Build and start production.
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml up -d --build

# Start staging independently when needed.
# docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml up -d --build
```

### Staging Only

Use the staging-only edge when production is not running on this server.
It requires only `.env.staging`, `.env.edge`, and the staging frontend volume:

```bash
docker network inspect timeful-edge >/dev/null 2>&1 || docker network create timeful-edge
docker volume inspect timeful-staging-frontend-dist >/dev/null 2>&1 || docker volume create timeful-staging-frontend-dist
docker compose --env-file .env.edge -f compose.edge.staging.yaml up -d
docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml up -d --build
```

Only one Caddy edge can bind ports 80 and 443.
Stop the staging-only edge before starting the shared production-and-staging edge.

## Ports

| Environment        | Frontend         | Backend host binding | Backend container port | PostgreSQL host binding |
| ------------------ | ---------------- | -------------------- | ---------------------- | ----------------------- |
| Development        | `127.0.0.1:4173` | `127.0.0.1:3002`     | `3002`                 | `127.0.0.1:5432`        |
| Test / browser E2E | `127.0.0.1:4174` | `127.0.0.1:3003`     | `3003`                 | `127.0.0.1:5433`        |
| Staging            | Caddy            | `127.0.0.1:3004`     | `3004`                 | `127.0.0.1:5434`        |
| Production         | Caddy            | `127.0.0.1:3005`     | `3005`                 | `127.0.0.1:5435`        |

The shared Caddy edge owns public TCP ports `80` and `443` and UDP port `443`. `VITE_PREVIEW_PORT=4173` in the staging and production app env files only configures local `vite preview`; Docker deployments serve frontend artifacts through Caddy.

## Shared Caddy Edge

One Docker Caddy service owns public ports 80 and 443 for both environments.
It routes each hostname to the matching private backend and frontend artifacts over the `timeful-edge` Docker network.
Caddy handles:

- Automatic HTTPS certificates
- HTTP → HTTPS redirect
- www → non-www redirect
- Compression (gzip/zstd)
- Security headers

The root config imports a shared Timeful snippet and separate staging/production site files:

```text
caddy/Caddyfile
caddy/snippets/timeful.caddy
caddy/sites/production.caddy
caddy/sites/staging.caddy
```

The production and staging hostnames and upstreams are read from `.env.edge`.
Each upstream must match its app stack's `APP_ENV` port: `staging-server:3004` for staging and `production-server:3005` for production.
DNS for every canonical and `www` hostname must point to the server before Caddy can obtain its certificates.

For `staging.timeful.fun` on a server with IPv4 address `192.144.13.176`, create these DNS records before starting Caddy:

| Type | Name          | Value            |
| ---- | ------------- | ---------------- |
| `A`  | `staging`     | `192.144.13.176` |
| `A`  | `www.staging` | `192.144.13.176` |

Do not create an `AAAA` record unless the server has a reachable IPv6 address.
Caddy logs an ACME DNS error and cannot issue HTTPS certificates until every configured hostname resolves.

## Observability

Staging and production send [Observability Data](docs/terminology/glossary.md#observability-data) to one shared OpenObserve instance that runs as its own `timeful-observability` Compose project, separate from the app stacks.
The instance joins the external `timeful-edge` network, so each server reaches it at `http://openobserve:5080`, uses single-node local-disk storage, and persists to the `timeful-observability-data` volume.
Its UI and HTTP API are published on `127.0.0.1:5080` only, no Caddy site routes to it, and operators reach it through an SSH tunnel:

```bash
ssh -L 5080:127.0.0.1:5080 <deploy-user>@<deployment-host>
```

Start and manage the shared instance with its own env file and override:

```bash
docker compose --env-file .env.observability -f compose.observability.yaml up -d
docker compose --env-file .env.observability -f compose.observability.yaml logs -f
docker compose --env-file .env.observability -f compose.observability.yaml down
```

The instance runs with explicit memory and disk caps instead of OpenObserve's defaults: `ZO_MEMORY_CACHE_MAX_SIZE=128`, `ZO_MEMORY_CACHE_DATAFUSION_MAX_SIZE=256`, `ZO_MEM_TABLE_MAX_SIZE=128`, and `ZO_DISK_CACHE_MAX_SIZE=1024`, all in megabytes (QR-015).
Every [Stream](docs/terminology/glossary.md#stream) deletes its records after the configured 14-day [Retention Window](docs/terminology/glossary.md#retention-window) through `ZO_COMPACT_DATA_RETENTION_DAYS=14`, and anonymous telemetry stays disabled (QR-018).
The shared instance data is expendable operational data and is not backed up; after volume loss, start from an empty instance and re-provision it.

### Provisioning

Provision the shared instance after its first start, and provision the development instance the same way after starting the development stack:

1. Open `http://127.0.0.1:5080` through the SSH tunnel for the shared instance, or directly for the development instance, and sign in with the root credentials from the instance's env file.
2. Create the environment's organization under the matching name (`timeful_development`, `timeful_staging`, or `timeful_production`); names allow only letters, digits, spaces, and underscores.
3. Copy the organization identifier that OpenObserve generated at creation.
4. In that organization, create a service account for ingestion and copy its token, which OpenObserve shows only at creation.
5. Set `OPENOBSERVE_ORGANIZATION_ID` to the identifier, `OPENOBSERVE_INGEST_USERNAME` to the service account email, and `OPENOBSERVE_INGEST_PASSWORD` to the token in the matching app env file.
6. Restart that environment's `server` service.

Each OpenObserve service account is an [Environment-Scoped Observability Credential](docs/terminology/glossary.md#environment-scoped-observability-credential) that reads and ingests only within its own environment, and OpenObserve rejects cross-organization requests (QR-019).
Environment credentials authenticate only against the organization identifier, not the organization name.
The isolated test stack runs no OpenObserve instance and needs none of these variables.

### Server Diagnostics

The server ships [Structured Log Records](docs/terminology/glossary.md#structured-log-records) over OTLP/HTTP to its environment's organization and writes them to the `timeful_server_logs` stream.
Each record carries the request correlation identifier, route, status, outcome, latency, redacted error context, and the service's readiness state, and the server returns the identifier as the `X-Request-ID` response header.
To diagnose a failed request, open the Logs view through the SSH tunnel, select the environment's organization and the `timeful_server_logs` stream, and query the identifier from the failed response.
Export runs on a bounded background pipeline, so OpenObserve being unavailable or slow never blocks requests, and the file and standard-stream diagnostics stay available without it.
When the environment's OpenObserve variables are incomplete, the server logs a warning and disables export instead of failing to start.

## Commands

> [!CAUTION]
> Use `down -v` only when intentionally discarding the Docker-managed data volumes for the selected environment.
>
> For production, this deletes the PostgreSQL data volume unless a backup is restored afterward.

```bash
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml up -d              # Start services
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml logs -f            # View logs
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml logs -f server     # View specific service logs
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml up -d --build      # Rebuild after code changes
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml down               # Stop services
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml down -v            # Stop and remove volumes (deletes data!)
```

Staging uses the same base commands with the staging env file and override:

```bash
docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml up -d --build
docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml logs -f
docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml down
```

## Upgrading an Existing Deployment

Back up PostgreSQL before changing the database authentication contract.

```bash
git pull --autostash origin main
```

If the pull reports a conflict for the retired root `Caddyfile`, retain the new `caddy/` layout and move any custom host rules into a file under `caddy/sites/`.
The autostash remains available until it is explicitly dropped.

## Validation

After deployment, confirm the selected application stack is healthy and the public edge is serving it:

```bash
docker compose --project-name timeful-staging --env-file .env.staging -f compose.yaml -f compose.staging.yaml ps
curl -fsS https://staging.timeful.fun/api/health
docker compose --env-file .env.edge -f compose.edge.staging.yaml logs --tail=50 caddy
```

## Data & Backup

Data is persisted in Docker volumes: `postgres_data`, `frontend_dist`, `server_logs`.
The shared observability instance persists to `timeful-observability-data`, which holds expendable operational data that is not backed up.

PostgreSQL uses a digest-pinned 18.6 image and a one-shot Goose migration service before the server starts.
Its bootstrap, migrator, application, and backup roles require separate credentials and role-specific connection URIs.
PostgreSQL backups use a custom-format `pg_dump` executed with the least-privilege backup role, and restores use `pg_restore`.
The commands run inside the database container, so role names and the database name come from the container environment and local socket authentication applies.
Reconcile the restored schema and representative records after every restore, and the full procedure lives in [PostgreSQL Operations Runbook](docs/postgres-operations.md).
Off-host replication, automated scheduling, recovery objectives, and destructive restore drills remain later operational work, so do not treat the provisioned backup role as a complete recovery mechanism.

The restore command below uses `--clean --if-exists` for PostgreSQL.

> [!CAUTION]
> Run them only when you intend to replace the configured database with the backup archive.

```bash
# Backup PostgreSQL with the least-privilege backup role
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml exec -T postgres \
  sh -ec 'pg_dump --format=custom --no-owner --username "$POSTGRES_BACKUP_USERNAME" --dbname "$POSTGRES_DB"' \
  > "timeful-$(date -u +%Y%m%dT%H%M%SZ).dump"

# Restore PostgreSQL into the configured database
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml exec -T postgres \
  sh -ec 'pg_restore --clean --if-exists --no-owner --exit-on-error --username "$POSTGRES_USER" --dbname "$POSTGRES_DB"' < "timeful-<timestamp>.dump"
```

## Troubleshooting

```bash
# Container won't start
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml logs server
ls -la .env.production

# Frontend not loading
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml logs frontend-artifacts
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml exec server ls -la /app/frontend/dist
```

---

## Configuration

### Required Environment Variables

Create `.env.production` from `.env.production.example` for production, or `.env.staging` from `.env.staging.example` for staging.
Create `.env.edge` from `.env.edge.example` for Caddy.
Create `.env.observability` from `.env.observability.example` for the shared observability instance.

The selected root app env file is the single source of truth for:

- Docker Compose interpolation
- frontend build args
- backend runtime configuration
- OpenObserve ingest contract for the server

See `docs/environments.md` for the full contract and development commands.

`.env.edge` contains only the public Caddy hostnames.
Its canonical production and staging domains must match the hostnames in the respective app file's `APP_BASE_URL`.

`.env.observability` contains the shared observability instance's loopback port and root credentials.
Each app env file also carries that environment's OpenObserve ingest contract; see [Observability](#observability).

#### Required To Start

| Variable                                               | Description                                                                         |
| ------------------------------------------------------ | ----------------------------------------------------------------------------------- |
| `ENCRYPTION_KEY`                                       | Key for encrypting sensitive data (generate with `openssl rand -base64 32`)         |
| `SESSION_SECRET`                                       | Session cookie encryption key (generate with `openssl rand -base64 32`)             |
| `APP_BASE_URL`                                         | Canonical public HTTPS origin used in generated links and payment redirects         |
| `POSTGRES_DATABASE`                                    | PostgreSQL database name; defaults are environment-specific                         |
| `POSTGRES_BOOTSTRAP_*`                                 | PostgreSQL container bootstrap account                                              |
| `POSTGRES_MIGRATOR_*` / `POSTGRES_MIGRATOR_URI`        | Goose migration role and URL-encoded connection URI                                 |
| `POSTGRES_APPLICATION_*` / `POSTGRES_APPLICATION_URI`  | Runtime role and URL-encoded connection URI                                         |
| `POSTGRES_BACKUP_*`                                    | Least-privilege read-only role for PostgreSQL backups                               |
| `OPENOBSERVE_ENDPOINT` / `OPENOBSERVE_ORGANIZATION_ID` | OpenObserve ingest endpoint and the environment's generated organization identifier |

`CADDY_PRODUCTION_DOMAIN`, `CADDY_PRODUCTION_WWW_DOMAIN`, and `CADDY_PRODUCTION_UPSTREAM`, or their staging equivalents, are required in `.env.edge` by the Caddy edge that serves that environment.
The upstream must match the server port selected by that app file's `APP_ENV`.

#### Required For Enabled Features

| Variable                                                      | Feature                                 |
| ------------------------------------------------------------- | --------------------------------------- |
| `CLIENT_ID` / `CLIENT_SECRET`                                 | Google sign-in and calendar integration |
| `OPENOBSERVE_INGEST_USERNAME` / `OPENOBSERVE_INGEST_PASSWORD` | Authenticated OpenObserve ingest        |

#### Optional — Payments

| Variable                | Description                        |
| ----------------------- | ---------------------------------- |
| `STRIPE_API_KEY`        | Stripe API key                     |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret      |
| `STRIPE_*_PRICE_ID`     | Stripe price IDs for various plans |

#### Optional — Additional Calendars

| Variable                  | Description                             |
| ------------------------- | --------------------------------------- |
| `MICROSOFT_CLIENT_ID`     | Microsoft OAuth client ID (for Outlook) |
| `MICROSOFT_CLIENT_SECRET` | Microsoft OAuth client secret           |

#### Optional — CORS

| Variable       | Description                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `CORS_ORIGINS` | Comma-separated additional browser origins, such as the `www` hostname. `APP_BASE_URL` is always allowed. |

#### Optional — Other Services

| Variable                                       | Description                                                                                 |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `ANALYTICS_USERNAME` / `ANALYTICS_PASSWORD`    | Basic auth for /api/analytics routes                                                        |
| `SERVICE_ACCOUNT_KEY_PATH`                     | Google Cloud service account for Cloud Tasks                                                |
| `SLACK_*_WEBHOOK_URL`                          | Slack webhooks for notifications                                                            |
| `GMAIL_APP_PASSWORD` / `TIMEFUL_EMAIL_ADDRESS` | Gmail SMTP for sending emails                                                               |
| `LISTMONK_*`                                   | Listmonk email service configuration, including `LISTMONK_OTP_FROM_ADDRESS` for OTP senders |
| `VITE_SUPPORT_EMAIL`                           | Support email embedded in frontend artifacts                                                |
| `DISCORD_BOT_TOKEN` / `GUILD_ID`               | Discord bot integration                                                                     |

See `.env.production.example`, `.env.staging.example`, and `.env.edge.example` for the complete lists.

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable the following APIs:
   - Google Calendar API
   - People API (Contacts)
   - Admin SDK API (Directory)
4. Create OAuth 2.0 credentials (Web application type)
5. Add authorized redirect URIs:
   - `https://yourdomain.com/api/auth/callback`
   - `http://localhost:3002/api/auth/callback` (for development)
6. Copy the Client ID and Client Secret to the applicable root app env file, such as `.env.production`
