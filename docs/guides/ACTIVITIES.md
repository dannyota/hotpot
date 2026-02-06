# Activities

Temporal activities for data ingestion and processing.

```mermaid
flowchart LR
    Workflow --> Activity
    Activity --> Session[Session Client]
    Activity --> Service
    Service --> DB[(PostgreSQL)]
    Service --> History[History Service]
```

## File Structure

```
pkg/ingest/{provider}/{resource}/
├── client.go        # External API client wrapper
├── session.go       # Session-based client management
├── converter.go     # API response → Bronze model
├── diff.go          # Change detection between old/new
├── history.go       # SCD Type 4 history tracking
├── service.go       # Business logic (CRUD + history)
├── activities.go    # Activity struct + methods
├── workflows.go     # Workflow calling activities
└── register.go      # Register with Temporal worker
```

## Activity Struct

Hold dependencies, not state:

```go
type Activities struct {
    configService *config.Service
    db            *gorm.DB
}

func NewActivities(configService *config.Service, db *gorm.DB) *Activities {
    return &Activities{configService: configService, db: db}
}
```

## Params/Result Structs

Dedicated structs for each activity:

```go
// Params: inputs to activity
type IngestComputeInstancesParams struct {
    SessionID string  // Required for session-based client
    ProjectID string
}

// Result: outputs from activity
type IngestComputeInstancesResult struct {
    ProjectID      string
    InstanceCount  int
    DurationMillis int64
}
```

## Activity Function Reference

Export a variable for workflow registration:

```go
// Function reference for workflow ExecuteActivity
var IngestComputeInstancesActivity = (*Activities).IngestComputeInstances
```

**Why:** Allows type-safe activity execution in workflows.

## Activity Method

```go
func (a *Activities) IngestComputeInstances(ctx context.Context, params IngestComputeInstancesParams) (*IngestComputeInstancesResult, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Starting ingestion", "projectID", params.ProjectID)

    // 1. Get session client
    client, err := GetOrCreateSessionClient(ctx, params.SessionID, a.configService)
    if err != nil {
        return nil, fmt.Errorf("get client: %w", err)
    }

    // 2. Create service and execute
    service := NewService(client, a.db)
    result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
    if err != nil {
        return nil, fmt.Errorf("ingest: %w", err)
    }

    // 3. Cleanup (delete stale)
    if err := service.DeleteStaleInstances(ctx, params.ProjectID, result.CollectedAt); err != nil {
        logger.Warn("Failed to delete stale", "error", err)
    }

    // 4. Return result
    return &IngestComputeInstancesResult{
        ProjectID:      result.ProjectID,
        InstanceCount:  result.InstanceCount,
        DurationMillis: result.DurationMillis,
    }, nil
}
```

## Session Client

Client lives for workflow duration, not worker lifetime. Credentials priority: Vault JSON > ADC.

```go
var sessionClients sync.Map

func GetOrCreateSessionClient(ctx context.Context, sessionID string, configService *config.Service) (*Client, error) {
    if client, ok := sessionClients.Load(sessionID); ok {
        return client.(*Client), nil
    }

    var opts []option.ClientOption
    if credJSON := configService.GCPCredentialsJSON(); len(credJSON) > 0 {
        opts = append(opts, option.WithCredentialsJSON(credJSON))
    }
    // If empty, uses Application Default Credentials (ADC)

    client, err := NewClient(ctx, opts...)
    if err != nil {
        return nil, err
    }
    sessionClients.Store(sessionID, client)
    return client, nil
}

func CloseSessionClient(sessionID string) {
    if client, ok := sessionClients.LoadAndDelete(sessionID); ok {
        client.(*Client).Close()
    }
}
```

**Cleanup activity:**

```go
var CloseSessionClientActivity = (*Activities).CloseSessionClient

func (a *Activities) CloseSessionClient(ctx context.Context, params CloseSessionClientParams) error {
    CloseSessionClient(params.SessionID)
    return nil
}
```

## Registration

```go
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
    activities := NewActivities(configService, db)
    w.RegisterActivity(activities.IngestComputeInstances)
    w.RegisterActivity(activities.CloseSessionClient)
    w.RegisterWorkflow(InstanceWorkflow)
}
```

## Workflow Calling

```go
func InstanceWorkflow(ctx workflow.Context, params InstanceWorkflowParams) (*InstanceWorkflowResult, error) {
    // Create session
    sess, _ := workflow.CreateSession(ctx, &workflow.SessionOptions{
        CreationTimeout:  time.Minute,
        ExecutionTimeout: 15 * time.Minute,
    })
    sessionID := workflow.GetSessionInfo(sess).SessionID

    // Cleanup on exit
    defer func() {
        workflow.ExecuteActivity(sess, CloseSessionClientActivity, CloseSessionClientParams{SessionID: sessionID})
        workflow.CompleteSession(sess)
    }()

    // Execute activity
    var result IngestComputeInstancesResult
    err := workflow.ExecuteActivity(sess, IngestComputeInstancesActivity, IngestComputeInstancesParams{
        SessionID: sessionID,
        ProjectID: params.ProjectID,
    }).Get(ctx, &result)

    return &InstanceWorkflowResult{...}, err
}
```

## Checklist

New resource implementation:

| Step | File | Action |
|------|------|--------|
| 1 | `client.go` | Wrap GCP API client, implement List method |
| 2 | `session.go` | Add session client functions (sync.Map + credentials) |
| 3 | `converter.go` | Convert API response → bronze model |
| 4 | `diff.go` | Implement change detection (parent + children) |
| 5 | `history.go` | Implement SCD Type 4 history tracking |
| 6 | `service.go` | Implement Ingest, save, delete stale with history |
| 7 | `activities.go` | Add Params/Result structs, activity var, method |
| 8 | `workflows.go` | Create workflow with session management |
| 9 | `register.go` | Register activities + workflow with worker |
| 10 | parent `register.go` | Import and call `Register()` |
| 11 | parent `workflows.go` | Add child workflow execution |

## Error Handling

| Scenario | Action |
|----------|--------|
| Client creation fails | Return error (Temporal retries) |
| Service error | Return error with context |
| Cleanup error (stale delete) | Log warning, don't fail activity |
| Session client not found | Create new client |

## History Integration

Service layer handles history tracking:

```go
// In service.go
func (s *Service) saveInstances(ctx context.Context, instances []bronze.GCPComputeInstance) error {
    // 1. Load existing with Preload (all relations)
    // 2. Compute diff
    // 3. Skip if no changes (update collected_at only)
    // 4. Delete old relations (manual cascade)
    // 5. Upsert resource
    // 6. Create new relations
    // 7. Track history (CreateHistory or UpdateHistory based on diff)
}

func (s *Service) DeleteStaleInstances(ctx context.Context, projectID string, collectedAt time.Time) error {
    // 1. Find stale (collected_at < latest)
    // 2. Close history (set valid_to)
    // 3. Delete relations
    // 4. Delete resource
}
```

See [HISTORY.md](../architecture/HISTORY.md) for history tracking details.
