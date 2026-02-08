# Configuration

Hotpot configuration via YAML file or HashiCorp Vault.

## Quick Start

```bash
# Copy example config
cp config.yaml.example config.yaml

# Replace ALL placeholders (search for <REPLACE_ and CHANGEME)
vim config.yaml

# Verify no placeholders remain
grep -E '<REPLACE_|CHANGEME' config.yaml && echo "⚠️  Found placeholders - replace them!"

# Run with file config
CONFIG_SOURCE=file CONFIG_FILE=config.yaml bin/ingest
```

**⚠️ Security Note:**
- `config.yaml.example` contains PLACEHOLDER values only
- Replace ALL `<PLACEHOLDER>` and `CHANGEME` values before use
- Never commit `config.yaml` to git (already in `.gitignore`)
- Use Vault for production credentials

## Configuration Sources

| Source | Use Case | How to Enable |
|--------|----------|---------------|
| YAML file | Development, simple deployments | `CONFIG_SOURCE=file CONFIG_FILE=config.yaml` |
| Vault | Production, secrets management | `CONFIG_SOURCE=vault VAULT_ADDR=... VAULT_TOKEN=...` |

## YAML Configuration

### Minimal Config

```yaml
# Minimum required fields
database:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  dbname: hotpot

temporal:
  host_port: localhost:7233
```

### Full Config

See `config.yaml.example` for complete configuration with all options.

## Environment Variables

### File Source

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CONFIG_SOURCE` | Yes | - | Set to `file` |
| `CONFIG_FILE` | Yes | - | Path to YAML file |

**Example:**
```bash
export CONFIG_SOURCE=file
export CONFIG_FILE=/path/to/config.yaml
bin/ingest
```

### Vault Source

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CONFIG_SOURCE` | Yes | - | Set to `vault` |
| `VAULT_ADDR` | Yes | - | Vault server address |
| `VAULT_TOKEN` | Yes | - | Vault authentication token |
| `VAULT_PATH` | Yes | - | Path to config in Vault |

**Example:**
```bash
export CONFIG_SOURCE=vault
export VAULT_ADDR=https://vault.example.com:8200
export VAULT_TOKEN=hvs.XXXXX
export VAULT_PATH=secret/hotpot/config
bin/ingest
```

## Configuration Fields

### GCP (Optional)

```yaml
gcp:
  credentials_json: |
    { ... }  # Service account JSON
  rate_limit_per_minute: 600  # Default: 600
```

**credentials_json:**
- **Optional** - Falls back to Application Default Credentials (ADC)
- Service account JSON with required permissions
- See [GCP Setup](../features/GCP.md) for required roles

**rate_limit_per_minute:**
- API requests per minute across all GCP clients
- Default: 600 (10 requests/second)
- Adjust based on quota limits

### Database (Required)

```yaml
database:
  host: localhost          # REQUIRED
  port: 5432              # REQUIRED
  user: postgres          # REQUIRED
  password: secret        # REQUIRED (use Vault in production)
  dbname: hotpot          # REQUIRED
  sslmode: require        # Optional - default: "require"
  dev_dbname: hotpot_dev  # Optional - default: "{dbname}_dev"
```

**sslmode options:**
- `disable` - No SSL (development only)
- `require` - SSL required (default)
- `verify-ca` - Verify certificate authority
- `verify-full` - Full certificate verification

**dev_dbname:**
- Scratch database for Atlas schema migrations
- Used by `bin/migrate diff` command
- Must be different from production database

### Temporal (Required)

```yaml
temporal:
  host_port: localhost:7233  # REQUIRED - no default
  namespace: default         # Optional - default: "default"
```

**host_port:**
- **REQUIRED** - Application fails if not set
- Temporal server address
- No default value (security: prevent accidental connections)

**namespace:**
- Optional - defaults to "default"
- Temporal namespace for workflows

### Redis (Optional)

```yaml
redis:
  address: localhost:6379  # REQUIRED if using Redis
  password: ""             # Optional
  db: 0                    # Optional - default: 0
```

**When to use:**
- Distributed rate limiting across multiple workers
- Production deployments with multiple instances

**Fallback:**
- If not configured, uses in-memory rate limiting (single process only)

## Hot Reload

Configuration changes are detected automatically:

**File source:**
- Watches `config.yaml` for changes
- Reloads on file modification
- Database connections gracefully reconnected

**Vault source:**
- Polls Vault for changes (interval configurable)
- Reloads on value changes

**What gets reloaded:**
- Database credentials
- GCP credentials
- Rate limits
- Temporal settings (on next workflow start)

**What doesn't reload:**
- Temporal namespace (requires restart)

## Validation

Configuration is validated on load and reload:

**Required fields checked:**
- `database.host`, `database.port`, `database.user`, `database.dbname`
- `temporal.host_port`

**Validation errors stop startup:**
```bash
$ bin/ingest
Failed to start: config validation failed: temporal.host_port is required
```

## Security Best Practices

| Practice | Reason |
|----------|--------|
| Use Vault in production | Never commit credentials to git |
| Set `sslmode: require` | Encrypt database connections |
| Rotate credentials regularly | Limit exposure window |
| Use separate dev database | Prevent accidental data loss |
| Use service accounts with minimal permissions | Principle of least privilege |

## Examples

### Development (File)

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: devpassword
  dbname: hotpot_dev
  sslmode: disable

temporal:
  host_port: localhost:7233
```

### Production (Vault)

**Vault path:** `secret/hotpot/prod/config`

**Vault data:**
```json
{
  "database": {
    "host": "prod-db.internal",
    "port": 5432,
    "user": "hotpot_prod",
    "password": "SECURE_PASSWORD_FROM_VAULT",
    "dbname": "hotpot",
    "sslmode": "verify-full"
  },
  "temporal": {
    "host_port": "temporal.internal:7233",
    "namespace": "production"
  },
  "gcp": {
    "rate_limit_per_minute": 1200
  },
  "redis": {
    "address": "redis.internal:6379",
    "password": "REDIS_PASSWORD_FROM_VAULT"
  }
}
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "config validation failed: temporal.host_port is required" | Set `temporal.host_port` in config |
| "database.host is required" | Ensure all required database fields are set |
| "Failed to connect to database" | Check database credentials and network access |
| Hot reload not working | Check file permissions, Vault connectivity |
| Rate limiting not working | Check Redis connection or use file-based config |

## References

- [MIGRATIONS.md](./MIGRATIONS.md) - Database migration setup
- [GCP.md](../features/GCP.md) - GCP service account setup
- [Vault Documentation](https://www.vaultproject.io/docs) - HashiCorp Vault
