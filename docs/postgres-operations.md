# PostgreSQL Operations Runbook

PostgreSQL is the only supported database.
This runbook documents the ongoing operational procedures: backup and restore, the event name constraint lifecycle, and retention cleanup.

## Backup And Restore

Backups use the PostgreSQL 18 client tools with a custom-format `pg_dump` and the least-privilege backup role.
The backup role needs read access to every table, granted at creation and re-applied to existing databases with:

```sql
GRANT pg_read_all_data TO <backup role> WITH INHERIT TRUE;
```

The commands run inside the `postgres` container so the role names and database name come from the container environment and local socket authentication applies; no host environment file is sourced and no password is interpolated.
Run the commands on the deployment host from a checkout of the release:

```sh
# Backup the selected environment.
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml exec -T postgres \
  sh -ec 'pg_dump --format=custom --no-owner --username "$POSTGRES_BACKUP_USERNAME" --dbname "$POSTGRES_DB"' \
  > "timeful-$(date -u +%Y%m%dT%H%M%SZ).dump"

# Restore into a scratch maintenance database for verification.
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml exec -T postgres \
  sh -ec 'createdb --username "$POSTGRES_USER" timeful-restore-check'
docker compose --project-name timeful-production --env-file .env.production -f compose.yaml -f compose.production.yaml exec -T postgres \
  sh -ec 'pg_restore --no-owner --exit-on-error --username "$POSTGRES_USER" --dbname timeful-restore-check' < "timeful-<timestamp>.dump"

# Reconcile per-table row counts and key content digests before declaring the backup verified.
```

Both restore commands run as the container's bootstrap superuser, because `pg_restore` creates and drops objects that the read-only backup role cannot.
The destructive restore over the live database adds `--clean --if-exists` and requires a change ticket.
A backup is verified only after its restore reconciliation reports matching counts and digests.
Off-host replication, automated scheduling, and recovery objectives remain later operational work.

## Event Name Constraint

The FR-119 event-name guard is added `NOT VALID` by `20260913000002_postgres_events_name_length.sql` so a database that predates the baseline cannot fail that migration on legacy rows, while every new insert and update is still checked.
`20260914000000_validate_postgres_events_name_length.sql` then validates the constraint, so a database that applies the full migration chain ends with a validated guard.
A database created from the baseline has no legacy rows, so the validation migration succeeds there without cleanup.
Do not validate or clean rows in a database that predates the baseline; recreate it from the baseline instead, which applies the validation migration as part of the chain.
Confirm the constraint state after the application stack reports healthy against a database:

```sql
SELECT convalidated FROM pg_constraint
WHERE conrelid = 'postgres_events'::regclass
    AND conname = 'postgres_events_name_length';
```

A result of `f` means the database has not applied the full chain, and an empty result means it has not applied the constraint migration at all.
Record either state and follow the recreate-from-baseline policy in [Retention Cleanup](#retention-cleanup) instead of cleaning rows in place.
The validation migration fails and aborts the goose run when a database still holds an offending row; that is the intended guard, not a reason to drop the constraint or leave it `NOT VALID`.
Detect offenders before rerunning the migration:

```sql
SELECT id, name FROM postgres_events
WHERE name = '' OR char_length(name) > 100;
```

On a baseline-created database, repair each row through the update API or direct SQL with a valid name, or delete the row, then rerun the migration.
On a database that predates the baseline, do not repair rows; recreate it from the baseline.

## Retention Cleanup

The schema is never dropped to roll back a release.
The pre-baseline migration chain is retired, and `server/migrations/` contains only the baseline migration plus any later incremental migrations.
The baseline does not create the retired migration-tooling tables `migration_ledger` and `migration_quarantine`, so a database created from the baseline has no migration records to clean up.
A database that still holds those tables predates the baseline; recreate it from the baseline instead of cleaning it in place, following [Environments](environments.md).
Until the replacement database is validated, keep the verified backup of the old database intact.
