# Workflows

Temporal workflow architecture for Hotpot ingestion.

## Hierarchy

```
GCPInventoryWorkflow          # All GCP resources, multiple projects
    └── ComputeWorkflow       # Compute Engine, single project (owns session)
            └── activities    # Instance, Disk, Network ingestion
```

Each level can be triggered independently.

## Client Lifecycle

GCP client lifetime = workflow session lifetime (not worker lifetime).

```
ComputeWorkflow Start
    │
    ├── CreateSession()              # Temporal session created
    │
    ├── IngestInstances activity     # GetOrCreateClient(sessionID)
    │       └── uses session client      ├── first call: creates client
    │                                    └── reuses for session
    ├── IngestDisks activity         # same session, same client
    │
    ├── CloseSessionClient activity  # CloseSessionClient(sessionID)
    │
    └── CompleteSession()
```

**Why session-based:**
- Fresh credentials each workflow (picks up Vault/config changes)
- Shared client across activities within workflow (efficient)
- Clean boundary - workflow = client lifetime
- No stale connections from long-running workers

## Triggering Workflows

### GCPInventoryWorkflow

```
Caller ─► ExecuteWorkflow(GCPInventoryWorkflow, {ProjectIDs: [a,b,c]})
                │
                ▼
          GCPInventoryWorkflow
                │
                ├─► ComputeWorkflow(a) [session] ─► activities ─► cleanup
                ├─► ComputeWorkflow(b) [session] ─► activities ─► cleanup
                └─► ComputeWorkflow(c) [session] ─► activities ─► cleanup
```

### ComputeWorkflow

```
Caller ─► ExecuteWorkflow(ComputeWorkflow, {ProjectID: "a"})
                │
                ▼
          ComputeWorkflow(a)
                │
                ├── CreateSession
                ├─► IngestInstances (session client)
                ├─► IngestDisks (session client, future)
                ├─► IngestNetworks (session client, future)
                ├── CloseSessionClient
                └── CompleteSession
```

### InstanceWorkflow

```
Caller ─► ExecuteWorkflow(InstanceWorkflow, {ProjectID: "a"})
                │
                ▼
          InstanceWorkflow(a)
                │
                ├── CreateSession
                ├─► IngestInstances (session client)
                ├── CloseSessionClient
                └── CompleteSession
```

## When to Use

| Workflow | Use Case |
|----------|----------|
| `GCPInventoryWorkflow` | Scheduled full sync |
| `ComputeWorkflow` | Re-sync Compute after incident |
| `InstanceWorkflow` | Debug/test, on-demand refresh |

## Task Queues

| Provider | Queue |
|----------|-------|
| GCP | `hotpot-ingest-gcp` |
| VNG Cloud | `hotpot-ingest-vng` (future) |
| SentinelOne | `hotpot-ingest-s1` (future) |
| Fortinet | `hotpot-ingest-fortinet` (future) |

## Session Management

Session clients stored in `sync.Map` keyed by session ID:

```
pkg/ingest/gcp/compute/instance/session.go
    │
    ├── GetOrCreateSessionClient(sessionID, configService)
    │       └── creates client on first call, reuses for session
    │
    └── CloseSessionClient(sessionID)
            └── closes and removes client from map
```

Worker must enable sessions: `worker.Options{EnableSessionWorker: true}`
