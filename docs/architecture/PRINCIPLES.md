# Principles

Architecture principles for Hotpot development.

See [OVERVIEW.md](./OVERVIEW.md) for system design.

## 📦 1. Package Structure

```
pkg/ingest/{provider}/
├── config.go               # Provider config (credentials, etc.)
├── register.go             # Register workflows/activities
├── workflows.go            # Top-level workflow (e.g., GCPInventoryWorkflow)
└── {resource}/
    ├── client.go           # External API client
    ├── service.go          # Business logic (upsert, delete stale)
    ├── converter.go        # API response → Bronze model
    ├── activities.go       # Temporal activities (creates client)
    ├── workflows.go        # Resource workflow (e.g., InstanceWorkflow)
    └── register.go         # Register resource activities

pkg/schema/                 # Ent schemas (auto-discovered)
├── bronze/                 # Bronze schemas by provider/service
│   ├── mixin/
│   │   └── timestamp.go
│   └── gcp/
│       ├── compute/
│       │   ├── instance.go  # BronzeGCPComputeInstance + children
│       │   ├── disk.go
│       │   └── ...
│       └── ...
├── bronzehistory/          # History schemas (separate from bronze)
├── silver/                 # Silver schemas
└── gold/                   # Gold schemas

pkg/storage/ent/            # Generated code (DO NOT EDIT)
├── client.go               # Unified ent client
├── bronzegcpcomputeinstance.go
└── ...
```

## 🗄️ 2. Database Schemas

Use PostgreSQL schemas to separate layers:

```sql
bronze.gcp_instances      -- Raw data from GCP API
bronze.vng_servers        -- Raw data from VNG API
silver.assets             -- Unified asset model
silver.vulnerabilities    -- Unified vuln model
gold.compliance           -- Compliance results
gold.alerts               -- Security alerts
```

## 🔄 3. Data Flow

| Layer | Reads | Writes |
|-------|-------|--------|
| Ingest | External APIs | `bronze.*` |
| Normalize | `bronze.*` | `silver.*` |
| Detect | `silver.*` | `gold.*` |
| Metabase | all schemas | nothing |
| Agent | all schemas | nothing |

## 🚫 4. No Cross-Layer Imports

```go
// Wrong: importing another layer
import "hotpot/pkg/ingest/gcp"

// Correct: use generated ent client
import "hotpot/pkg/storage/ent"

instances, err := client.BronzeGCPComputeInstance.Query().All(ctx)
```

Layers communicate through database, not imports. Exception: `pkg/base/` can be imported by all layers.

## ♻️ 5. Activity Client Lifecycle

Activities create and close their own API client:

```go
func (a *Activities) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
    client, err := a.createClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("create client: %w", err)
    }
    defer client.Close()

    service := NewService(client, a.entClient)
    // ...
}
```

**Why:** Fresh credentials per activity invocation. No shared state needed for single-activity workflows. Retries can run on any worker.

**When to use sessions:** Only if a workflow runs multiple activities sharing an expensive resource on the same worker. No current workflows require this.

See [WORKFLOWS.md](../guides/WORKFLOWS.md) for details.

## 🏗️ 6. Activities Pattern

Activities use a struct to hold dependencies:

```go
// activities.go
type Activities struct {
    configService *config.Service
    db            *ent.Client
    limiter       *rate.Limiter
}

func NewActivities(configService *config.Service, db *ent.Client, limiter *rate.Limiter) *Activities {
    return &Activities{configService: configService, entClient: entClient, limiter: limiter}
}

// Activity params/results use dedicated structs
type IngestParams struct {
    ProjectID string
}

type IngestResult struct {
    InstanceCount int
}

func (a *Activities) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
    client, err := a.createClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("create client: %w", err)
    }
    defer client.Close()
    // ...
}
```

See [ACTIVITIES.md](../guides/ACTIVITIES.md) for details.

## 📋 7. Register Pattern

Each package has `register.go` to register workflows and activities:

```go
// pkg/ingest/gcp/compute/register.go
func Register(w worker.Worker, configService *config.Service, db *ent.Client, limiter *rate.Limiter) {
    instance.Register(w, configService, entClient, limiter)
    w.RegisterWorkflow(ComputeWorkflow)
}
```

## ⚙️ 8. Config Defaults

Defaults live in `config.Service` accessors, not in consumers:

```go
// Wrong: default in consumer (run.go)
hostPort := cfg.Temporal.HostPort
if hostPort == "" {
    hostPort = "localhost:7233"
}

// Correct: default in config service accessor
func (s *Service) TemporalHostPort() string {
    if s.config == nil || s.config.Temporal.HostPort == "" {
        return "localhost:7233"
    }
    return s.config.Temporal.HostPort
}
```

**Why:** Single source of truth for defaults, consumers don't duplicate logic.

| Field | Default |
|-------|---------|
| `TemporalHostPort` | `localhost:7233` |
| `TemporalNamespace` | `default` |
| `SSLMode` | `require` |
| `GCPRateLimitPerMinute` | `600` |

## 📐 9. Model Conventions

All models live in `pkg/base/models/{layer}/`.

**File organization** — group parent and child models in a single file named after the parent resource:

```
pkg/base/models/bronze/
├── gcp_compute_instance.go   # Instance + Disk + NIC + Label + Tag + ...
├── gcp_compute_disk.go       # Disk + Label + License
├── gcp_compute_network.go    # Network + Peering
├── gcp_compute_subnetwork.go # Subnetwork + SecondaryRange
└── gcp_project.go            # Project + Label
```

**Why:**
- Related models stay together, easier to understand relationships
- Mirrors API structure (GCP returns nested objects)
- Fewer files to navigate
- Changes to parent often require changes to children

**Model order within file:**
1. Parent schema with `Fields()`, `Edges()`, `Annotations()` methods
2. Child schemas alphabetically, each with same methods

See [ENT_SCHEMAS.md](../guides/ENT_SCHEMAS.md) for ent schema patterns and [CODE_STYLE.md](../guides/CODE_STYLE.md) for field conventions.

## 📜 10. History Tables (SCD Type 4)

Separate `*_history` packages for change tracking with granular time ranges:

```go
// Current: pkg/schema/bronze/gcp/compute/
package compute

type BronzeGCPComputeInstance struct {
    ent.Schema
}

func (BronzeGCPComputeInstance) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").StorageKey("resource_id").Immutable(),  // GCP API ID
        // ...
    }
}

func (BronzeGCPComputeInstance) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{Table: "gcp_compute_instances"},
    }
}
```

```go
// History: pkg/schema/bronzehistory/gcp/compute/
package compute

type BronzeHistoryGCPComputeInstance struct {
    ent.Schema
}

func (BronzeHistoryGCPComputeInstance) Fields() []ent.Field {
    return []ent.Field{
        field.Uint("history_id").Unique().Immutable(),
        field.String("resource_id").NotEmpty(),
        field.Time("valid_from").Immutable(),
        field.Time("valid_to").Optional().Nillable(),  // NULL = current
        // ... same fields as bronze
    }
}

func (BronzeHistoryGCPComputeInstance) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{Table: "gcp_compute_instances_history"},
    }
}
```

**All levels have `valid_from/valid_to`** for granular change tracking:

```go
// Child history: own time range (can change independently)
type BronzeHistoryGCPComputeInstanceNIC struct {
    ent.Schema
}

func (BronzeHistoryGCPComputeInstanceNIC) Fields() []ent.Field {
    return []ent.Field{
        field.Uint("history_id").Unique().Immutable(),
        field.Uint("instance_history_id"),
        field.Time("valid_from").Immutable(),  // NIC's own time range
        field.Time("valid_to").Optional().Nillable(),
        // ...
    }
}
```

| Schema | Package | Purpose |
|--------|---------|---------|
| `bronze` | `pkg/schema/bronze/` | Current state |
| `bronze_history` | `pkg/schema/bronzehistory/` | All versions with time ranges |

See [HISTORY.md](./HISTORY.md) for details.

## 🥉 11. Bronze Data Design

Bronze stores API responses with minimal transformation. Two storage options:

| API Data Type | Storage | Example |
|---------------|---------|---------|
| Scalar fields (top-level) | Columns | `name`, `status`, `endpoint` |
| Unsigned integers (`uint64`, `uint32`) | String column | `id` → `resource_id varchar(255)` |
| Arrays of objects | Separate table | `nodePools[]` → `cluster_node_pools` |
| Maps (key-value) | Separate table | `labels` → `cluster_labels` |
| Nested objects | JSONB column | `privateClusterConfig` → `private_cluster_config_json` |
| Arrays of primitives | JSONB column | `users[]` → `users_json` |

**Rule: Tables or JSONB, never extract nested fields as columns.** Use JSONB (`type:jsonb`) for any JSON data not stored in a separate table. Query with PostgreSQL JSON operators if needed.

```
# Wrong: extracting nested fields as columns
enable_private_nodes    ← from privateClusterConfig.enablePrivateNodes
master_ipv4_cidr_block  ← from privateClusterConfig.masterIpv4CidrBlock

# Correct: store entire nested object as JSONB
private_cluster_config_json JSONB  ← entire privateClusterConfig object
```

**Rule: Store unsigned integers as strings.** PostgreSQL has no unsigned integer types — `bigint` (signed int64) overflows for large `uint64`, and `integer` (signed int32) overflows for large `uint32`. Store as `string` (`varchar(255)`) and convert via `fmt.Sprintf("%d", value)` in the converter.

**Separate table** — use for top-level arrays and maps:
- Arrays of objects: `nodePools[]` → `cluster_node_pools` table
- Maps: `resourceLabels` → `cluster_labels` table (key, value columns)

**Nested arrays/maps** — judgment call based on query needs:
- If queryable via parent table link, store in JSONB for audit and completeness
- Example: `nodePool.config.taints[]` → stays in `config_json` (can join to node_pool for queries)
- Create separate table only if direct querying is required and parent link isn't sufficient

**JSONB column** — use for nested objects and primitive arrays:
- Preserves raw API structure
- Query with JSON operators if needed: `WHERE config_json->>'enabled' = 'true'`
- No need to update schema when API adds fields

See [CODE_STYLE.md](../guides/CODE_STYLE.md#jsonb-fields) for implementation conventions.

## 🔗 12. Cross-Layer References

Layers are loosely coupled. No FK constraints between layers:

```
Within layer:  Surrogate FK OK (bronze.disk → bronze.instance)
Across layers: Business key only (silver.source_id stores resource_id, not bronze.id)
```

| Approach | Issue |
|----------|-------|
| FK CASCADE | Bronze delete cascades to silver → data loss |
| FK RESTRICT | Blocks bronze delete |
| FK SET NULL | Silver loses reference |
| **No FK** | Layers independent, each owns its data |

**Why:**
- Bronze is current state, can be re-ingested anytime
- Silver transforms bronze but doesn't depend on it existing
- Each layer has its own retention policy
- Deleting bronze doesn't affect silver or gold

```go
// Silver references bronze by business key, not surrogate ID
type Asset struct {
    field.String("id").Unique().Immutable(),
    field.String("source_type"),       // "gcp_compute_instance"
    field.String("source_id"),         // resource_id (API ID), not bronze.id
}
```

## 🚦 13. Rate Limiting

External API clients share a per-provider rate limiter (`pkg/base/ratelimit`).
A `ratelimit.Service` is created per provider in `Register()`, returning a
`ratelimit.Limiter` interface passed down to activities.

**Backend priority:** Redis (distributed, per-second INCR counter) → local
`x/time/rate` (fallback). Temporal `TaskQueueActivitiesPerSecond` is always
set as a server-side safety net (`rate_limit_per_minute / 60`).

Three integration methods — all accept the `ratelimit.Limiter` interface:

| Method | When to use |
|--------|------------|
| `limiter.Wait(ctx)` | SDK-agnostic fallback — call before each API request |
| `NewRateLimitedTransport(limiter, base)` | SDK accepts custom `http.Client` (REST) |
| `UnaryInterceptor(limiter)` | SDK accepts `grpc.DialOption` (gRPC) |

Choose the appropriate method per client type. Prefer transport/interceptor
injection when the SDK supports it — keeps client code clean.

Config: `rate_limit_per_minute` per provider (default 600), `redis.address`
for distributed limiting.
