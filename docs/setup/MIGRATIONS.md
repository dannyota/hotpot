# Database Migrations

Atlas schema migrations for Hotpot's multi-layer database (bronze, bronzehistory, silver, gold).

## Quick Start

```bash
# Build the tool
make migrate-tool
# or: go build -o bin/migrate ./cmd/migrate

# Generate migration
bin/migrate diff migration_name

# Apply migration
bin/migrate apply
```

## How It Works

The `migrate` tool:

1. Loads database config from **Vault or YAML** (same as ingest)
2. Converts DSN to PostgreSQL URL format
3. Sets `HOTPOT_DATABASE_URL` environment variable
4. Runs `atlas migrate` commands for each layer in order
5. Atlas reads the database URL from the env var (configured in `atlas.hcl`)
6. Applies migrations sequentially: bronze → bronzehistory → silver → gold

**Security:** Credentials are passed to Atlas via the `HOTPOT_DATABASE_URL` environment variable, not command-line arguments. This prevents passwords from appearing in `ps aux` or process lists.

## Commands

### Generate Migration

```bash
bin/migrate diff <name>
```

Creates new migration files in `migrations/{layer}/` for each layer.

**Example:**
```bash
bin/migrate diff add_firewall_tables
```

**Output:**
```
==> bronze: atlas migrate diff add_firewall_tables --env bronze
==> bronzehistory: atlas migrate diff add_firewall_tables --env bronzehistory
==> silver: atlas migrate diff add_firewall_tables --env silver
==> gold: atlas migrate diff add_firewall_tables --env gold
✅ Migration complete
```

### Apply Migration

```bash
bin/migrate apply
```

Applies all pending migrations to the database.

**Example:**
```bash
bin/migrate apply
```

## Configuration

The tool uses the same config system as the ingest worker:

**YAML file (config.yaml):**
```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  dbname: hotpot
  sslmode: disable
  dev_dbname: hotpot_dev  # Optional - defaults to "{dbname}_dev"
```

**Database vs Dev Database:**

| Field | Purpose | Default |
|-------|---------|---------|
| `dbname` | Target database for migrations | Required |
| `dev_dbname` | Scratch database for Atlas diffs | `{dbname}_dev` |

The dev database is a **scratch/sandbox database** used by Atlas to compute schema diffs. It never contains production data.

**⚠️ CRITICAL:** The dev database **MUST be different** from your production database. Atlas will DROP and RECREATE tables in the dev database during `migrate diff`. The migrate tool includes a safety check to prevent this.

**Environment variables:**
```bash
CONFIG_SOURCE=file
CONFIG_FILE=config.yaml
```

**Vault:**
```bash
CONFIG_SOURCE=vault
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=your-token
VAULT_PATH=secret/hotpot/config
```

## Migration Files

Migrations are stored in `deploy/migrations/` per layer:

```
deploy/
└── migrations/
    ├── atlas.hcl           # Atlas configuration
    ├── bronze/
    │   ├── 20260101_init.sql
    │   └── 20260108_add_firewall.sql
    ├── bronzehistory/
    │   ├── 20260101_init.sql
    │   └── 20260108_add_firewall.sql
    ├── silver/
    │   └── 20260101_init.sql
    └── gold/
        └── 20260101_init.sql
```

## Atlas Configuration

Atlas config is in `deploy/migrations/atlas.hcl`:

```hcl
env "bronze" {
  src = "ent://../../pkg/storage/ent/bronze/atlas_schema"
  dev = env("HOTPOT_DEV_DATABASE_URL") # Scratch DB for diffing
  url = env("HOTPOT_DATABASE_URL")     # Target DB for migrations
  migration {
    dir = "file://bronze"
  }
}
```

**Environment variables set by migrate tool:**
- `HOTPOT_DATABASE_URL` - Target database from config
- `HOTPOT_DEV_DATABASE_URL` - Dev database (defaults to `{dbname}_dev`)

Each layer has its own environment with separate schema and migration directory.

## Workflow

### After Schema Changes

1. **Modify ent schemas** in `pkg/schema/`
2. **Generate ent code**: `cd pkg/storage && go generate`
3. **Generate migration**: `./migrate diff description`
4. **Review migration** files in `migrations/{layer}/`
5. **Apply migration**: `./migrate apply`

### Example

```bash
# 1. Edit schema
vim pkg/schema/bronze/gcp/compute/firewall.go

# 2. Generate ent code
cd pkg/storage && go generate && cd ../..

# 3. Generate migration
./migrate diff add_gcp_compute_firewall

# 4. Review
cat migrations/bronze/20260108_add_gcp_compute_firewall.sql

# 5. Apply
./migrate apply
```

## Setup

### Create Dev Database

Before running migrations, create the dev database:

```bash
# If your main database is "hotpot", create "hotpot_dev"
psql -U postgres -c "CREATE DATABASE hotpot_dev;"
```

Or specify a custom dev database in config:

```yaml
database:
  dbname: hotpot
  dev_dbname: my_custom_dev_db
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Database DSN not configured" | Set database config in YAML or Vault |
| "Dev database DSN not configured" | Create `{dbname}_dev` database or set `dev_dbname` |
| "atlas: command not found" | Install Atlas CLI: `brew install ariga/tap/atlas` |
| Migration fails on one layer | Check `migrations/{layer}/` for conflicts |
| "incomplete DSN" | Ensure database.host, port, user, dbname are set |
| Dev database doesn't exist | Run `CREATE DATABASE hotpot_dev;` |

## Migration Tool Features

| Feature | Details |
|---------|---------|
| Database config | From Vault/YAML (centralized) |
| Credentials | Never hardcoded |
| Config validation | Required fields checked on load |
| Layer order | Bronze → bronzehistory → silver → gold |
| Error handling | Stops on first failure |
