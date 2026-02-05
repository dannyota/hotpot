# Principles

Architecture principles for Hotpot development.

See [OVERVIEW.md](./OVERVIEW.md) for system design.

## 1. Package Structure

```
pkg/ingest/{provider}/
├── config.go               # Provider config (credentials, etc.)
├── register.go             # Register workflows/activities
├── workflows.go            # Top-level workflow (e.g., GCPInventoryWorkflow)
└── {resource}/
    ├── client.go           # External API client
    ├── session.go          # Session-based client management
    ├── service.go          # Business logic (upsert, delete stale)
    ├── converter.go        # API response → Bronze model
    ├── activities.go       # Temporal activities
    ├── workflows.go        # Resource workflow (e.g., InstanceWorkflow)
    └── register.go         # Register resource activities

pkg/base/models/            # Shared models (all layers import from here)
├── bronze/                 # Bronze models by domain
│   ├── gcp_compute_instance.go
│   ├── gcp_compute_disk.go
│   └── ...
├── silver/                 # Silver models (assets.go, vulns.go...)
└── gold/                   # Gold models (alerts.go, compliance.go...)
```

## 2. Database Schemas

Use PostgreSQL schemas to separate layers:

```sql
bronze.gcp_instances      -- Raw data from GCP API
bronze.vng_servers        -- Raw data from VNG API
silver.assets             -- Unified asset model
silver.vulnerabilities    -- Unified vuln model
gold.compliance           -- Compliance results
gold.alerts               -- Security alerts
```

## 3. Data Flow

| Layer | Reads | Writes |
|-------|-------|--------|
| Ingest | External APIs | `bronze.*` |
| Normalize | `bronze.*` | `silver.*` |
| Detect | `silver.*` | `gold.*` |
| Metabase | all schemas | nothing |
| Agent | all schemas | nothing |

## 4. No Cross-Layer Imports

```go
// Wrong: importing another layer
import "hotpot/pkg/ingest/gcp"

// Correct: import shared models from base/
import "hotpot/pkg/base/models/bronze"

var instances []bronze.GCPInstance
db.Find(&instances)
```

Layers communicate through database, not imports. Exception: `pkg/base/` can be imported by all layers.

## 5. Session-Based Client Pattern

External API clients live for workflow duration, not worker lifetime:

```go
// workflow creates session
sess, _ := workflow.CreateSession(ctx, sessionOpts)
sessionID := workflow.GetSessionInfo(sess).SessionID

defer func() {
    workflow.ExecuteActivity(sess, CloseSessionClientActivity, sessionID)
    workflow.CompleteSession(sess)
}()

// activities use session to get/create client
client, _ := GetOrCreateClient(ctx, sessionID, credentialsFile)
```

**Why:** Fresh credentials each workflow, picks up config file changes.

See [WORKFLOWS.md](../guides/WORKFLOWS.md) for details.

## 6. Activities Pattern

Activities use a struct to hold dependencies:

```go
// activities.go
type Activities struct {
    credentialsFile string
    db              *gorm.DB
}

func NewActivities(credentialsFile string, db *gorm.DB) *Activities {
    return &Activities{credentialsFile: credentialsFile, db: db}
}

// Activity params/results use dedicated structs
type IngestParams struct {
    SessionID string
    ProjectID string
}

type IngestResult struct {
    InstanceCount int
}

func (a *Activities) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
    client, err := GetOrCreateSessionClient(ctx, params.SessionID, a.credentialsFile)
    if err != nil {
        return nil, fmt.Errorf("get client: %w", err)
    }
    // ...
}
```

See [ACTIVITIES.md](../guides/ACTIVITIES.md) for details.

## 7. Register Pattern

Each package has `register.go` to register workflows and activities:

```go
// pkg/ingest/gcp/compute/register.go
func Register(w worker.Worker, credentialsFile string, db *gorm.DB) {
    instance.Register(w, credentialsFile, db)
    w.RegisterWorkflow(ComputeWorkflow)
}
```

Worker requires `EnableSessionWorker: true` for session support.

## 8. Model Conventions

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
1. Parent model with `TableName()` method
2. Child models alphabetically, each with `TableName()` method

**Bronze models** — document original API field in `json` tag for traceability:

```go
// pkg/base/models/bronze/gcp_compute_instance.go
package bronze

type GCPComputeInstance struct {
    ID uint `gorm:"primaryKey"`

    // gorm = our name + type, json = original API field
    ResourceID  string `gorm:"column:resource_id;type:varchar(255);uniqueIndex" json:"id"`
    Name        string `gorm:"column:name;type:varchar(255);not null" json:"name"`
    MachineType string `gorm:"column:machine_type;type:text" json:"machineType"`
    Status      string `gorm:"column:status;type:varchar(50);index" json:"status"`

    // Collection metadata (not from API)
    ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
    CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

    // Relationships
    Disks []GCPComputeInstanceDisk `gorm:"foreignKey:InstanceID;constraint:OnDelete:CASCADE"`
}

func (GCPComputeInstance) TableName() string {
    return "bronze.gcp_compute_instances"
}
```

**Silver/Gold models** — `json` tag matches column name (for API responses):

```go
// pkg/base/models/silver/assets.go
package silver

type Asset struct {
    ID       string `gorm:"column:id" json:"id"`
    Name     string `gorm:"column:name" json:"name"`
    Type     string `gorm:"column:type" json:"type"`
    SourceID string `gorm:"column:source_id" json:"source_id"`  // FK to bronze
}
```

| Layer | `json` tag purpose |
|-------|-------------------|
| Bronze | Original external API field name |
| Silver/Gold | API response field (matches column) |

## 9. History Tables (SCD Type 4)

Separate `*_history` packages for change tracking with granular time ranges:

```go
// Current: pkg/base/models/bronze/
package bronze

type GCPComputeInstance struct {
    ResourceID string `gorm:"primaryKey"`  // GCP API ID
    // ...
}

func (GCPComputeInstance) TableName() string {
    return "bronze.gcp_compute_instances"
}
```

```go
// History: pkg/base/models/bronze_history/
package bronze_history

type GCPComputeInstance struct {
    HistoryID  uint       `gorm:"primaryKey"`
    ResourceID string     `gorm:"index"`
    ValidFrom  time.Time  `gorm:"not null"`
    ValidTo    *time.Time                       // NULL = current
    // ... same fields
}

func (GCPComputeInstance) TableName() string {
    return "bronze_history.gcp_compute_instances"
}
```

**All levels have `valid_from/valid_to`** for granular change tracking:

```go
// Child history: own time range (can change independently)
type GCPComputeInstanceNIC struct {
    HistoryID         uint       `gorm:"primaryKey"`
    InstanceHistoryID uint       `gorm:"index"`
    ValidFrom         time.Time  `gorm:"not null"`  // NIC's own time range
    ValidTo           *time.Time
    // ...
}
```

| Schema | Package | Purpose |
|--------|---------|---------|
| `bronze` | `bronze/` | Current state |
| `bronze_history` | `bronze_history/` | All versions with time ranges |

See [HISTORY.md](./HISTORY.md) for details.

## 10. Bronze Data Design

Bronze stores API responses with minimal transformation. Two storage options:

| API Data Type | Storage | Example |
|---------------|---------|---------|
| Scalar fields (top-level) | Columns | `name`, `status`, `endpoint` |
| Arrays | Separate table | `nodePools[]` → `cluster_node_pools` |
| Maps (key-value) | Separate table | `labels` → `cluster_labels` |
| Nested objects | JSONB column | `privateClusterConfig` → `private_cluster_config_json` |

**Rule: Tables or JSONB, never extract nested fields as columns.**

Don't extract fields from nested objects into parent columns—it's confusing and breaks traceability. Keep nested objects as JSONB; if you need to query them, use PostgreSQL JSON operators.

```
# Wrong: extracting nested fields as columns
enable_private_nodes    ← from privateClusterConfig.enablePrivateNodes
master_ipv4_cidr_block  ← from privateClusterConfig.masterIpv4CidrBlock

# Correct: store entire nested object as JSONB
private_cluster_config_json JSONB  ← entire privateClusterConfig object
```

**Separate table** — use for top-level arrays and maps:
- Arrays of objects: `nodePools[]` → `cluster_node_pools` table
- Maps: `resourceLabels` → `cluster_labels` table (key, value columns)

**Nested arrays/maps** — judgment call based on query needs:
- If queryable via parent table link, store in JSONB for audit and completeness
- Example: `nodePool.config.taints[]` → stays in `config_json` (can join to node_pool for queries)
- Create separate table only if direct querying is required and parent link isn't sufficient

**JSONB column** — use for nested config objects:
- Preserves raw API structure
- Query with JSON operators if needed: `WHERE config_json->>'enabled' = 'true'`
- No need to update schema when API adds fields

**Example:**

```go
// API: { "name": "x", "labels": {...}, "nodePools": [...], "privateClusterConfig": {...} }

type GCPContainerCluster struct {
    // Top-level scalars → columns
    Name string `gorm:"column:name" json:"name"`

    // Nested object → JSONB (not extracted as columns)
    PrivateClusterConfigJSON string `gorm:"column:private_cluster_config_json;type:jsonb" json:"privateClusterConfig"`

    // Arrays/maps → separate tables
    Labels    []GCPContainerClusterLabel    `gorm:"foreignKey:ClusterResourceID"`
    NodePools []GCPContainerClusterNodePool `gorm:"foreignKey:ClusterResourceID"`
}
```

## 11. Cross-Layer References

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
    ID         string `gorm:"primaryKey"`
    SourceType string `gorm:"column:source_type"`       // "gcp_compute_instance"
    SourceID   string `gorm:"column:source_id;index"`   // resource_id (API ID), not bronze.id
}
```
