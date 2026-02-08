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

    -- =========================================================
    -- Hotpot schemas
    -- =========================================================
    CREATE SCHEMA IF NOT EXISTS bronze;
    CREATE SCHEMA IF NOT EXISTS bronze_history;
    CREATE SCHEMA IF NOT EXISTS silver;
    CREATE SCHEMA IF NOT EXISTS gold;

    -- =========================================================
    -- Metabase read-only user
    -- =========================================================
    DO $$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'metabase') THEN
            CREATE USER metabase WITH PASSWORD 'metabase';
        END IF;
    END
    $$;

    GRANT USAGE ON SCHEMA bronze TO metabase;
    GRANT USAGE ON SCHEMA bronze_history TO metabase;
    GRANT USAGE ON SCHEMA silver TO metabase;
    GRANT USAGE ON SCHEMA gold TO metabase;

    GRANT SELECT ON ALL TABLES IN SCHEMA bronze TO metabase;
    GRANT SELECT ON ALL TABLES IN SCHEMA bronze_history TO metabase;
    GRANT SELECT ON ALL TABLES IN SCHEMA silver TO metabase;
    GRANT SELECT ON ALL TABLES IN SCHEMA gold TO metabase;

    ALTER DEFAULT PRIVILEGES IN SCHEMA bronze GRANT SELECT ON TABLES TO metabase;
    ALTER DEFAULT PRIVILEGES IN SCHEMA bronze_history GRANT SELECT ON TABLES TO metabase;
    ALTER DEFAULT PRIVILEGES IN SCHEMA silver GRANT SELECT ON TABLES TO metabase;
    ALTER DEFAULT PRIVILEGES IN SCHEMA gold GRANT SELECT ON TABLES TO metabase;
EOSQL

echo "Hotpot dev databases and schemas initialized successfully"
