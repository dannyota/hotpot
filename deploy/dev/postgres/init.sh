#!/bin/bash
set -e

# Dev environment init script for PostgreSQL
# Creates databases and schemas needed by hotpot and Temporal

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-'EOSQL'
    -- =========================================================
    -- Temporal: user + databases
    -- =========================================================
    DO $$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'temporal') THEN
            CREATE USER temporal WITH PASSWORD 'temporal' CREATEDB;
        END IF;
    END
    $$;

    SELECT 'CREATE DATABASE temporal OWNER temporal'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'temporal')\gexec

    SELECT 'CREATE DATABASE temporal_visibility OWNER temporal'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'temporal_visibility')\gexec

    GRANT ALL PRIVILEGES ON DATABASE temporal TO temporal;
    GRANT ALL PRIVILEGES ON DATABASE temporal_visibility TO temporal;

    -- =========================================================
    -- Hotpot dev database (for Atlas migration diffing)
    -- =========================================================
    SELECT 'CREATE DATABASE hotpot_dev OWNER postgres'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'hotpot_dev')\gexec


EOSQL

# NOTE: Do NOT create schemas in hotpot_dev — Atlas manages the dev DB from
# scratch during 'migrate diff' and will create schemas itself.

echo "Hotpot dev databases and schemas initialized successfully"
