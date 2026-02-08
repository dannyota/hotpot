# GORM to Ent Migration Guide

This guide explains the completed migration from GORM to Ent for Hotpot's data layer.

## Why Ent?

| Feature | GORM | Ent |
|---------|------|-----|
| Schema definition | Struct tags | Type-safe Go code |
| Code generation | No | Yes (reduces boilerplate) |
| Type safety | Weak (string-based queries) | Strong (compile-time checks) |
| Schema-qualified SQL | Manual `search_path` | Built-in `AlternateSchema()` |
| Migration tool | Manual | Atlas integration |
| Multi-schema support | Limited | First-class with `FeatureSchemaConfig` |
| Auto-discovery | No | Yes (via custom entc.go) |

## Key Differences

### Model Definition

**GORM (old):**
```go
type GCPComputeInstance struct {
    ResourceID  string    `gorm:"primaryKey;column:resource_id"`
    Name        string    `gorm:"column:name;not null"`
    Status      string    `gorm:"column:status;index"`
    ProjectID   string    `gorm:"column:project_id;index"`
    CollectedAt time.Time `gorm:"column:collected_at;index"`

    // Relationships
    Disks []GCPComputeInstanceDisk `gorm:"foreignKey:InstanceResourceID"`
}
```

**Ent (new):**
```go
// pkg/schema/bronze/gcp/compute/instance.go
type BronzeGCPComputeInstance struct {
    ent.Schema
}

func (BronzeGCPComputeInstance) Mixin() []ent.Mixin {
    return []ent.Mixin{mixin.Timestamp{}}
}

func (BronzeGCPComputeInstance) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").StorageKey("resource_id").Immutable(),
        field.String("name").NotEmpty(),
        field.String("status").Optional(),
        field.String("project_id").NotEmpty(),
    }
}

func (BronzeGCPComputeInstance) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("disks", BronzeGCPComputeInstanceDisk.Type),
    }
}
```

### Querying

**GORM (old):**
```go
var instances []bronze.GCPComputeInstance
err := db.Where("project_id = ? AND status = ?", projectID, "RUNNING").
    Preload("Disks").
    Find(&instances).Error
```

**Ent (new):**
```go
instances, err := client.BronzeGCPComputeInstance.Query().
    Where(
        bronzegcpcomputeinstance.ProjectIDEQ(projectID),
        bronzegcpcomputeinstance.StatusEQ("RUNNING"),
    ).
    WithDisks().
    All(ctx)
```

### Creating Records

**GORM (old):**
```go
instance := &bronze.GCPComputeInstance{
    ResourceID: "123",
    Name: "my-instance",
}
err := db.Create(instance).Error
```

**Ent (new):**
```go
instance, err := client.BronzeGCPComputeInstance.Create().
    SetID("123").
    SetName("my-instance").
    Save(ctx)
```

### Transactions

**GORM (old):**
```go
tx := db.Begin()
defer tx.Rollback()

if err := tx.Create(&instance).Error; err != nil {
    return err
}
if err := tx.Create(&history).Error; err != nil {
    return err
}
return tx.Commit().Error
```

**Ent (new):**
```go
tx, err := client.Tx(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

if _, err := tx.BronzeGCPComputeInstance.Create()...; err != nil {
    return err
}
if _, err := tx.BronzeHistoryGCPComputeInstance.Create()...; err != nil {
    return err
}
return tx.Commit()
```

## Migration Checklist

### Phase 1: Schema Creation ✅
- [x] Create `pkg/schema/bronze/` with all GCP resource schemas
- [x] Create `pkg/schema/bronzehistory/` with all history schemas
- [x] Generate ent code via `cd pkg/storage && go generate`
- [x] Verify compilation

### Phase 2: Service Layer (In Progress)
- [ ] Update `service.go` files to use ent client
- [ ] Update `history.go` files to use ent client
- [ ] Update `diff.go` files (replace `jsonb.Changed` with `bytes.Equal`)
- [ ] Update `converter.go` to return ent types

### Phase 3: Activities
- [ ] Update `activities.go` to accept `*ent.Client` instead of `*gorm.DB`
- [ ] Update `register.go` signatures

### Phase 4: Infrastructure
- [ ] Update `pkg/base/app/database.go` to create ent client
- [ ] Update `pkg/ingest/gcp/run.go` to pass ent client
- [ ] Remove GORM initialization

### Phase 5: Cleanup
- [ ] Delete `pkg/base/models/bronze/`
- [ ] Delete `pkg/base/models/bronze_history/`
- [ ] Delete `pkg/base/jsonb/`
- [ ] Remove `gorm.io/gorm` from `go.mod`
- [ ] Remove `gorm.io/driver/postgres` from `go.mod`

## Common Patterns

### JSONB Fields

**GORM:**
```go
import "hotpot/pkg/base/jsonb"

UsersJSON jsonb.JSON `gorm:"column:users_json;type:jsonb"`

// Diff
if jsonb.Changed(old.UsersJSON, new.UsersJSON) { ... }
```

**Ent:**
```go
import "encoding/json"

field.JSON("users_json", json.RawMessage{}).Optional()

// Diff
if !bytes.Equal(old.UsersJSON, new.UsersJSON) { ... }
```

### Checking for Changes

**GORM:**
```go
if jsonb.Changed(old.SettingsJSON, new.SettingsJSON) {
    // changed
}
```

**Ent:**
```go
import "bytes"

if !bytes.Equal(old.SettingsJSON, new.SettingsJSON) {
    // changed
}
```

### Cascade Delete

**GORM:**
Manually delete children:
```go
db.Where("instance_resource_id = ?", id).Delete(&GCPComputeInstanceDisk{})
db.Delete(&instance)
```

**Ent:**
Automatic via edge annotation:
```go
edge.To("disks", BronzeGCPComputeInstanceDisk.Type)
// CASCADE DELETE happens automatically
client.BronzeGCPComputeInstance.DeleteOneID(id).Exec(ctx)
```

## Benefits Realized

✅ **Type safety** - Compile-time query validation
✅ **Schema-qualified SQL** - Zero `search_path` overhead
✅ **Auto-discovery** - No manual schema registration
✅ **Code generation** - Less boilerplate
✅ **Better migrations** - Atlas integration
✅ **Cleaner code** - No struct tags, proper Go methods

## References

- [Ent Documentation](https://entgo.io/docs/getting-started)
- [ENT_SCHEMAS.md](ENT_SCHEMAS.md) - Hotpot schema patterns
- [Atlas Documentation](https://atlasgo.io/docs)
- [Ent Issue #2330](https://github.com/ent/ent/issues/2330) - Table name workaround
