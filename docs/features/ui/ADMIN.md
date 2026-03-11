# Admin

Built-in web interface for viewing Hotpot data. Replaces Metabase.

## 🎯 Overview

```
  Browser ──► Admin (Go HTTP + embed.FS) ──► PostgreSQL
                │                               │
                ├── /api/v1/*  (JSON API)        ├── bronze.*
                └── /* (Vue 3 SPA)               ├── silver.*
                                                 └── gold.*
```

Single Go binary serves Vue 3 frontend via `embed.FS`. Read-only access to all layers.

## 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go net/http (1.22+ routing) |
| Frontend | Vue 3 + TypeScript + Vite |
| Components | Custom `H*` components + reka-ui (tooltips) |
| Styling | Tailwind CSS (zinc palette) |
| Embedding | Go embed.FS (single binary) |

## 📦 Package Structure

| Package | Purpose |
|---------|---------|
| `cmd/admin/` | Binary entry point |
| `pkg/admin/` | HTTP server, registry, response helpers |
| `pkg/admin/query/` | URL params parsing + validation |
| `pkg/admin/listhandler/` | Config-driven paginated list handler |
| `pkg/admin/{layer}/{provider}/{service}/` | Per-entity route registration |
| `admin/ui/` | Vue 3 frontend source |

## 🖥️ Pages

| Page | Route | API Endpoint | Status |
|------|-------|-------------|:------:|
| Dashboard | `/` | `GET /api/v1/stats/overview` | ⬜ |
| GCP Instances | `/bronze/gcp/compute/instances` | `GET /api/v1/bronze/gcp/compute/instances` | ⬜ |
| S1 Agents | `/bronze/s1/agents` | `GET /api/v1/bronze/s1/agents` | ⬜ |
| Vault PKI Certificates | `/bronze/vault/pki/certificates` | `GET /api/v1/bronze/vault/pki/certificates` | ⬜ |
| API Catalog Endpoints | `/bronze/apicatalog/endpoints` | `GET /api/v1/bronze/apicatalog/endpoints` | ⬜ |
| Machines | `/inventory/machines` | `GET /api/v1/inventory/machines` | ⬜ |
| Software EOL | `/gold/lifecycle/software` | `GET /api/v1/gold/lifecycle/software` | ⬜ |

## 🧩 Base Components

| Component | Responsibility |
|-----------|---------------|
| `HDataTable` | Paginated sortable table with column slots, row click, page size selector |
| `HMultiSelect` | Multi-select dropdown with option counts from API |
| `HDetailDrawer` | Slide-in panel from right for row detail view |
| `HJsonPeek` | Inline JSON preview with tooltip + click to expand |
| `HJsonViewer` | Full-screen JSON modal with copy button |
| `HAlert` | Inline alert banner (error/warning/info) |
| `AppNotifications` | Bell icon dropdown in topbar, persistent notification history |

## 🔧 Composables

| Composable | Responsibility |
|------------|---------------|
| `useTableQuery` | Wires pagination + sort + filter state to `useListApi` |
| `useListApi` / `useApi` | Fetch wrappers with error typing + auto-notification |
| `exportCsv` | Fetches all filtered rows and downloads as CSV |
| `useTimezone` | Date formatting with timezone preference |
| `useNotifications` | Notification store backed by localStorage (30-day expiry) |
| `columnLabel` | Maps field names to display labels |

## 🔌 Go API

Routes are registered via `admin.RegisterRoute()` in per-entity `register.go` files. Two handler types:

| Type | Function | Usage |
|------|----------|-------|
| Ent-backed list | `lh.Handler(lh.Config{...})` | Dedicated pages with typed queries |
| Raw SQL list | `lh.RegisterSQL(db, tables)` | Generic read-only tables |

All list endpoints return the standard envelope: `{ data, meta, filter_options }`.

See `docs/guides/ADMIN_PAGES.md` for the full guide on adding new pages.

## 📊 Dashboard Stats API

`GET /api/v1/stats/overview` powers the dashboard. Returns counts across all three layers with optional blueprint delta comparison.

### Response Structure

```json
{
  "data": {
    "bronze": {
      "gcp": { "resources": [{ "label": "Compute Instances", "count": 677, "delta": 12 }, ...] },
      "greennode": { "resources": [...] },
      "s1": { "resources": [...] },
      "meec": { "resources": [...] },
      "vault": { "resources": [...] },
      "apicatalog": { "resources": [...] }
    },
    "silver": {
      "machines": { "count": 1444, "delta": 23 },
      "k8s_nodes": { "count": 197 },
      "software": { "count": 47793 },
      "api_endpoints": { "count": 0 },
      "traffic_5m": { "count": 0 },
      "client_ips_5m": { "count": 0 },
      "user_agents_5m": { "count": 0 }
    },
    "gold": {
      "software_eol": { "count": 89, "delta": 3 },
      "software_eoes": { "count": 38 },
      "os_eol": { "count": 31 },
      "os_eoes": { "count": 12 },
      "anomalies": { "count": 7 },
      "anomalies_critical": { "count": 2 },
      "anomalies_high": { "count": 3 },
      "anomalies_medium": { "count": 2 }
    }
  }
}
```

### Bronze Provider Definitions

Each provider lists highlighted resources with their bronze table name:

| Provider | Key | Table | Label |
|----------|-----|-------|-------|
| GCP | `gcp` | `gcp_compute_instances` | Compute Instances |
| | | `gcp_container_clusters` | GKE Clusters |
| | | `gcp_compute_disks` | Disks |
| | | `gcp_compute_snapshots` | Snapshots |
| | | `gcp_compute_firewalls` | Firewalls |
| | | `gcp_storage_buckets` | Storage Buckets |
| GreenNode | `greennode` | `greennode_compute_servers` | Servers |
| | | `greennode_network_vpcs` | VPCs |
| | | `greennode_network_secgroups` | Security Groups |
| | | `greennode_volume_block_volumes` | Block Volumes |
| SentinelOne | `s1` | `s1_agents` | Agents |
| | | `s1_app_inventory` | App Inventory |
| | | `s1_network_discoveries` | Network Discoveries |
| MEEC | `meec` | `meec_inventory_computers` | Computers |
| | | `meec_inventory_installed_software` | Installed Software |
| Vault | `vault` | `vault_pki_certificates` | PKI Certificates |
| API Catalog | `apicatalog` | `apicatalog_endpoints_raw` | API Endpoints |

### Silver Keys

| Key | Table | Description |
|-----|-------|-------------|
| `machines` | `silver.inventory_machines` | Unified machine inventory |
| `k8s_nodes` | `silver.inventory_k8s_nodes` | Unified K8s nodes |
| `software` | `silver.inventory_software` | Unified software inventory |
| `api_endpoints` | `silver.inventory_api_endpoints` | API endpoint catalog |
| `traffic_5m` | `silver.httptraffic_traffic_5m` | HTTP traffic (5-min windows) |
| `client_ips_5m` | `silver.httptraffic_client_ip_5m` | Client IPs (5-min windows) |
| `user_agents_5m` | `silver.httptraffic_user_agent_5m` | User agents (5-min windows) |

### Gold Keys

| Key | Query | Description |
|-----|-------|-------------|
| `software_eol` | `WHERE eol_status = 'eol_expired'` | Software past end-of-life |
| `software_eoes` | `WHERE eol_status = 'eoes_expired'` | Software past end-of-support |
| `os_eol` | `WHERE eol_status = 'eol_expired'` | OS past end-of-life |
| `os_eoes` | `WHERE eol_status = 'eoes_expired'` | OS past end-of-support |
| `anomalies` | `COUNT(*)` | Total anomalies |
| `anomalies_critical` | `WHERE severity = 'critical'` | Critical anomalies |
| `anomalies_high` | `WHERE severity = 'high'` | High anomalies |
| `anomalies_medium` | `WHERE severity = 'medium'` | Medium anomalies |

### Delta Semantics

Dashboard-only feature. Compares current counts against the last saved blueprint snapshot.

| Value | Display | Color |
|-------|---------|-------|
| `null` | Hidden | — |
| `0` | Hidden | — |
| `> 0` | `+N` | Green (bronze/silver), Red (gold) |
| `< 0` | `-N` | Red (bronze/silver), Green (gold) |

Gold layer inverts colors: more issues = bad (red), fewer issues = good (green). Deltas appear only on the dashboard stats overview, not on individual table pages.

### Backend Location

`pkg/admin/stats/register.go` — `bronzeProviderDefs`, `querySilver()`, `queryGold()`.

## ⚠️ Error Handling

### Backend — Never Expose Internal Errors

API handlers **never** send raw error details to the client. Use `WriteServerError` for all 500s:

```go
// Good — logs real error server-side, sends generic message to client
admin.WriteServerError(w, "failed to load instances", err)
// Server log: level=ERROR msg="failed to load instances" error="dial tcp 127.0.0.1:5432: connect: connection refused"
// Client sees: {"error":{"code":500,"message":"failed to load instances"}}

// Bad — leaks DB connection details, hostnames, table names
admin.WriteError(w, 500, "count instances: "+err.Error())
```

| Function | Use for | Logs | Client sees |
|----------|---------|------|-------------|
| `WriteServerError(w, msg, err)` | 500 errors (DB, internal) | `slog.Error(msg, "error", err)` | Generic `msg` only |
| `WriteError(w, 400, msg)` | 400 errors (bad input) | Nothing | Validation message |

**Rule:** `WriteError` with status 500 must **never** include `err.Error()` in the message.

### Frontend — Two-Level Error Display

Errors surface in two places so users always know what happened:

| Level | Component | Where | Behavior |
|-------|-----------|-------|----------|
| Inline | `HAlert` via `HDataTable :error` | In the table area | Shows instead of "No data found" when API fails |
| Persistent | `AppNotifications` | Topbar bell icon | Stores all errors in localStorage, 30-day expiry |

**For table pages** — pass `error` from `useTableQuery` to `HDataTable`:

```vue
<script setup>
const { data, meta, loading, error, ... } = useTableQuery({ endpoint: '...' })
</script>

<HDataTable :data="data" :meta="meta" :loading="loading" :error="error" ... />
```

**For non-table pages** — use `HAlert` directly:

```vue
<script setup>
import HAlert from '@/components/app/HAlert.vue'
const { data, error, fetch } = useApi('/api/v1/stats/overview')
</script>

<HAlert v-if="error" type="error" :message="error" />
```

`HAlert` supports three variants: `error` (red), `warning` (amber), `info` (blue). Optional `dismissible` prop adds a close button.

### Notification System

`useNotifications` composable manages a persistent notification store:

| Feature | Detail |
|---------|--------|
| Storage | `localStorage` key `hotpot-notifications` |
| Expiry | Auto-prune entries older than 30 days |
| Dedup | Same source + message within 5 minutes = skipped |
| Actions | Mark read, mark all read, dismiss, clear all |
| Auto-capture | `useApi` and `useListApi` call `add('error', ...)` on every API failure |

## 🔒 Access Control

| Concern | Approach |
|---------|----------|
| Authentication | JWT / session (planned) |
| Authorization | Role-based route guard (planned) |
| Audit | Structured slog per request (planned) |
| Data Access | Read-only Ent queries, no mutations |

## ⚙️ Configuration

```yaml
admin:
  addr: ":8080"  # Default: :8080
```

## 📋 Build

```bash
make admin-dev    # Dev mode (Vite HMR + Go)
make build-admin  # Production (Vite build → Go embed → single binary)
```

