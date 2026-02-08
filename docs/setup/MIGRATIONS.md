# Database Migrations

Atlas schema migrations for Hotpot's multi-layer database (bronze, bronzehistory, silver, gold).

## Tools

There are two separate tools:

| Tool | Purpose | When to use |
|------|---------|-------------|
| `bin/migrate` | Apply migrations to a database | Production deployments |
| `bin/genmigrate` | Generate migration SQL files | Development (after schema changes) |

The `migrate` binary is self-contained — it embeds all SQL migration files, so no extra files are needed at deploy time.

## Quick Start

```bash
# Build the tools
make build

# Generate migration (dev)
bin/genmigrate add_firewall_tables

# Apply migration (production)
bin/migrate
```

## Generating Migrations (Dev)

```bash
bin/genmigrate [--schema <dir>] [--out <dir>] <name>
```

| Flag | Default | Description |
|------|---------|-------------|
| `--schema` | `pkg/storage/ent` | Ent schema root directory |
| `--out` | `deploy/migrations` | Migrations output directory |
| `<name>` | (required) | Migration name |

**Example:**
```bash
bin/genmigrate add_firewall_tables
```

**Output:**
```
==> bronze: atlas migrate diff add_firewall_tables
==> bronzehistory: atlas migrate diff add_firewall_tables
    rename: 20260208154545_add_firewall_tables.sql -> 0003_add_firewall_tables.sql
==> silver: atlas migrate diff add_firewall_tables
==> gold: atlas migrate diff add_firewall_tables

✅ Migration diff complete
```

Migration files use globally sequential numbering across all layers so versions don't collide in the shared `atlas_schema_revisions` table.

**Safety rule:** The configured `dbname` must end with `_dev`. Atlas drops and recreates tables during diff, so `genmigrate` refuses to run against a non-dev database.

## Applying Migrations (Production)

```bash
bin/migrate
```

Applies all pending migrations to the database. Only the target DB URL is needed.

The `migrate` binary embeds the SQL files from `deploy/migrations/` at build time, so it works as a single binary without needing the source tree.

## How It Works

Both tools:

1. Load database config from **Vault or YAML** (same as ingest)
2. Convert DSN to PostgreSQL URL format
3. Generate Atlas HCL config in memory with embedded URLs
4. Pipe the config to Atlas via a platform-specific mechanism (stdin on Unix, named pipe on Windows) — credentials never appear in CLI args or on disk

**`genmigrate` additionally:**
5. Safety-checks that `dbname` ends with `_dev` (Atlas drops and recreates tables during diff)
6. Runs `atlas migrate diff` for each layer: bronze → bronzehistory → silver → gold
7. Renames timestamp-based filenames to globally sequential numbering (0001_, 0002_, …)

**Security:** Credentials are passed to Atlas through a pipe (kernel memory only), not command-line arguments or files. This prevents passwords from appearing in `ps aux`, process lists, or the filesystem.

## Configuration

Both tools use the same config system as the ingest worker:

**YAML file (config.yaml):**
```yaml
# For genmigrate (dev) — dbname must end with _dev
database:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  dbname: hotpot_dev
  sslmode: disable

# For migrate (production)
database:
  host: prod-db.internal
  port: 5432
  user: hotpot
  password: secret
  dbname: hotpot
  sslmode: require
```

**CRITICAL:** `genmigrate` requires `dbname` to end with `_dev`. Atlas drops and recreates tables during diff, so the tool refuses to run against a production database. Use separate config files for dev and production.

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
    ├── embed.go          # Embeds SQL files into migrate binary
    ├── bronze/
    │   ├── 0001_initial.sql
    │   └── atlas.sum
    ├── bronzehistory/
    │   ├── 0002_initial.sql
    │   └── atlas.sum
    ├── silver/
    └── gold/
```

There is no `atlas.hcl` file — both tools generate the Atlas config in memory and pipe it directly to Atlas.

## Workflow

### After Schema Changes

1. **Modify ent schemas** in `pkg/schema/`
2. **Generate ent code**: `make generate`
3. **Generate migration**: `bin/genmigrate description`
4. **Review migration** files in `deploy/migrations/{layer}/`
5. **Apply migration**: `bin/migrate`

### Example

```bash
# 1. Edit schema
vim pkg/schema/bronze/gcp/compute/firewall.go

# 2. Generate ent code
make generate

# 3. Generate migration
bin/genmigrate add_gcp_compute_firewall

# 4. Review
cat deploy/migrations/bronze/0003_add_gcp_compute_firewall.sql

# 5. Apply
bin/migrate
```

## Setup

### Create Dev Database

Before running `genmigrate`, create the dev database:

```bash
psql -U postgres -c "CREATE DATABASE hotpot_dev;"
```

Then point your config at it (`dbname: hotpot_dev`).

## Platform Support

Both tools work on Linux, macOS, and Windows:

| Platform | Config delivery | Details |
|----------|----------------|---------|
| Linux/macOS | stdin (`/dev/stdin`) | Config piped via `cmd.Stdin` |
| Windows | Named pipe (`\\.\pipe\hotpot-atlas-<pid>`) | Config served via Win32 named pipe |

Both approaches keep credentials in kernel memory only — nothing touches disk or CLI args.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Database DSN not configured" | Set database config in YAML or Vault |
| "database name must end with _dev" | Set `dbname` to a `_dev` database (e.g. `hotpot_dev`) |
| "atlas: command not found" | Install Atlas CLI: `brew install ariga/tap/atlas` |
| Migration fails on one layer | Check `deploy/migrations/{layer}/` for conflicts |
| "incomplete DSN" | Ensure database.host, port, user, dbname are set |
| Dev database doesn't exist | Run `CREATE DATABASE hotpot_dev;` |
| "config pipe" error on Windows | Check that no other migrate/genmigrate process is running (pipe name collision) |
