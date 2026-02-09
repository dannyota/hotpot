# Workflows

Temporal workflow architecture for Hotpot ingestion.

## 📊 Hierarchy

```
GCPInventoryWorkflow              # All GCP resources, multiple projects
    ├── ComputeWorkflow           # Compute Engine, single project (orchestrator)
    │       ├── InstanceWorkflow  # Instances
    │       ├── DiskWorkflow      # Disks
    │       └── ...               # 10 resource workflows total
    ├── ContainerWorkflow         # GKE, single project (orchestrator)
    │       └── ClusterWorkflow   # Clusters
    └── ResourceManagerWorkflow   # Project discovery (orchestrator)
            └── ProjectWorkflow   # Projects
```

Each level can be triggered independently.

## ♻️ Client Lifecycle

GCP client lifetime = activity invocation (not worker lifetime).

```
InstanceWorkflow Start
    │
    ├── IngestInstances activity
    │       ├── createClient()        # new client per activity
    │       ├── ingest + save
    │       └── defer client.Close()  # closed when activity returns
    │
    └── Done
```

**Why activity-scoped:**
- Fresh credentials each invocation (picks up Vault/config changes)
- No shared state needed for single-activity workflows
- Retries can run on any worker (not pinned to one)
- No stale connections from long-running workers

## ▶️ Triggering Workflows

### GCPInventoryWorkflow

```
Caller ─► ExecuteWorkflow(GCPInventoryWorkflow, {ProjectIDs: [a,b,c]})
                │
                ▼
          GCPInventoryWorkflow
                │
                ├─► ComputeWorkflow(a) ─► 10 resource workflows
                ├─► ComputeWorkflow(b) ─► 10 resource workflows
                └─► ComputeWorkflow(c) ─► 10 resource workflows
```

### ComputeWorkflow

```
Caller ─► ExecuteWorkflow(ComputeWorkflow, {ProjectID: "a"})
                │
                ▼
          ComputeWorkflow(a)
                │
                ├─► InstanceWorkflow ─► IngestInstances activity
                ├─► DiskWorkflow ─► IngestDisks activity
                ├─► NetworkWorkflow ─► IngestNetworks activity
                └─► ... (10 total)
```

### InstanceWorkflow

```
Caller ─► ExecuteWorkflow(InstanceWorkflow, {ProjectID: "a"})
                │
                ▼
          InstanceWorkflow(a)
                │
                └─► IngestInstances activity (creates + closes own client)
```

## 🤔 When to Use

| Workflow | Use Case |
|----------|----------|
| `GCPInventoryWorkflow` | Scheduled full sync |
| `ComputeWorkflow` | Re-sync Compute after incident |
| `InstanceWorkflow` | Debug/test, on-demand refresh |

## 📋 Task Queues

| Provider | Queue |
|----------|-------|
| GCP | `hotpot-ingest-gcp` |
| VNG Cloud | `hotpot-ingest-vng` (future) |
| SentinelOne | `hotpot-ingest-s1` (future) |
| Fortinet | `hotpot-ingest-fortinet` (future) |
