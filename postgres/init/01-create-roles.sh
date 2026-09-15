#!/bin/sh
set -eu

psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
  --set=ON_ERROR_STOP=1 \
  --set=database_name="$POSTGRES_DB" \
  --set=migrator_username="$POSTGRES_MIGRATOR_USERNAME" \
  --set=migrator_password="$POSTGRES_MIGRATOR_PASSWORD" \
  --set=application_username="$POSTGRES_APPLICATION_USERNAME" \
  --set=application_password="$POSTGRES_APPLICATION_PASSWORD" \
  --set=backup_username="$POSTGRES_BACKUP_USERNAME" \
  --set=backup_password="$POSTGRES_BACKUP_PASSWORD" <<'EOSQL'
CREATE ROLE :"migrator_username" LOGIN PASSWORD :'migrator_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE :"application_username" LOGIN PASSWORD :'application_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE :"backup_username" LOGIN PASSWORD :'backup_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;

REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE, CREATE ON SCHEMA public TO :"migrator_username";
GRANT CONNECT ON DATABASE :"database_name" TO :"migrator_username", :"application_username", :"backup_username";
GRANT pg_read_all_data TO :"backup_username" WITH INHERIT TRUE;
EOSQL
