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

## Transpiled GALA sources

Generated Go files are committed because the Go build never invokes GALA.
Each one has a `.gala` source beside it, and some packages also have a handwritten sibling for members GALA cannot express.

| Generated file                              | GALA source                                   | Regeneration command                                                                                 |
| ------------------------------------------- | --------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `eventid/eventid.go`                        | `eventid/eventid.gala`                        | `cd server/eventid && gala transpile -i eventid.gala -o eventid.go`                                  |
| `observability/redact.go`                   | `observability/redact.gala`                   | `cd server/observability && gala transpile -i redact.gala -o redact.go`                              |
| `logger/logger.go`                          | `logger/logger.gala`                          | `cd server/logger && gala transpile -i logger.gala -o logger.go`                                     |
| `appenv/appenv.go`                          | `appenv/appenv.gala`                          | `cd server/appenv && gala transpile -i appenv.gala -o appenv.go`                                     |
| `utils/array_utils.go`                      | `utils/array_utils.gala`                      | `cd server/utils && gala transpile -i array_utils.gala -o array_utils.go`                            |
| `utils/request_utils.go`                    | `utils/request_utils.gala`                    | `cd server/utils && gala transpile -i request_utils.gala -o request_utils.go`                        |
| `services/services.go`                      | `services/services.gala`                      | `cd server/services && gala transpile -i services.gala -o services.go`                               |
| `services/providerconfig/providerconfig.go` | `services/providerconfig/providerconfig.gala` | `cd server/services/providerconfig && gala transpile -i providerconfig.gala -o providerconfig.go`    |
| `routes/guest_response_ownership.go`        | `routes/guest_response_ownership.gala`        | `cd server/routes && gala transpile -i guest_response_ownership.gala -o guest_response_ownership.go` |
| `discord_bot/commands/help.go`              | `discord_bot/commands/help.gala`              | `cd server/discord_bot/commands && gala transpile -i help.gala -o help.go`                           |
| `discord_bot/commands/num_users.go`         | `discord_bot/commands/num_users.gala`         | `cd server/discord_bot/commands && gala transpile -i num_users.gala -o num_users.go`                 |
| `discord_bot/init.go`                       | `discord_bot/init.gala`                       | `cd server/discord_bot && gala transpile -i init.gala -o init.go`                                    |
| `slackbot/commands/num_users.go`            | `slackbot/commands/num_users.gala`            | `cd server/slackbot/commands && gala transpile -i num_users.gala -o num_users.go`                    |

Handwritten siblings beside generated files are `appenv/appenv_port.go` (`ResolvePort`), `utils/array_utils_extra.go` (`ArrayToSet`, `ElementWithIndex`, `FindAddedRemovedKept`), `services/providerconfig/doc.go` (package comment), and `slackbot/commands/utils.go` (the `newResponse` constructor).
They are not generated and have no regeneration command.
See `GALA.md` for the two output styles, the vendored runtime under `third_party/gala/`, and the findings from the transpilation spikes.

## Tests

Pure unit tests can run on the host or in a container.

PostgreSQL-backed route tests use the isolated Compose test stack from the repo root.
This is the canonical command sequence for backend tests; `docs/environments.md` documents the isolation semantics.

```sh
cp .env.test.example .env.test
docker volume create timeful-test-go-build-cache timeful-test-go-mod-cache
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test postgres-test-bootstrap postgres-test-migrate
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test
```

Compose never creates the external Go cache volumes, and `docker volume create` is idempotent when they already exist.
Test state is retained by default; remove it only when it is no longer needed:

```sh
docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml down -v
```

The `down -v` cleanup is scoped to the isolated `timeful-test` Compose stack and its test-only PostgreSQL volume.
It does not remove the external Go cache volumes; `docs/environments.md` documents how to reset them.

When running PostgreSQL-backed tests directly on the host, set `POSTGRES_APPLICATION_URI` to a dedicated test database first.
The database must be `timeful-test` or start with `timeful-test-`.
