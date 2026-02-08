# Code Style

Code style and testing conventions for Hotpot.

## File Organization

Keep files focused on a single responsibility. Split when:
- A file covers multiple unrelated concerns
- Navigation becomes difficult

No hard line limit—prefer logical grouping.

**Ent schema files** — group parent and child schemas together:

```go
// pkg/schema/bronze/gcp/compute/instance.go - all instance-related schemas
type BronzeGCPComputeInstance struct { ent.Schema }
type BronzeGCPComputeInstanceDisk struct { ent.Schema }
type BronzeGCPComputeInstanceNIC struct { ent.Schema }
// ... other children
```

See [ENT_SCHEMAS.md](ENT_SCHEMAS.md) for ent schema patterns.

## Naming

Follow [Google Go Style Guide](https://google.github.io/styleguide/go/):
- Short names for small scopes (`i`, `n`, `err`)
- Descriptive names for package-level or wide scope
- MixedCaps, not underscores
- Acronyms: consistent case (`HTTPClient` or `httpClient`)

### Ent Schema Naming

| Component | Pattern | Example |
|-----------|---------|---------|
| Type name | `{Layer}{Provider}{Service}{Resource}` | `BronzeGCPComputeInstance` |
| Table name | `{provider}_{service}_{resource}` | `gcp_compute_instances` |
| History table | `{provider}_{service}_{resource}_history` | `gcp_compute_instances_history` |

## Code Patterns

- Early returns over deep nesting
- One responsibility per function
- No dead code or commented-out blocks

## Comments

- Doc comments: start with the name being documented
- Explain *why*, not *what*
- Use `/* param */` for unclear arguments ([Uber guide](https://github.com/uber-go/guide))

## Error Handling

- Always handle errors; never use `_`
- Wrap with context: `fmt.Errorf("fetch assets: %w", err)`
- Use sentinel errors for expected conditions
- Always propagate `json.Marshal` errors — never silently ignore them

## Ent Schema Patterns

**Bronze schemas** — use API field names for traceability:

```go
// pkg/schema/bronze/gcp/compute/instance.go
func (BronzeGCPComputeInstance) Fields() []ent.Field {
    return []ent.Field{
        // Primary key - maps to resource_id column
        field.String("id").StorageKey("resource_id").Immutable(),

        // GCP API fields - names match API response
        field.String("name").NotEmpty(),
        field.String("machine_type").Optional(),
        field.String("status").Optional(),

        // Collection metadata (not from API)
        field.String("project_id").NotEmpty(),
        // collected_at comes from mixin
    }
}

func (BronzeGCPComputeInstance) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{Table: "gcp_compute_instances"},
    }
}
```

**Bronze history schemas** — denormalized snapshots:

```go
// pkg/schema/bronzehistory/gcp/compute/instance.go
func (BronzeHistoryGCPComputeInstance) Fields() []ent.Field {
    return []ent.Field{
        // History tracking
        field.Uint("history_id").Unique().Immutable(),
        field.String("resource_id").NotEmpty(),
        field.Time("valid_from").Immutable(),
        field.Time("valid_to").Optional().Nillable(),

        // All bronze fields repeated
        field.String("name").NotEmpty(),
        field.String("machine_type").Optional(),
        field.String("status").Optional(),
        field.String("project_id").NotEmpty(),
        field.Time("collected_at"),  // Explicit in history
    }
}

func (BronzeHistoryGCPComputeInstance) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{Table: "gcp_compute_instances_history"},
    }
}
```

## Field Types

### String Fields

```go
field.String("name").NotEmpty()          // Required, non-empty
field.String("description").Optional()   // Nullable
field.Text("long_description").Optional() // TEXT type for long content
```

### Numeric Fields

```go
field.Int("count").Default(0)            // int
field.Int32("small_number").Default(0)   // int32
field.Int64("big_number").Default(0)     // int64
field.Float("decimal").Optional()        // float64
field.Uint("id").Unique()                // unsigned int (for history_id)
```

### Boolean Fields

```go
field.Bool("enabled").Default(false)     // Required, default false
field.Bool("deleted").Optional()         // Nullable
```

### Time Fields

```go
field.Time("created_at")                 // Required
field.Time("deleted_at").Optional().Nillable() // Nullable
```

### JSON Fields

**Always use `json.RawMessage{}`** for JSONB columns:

```go
import "encoding/json"

field.JSON("settings_json", json.RawMessage{}).Optional()
field.JSON("metadata_json", json.RawMessage{}).Optional()
```

**Converter** — nil-check + `json.Marshal`:
```go
if obj.Settings != nil {
    model.SettingsJSON, err = json.Marshal(obj.Settings)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal settings: %w", err)
    }
}
// nil field → stays nil → SQL NULL
// non-nil field → JSON bytes → valid JSONB
```

**Diff** — use `bytes.Equal` for comparison:
```go
import "bytes"

if !bytes.Equal(old.SettingsJSON, new.SettingsJSON) {
    // changed
}
```

## Imports

Group and separate with blank lines:

```go
import (
    "context"
    "encoding/json"
    "fmt"

    "entgo.io/ent"
    "entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema"
    "entgo.io/ent/schema/field"

    "hotpot/pkg/schema/bronze/mixin"
    "hotpot/pkg/storage/ent"
)
```

## Testing

### Test Structure

- Table-driven tests with `name`, `give`, `want`
- Use `t.Run()` for subtests
- Test file next to source: `client.go` → `client_test.go`

### Testing by Layer

| Layer | Strategy |
|-------|----------|
| Bronze | Interface mocking for external clients |
| Silver | Test fixtures with sample bronze data |
| Gold | Test fixtures with sample silver data |

Future: integration tests with recorded responses (go-vcr).

## Common Ent Patterns

### Parent-Child Relationships

```go
// Parent edge
func (BronzeGCPComputeInstance) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("disks", BronzeGCPComputeInstanceDisk.Type),
        edge.To("labels", BronzeGCPComputeInstanceLabel.Type),
    }
}

// Child edge
func (BronzeGCPComputeInstanceDisk) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("instance", BronzeGCPComputeInstance.Type).
            Ref("disks").
            Unique().
            Required(),
    }
}
```

### Querying with Edges

```go
// Load all edges
instance, err := client.BronzeGCPComputeInstance.Query().
    Where(bronzegcpcomputeinstance.IDEQ(id)).
    WithDisks().
    WithLabels().
    WithNics().
    Only(ctx)

// Filter through edges
instances, err := client.BronzeGCPComputeInstance.Query().
    Where(bronzegcpcomputeinstance.HasDisksWith(
        bronzegcpcomputeinstancedisk.SourceEQ("image-1"),
    )).
    All(ctx)
```

### Transactions

```go
tx, err := client.Tx(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

// All operations in same transaction
if _, err := tx.BronzeGCPComputeInstance.Create()...; err != nil {
    return err
}
if _, err := tx.BronzeHistoryGCPComputeInstance.Create()...; err != nil {
    return err
}

return tx.Commit()
```

## References

- [ENT_SCHEMAS.md](ENT_SCHEMAS.md) - Complete ent schema guide
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Guide](https://github.com/uber-go/guide)
