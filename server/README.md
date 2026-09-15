# Timeful API

API docs (available when the server is running): http://localhost:3002/swagger/index.html

## Development

The server is configured and run through Docker Compose from the repository root.
Create the canonical root development env file, then start the development stack:

```sh
cp .env.development.example .env.development
docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml up --build postgres postgres-migrate server
```

`postgres-migrate` is named on the command line because Compose `up --build` builds images only for the services listed there.

See `docs/environments.md` for the complete configuration contract.
Direct server execution and `server/.env` are unsupported.

## Tests

Pure unit tests can run on the host or in a container.

PostgreSQL-backed route tests should use the isolated Compose test stack from the repo root:

```sh
cp .env.test.example .env.test
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml down -v
```

The `down -v` cleanup is scoped to the isolated `timeful-test` Compose stack and its test-only PostgreSQL volume.

When running PostgreSQL-backed tests directly on the host, set `POSTGRES_APPLICATION_URI` to a dedicated test database first.
The database must be `timeful-test` or start with `timeful-test-`.
