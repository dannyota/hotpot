# Ent Schema Guide

Type-safe ent schema patterns for Hotpot's multi-layer data model (bronze, bronzehistory, silver, gold).

## Quick Start Checklist

When adding a new GCP resource type:

- [ ] Create bronze schema in `pkg/schema/bronze/gcp/{service}/{resource}.go`
- [ ] Create bronzehistory schema in `pkg/schema/bronzehistory/gcp/{service}/{resource}.go`
- [ ] Run `cd pkg/storage && go generate`
- [ ] Verify compilation: `go build ./...`

## Naming Conventions

### Go Type Names (CRITICAL)

**Always prefix type names with the layer:**

```go
// ✅ CORRECT - Prefixed with layer
type BronzeGCPComputeInstance struct {
    ent.Schema
}

type BronzeHistoryGCPComputeInstance struct {
    ent.Schema
}

// ❌ WRONG - No layer prefix
type GCPComputeInstance struct {  // DON'T DO THIS
    ent.Schema
}
```

**Pattern**: `{Layer}{Provider}{Service}{Resource}`
- Layer: `Bronze`, `BronzeHistory`, `Silver`, `Gold`
- Provider: `GCP`, `AWS`, `Azure`, `VNG`
- Service: `Compute`, `Networking`, `IAM`, `Container`
- Resource: `Instance`, `Disk`, `Network`, `Cluster`

### Table Names

**Use service_resource format, history tables add `_history` suffix:**

```go
// Bronze
entsql.Annotation{Table: "gcp_compute_instances"}

// Bronze History - add _history suffix to TABLE NAME
entsql.Annotation{Table: "gcp_compute_instances_history"}
```

**Pattern**: `{provider}_{service}_{resource}[_history]`

⚠️ **IMPORTANT**: The `_history` suffix goes in the TABLE NAME annotation, NOT in the Go type name!

## Bronze Schema Pattern

```go
package compute

import (
    "encoding/json"

    "entgo.io/ent"
    "entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"

    "hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeInstance - main entity
type BronzeGCPComputeInstance struct {
    ent.Schema
}

// Use mixin for collected_at
func (BronzeGCPComputeInstance) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.Timestamp{},
    }
}

func (BronzeGCPComputeInstance) Fields() []ent.Field {
    return []ent.Field{
        // Primary key - use resource_id from API
        field.String("id").
            StorageKey("resource_id").
            Immutable(),

        // String fields
        field.String("name").NotEmpty(),
        field.String("status").Optional(),

        // Boolean fields
        field.Bool("deletion_protection").Default(false),

        // Integer fields
        field.Int32("cpu_count").Default(0),

        // JSON fields - use json.RawMessage{}
        field.JSON("metadata_json", json.RawMessage{}).Optional(),

        // Collection metadata
        field.String("project_id").NotEmpty(),
        // collected_at comes from mixin - DON'T add it here
    }
}

func (BronzeGCPComputeInstance) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("labels", BronzeGCPComputeInstanceLabel.Type),
        edge.To("disks", BronzeGCPComputeInstanceDisk.Type),
    }
}

func (BronzeGCPComputeInstance) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("status"),
        index.Fields("project_id"),
        index.Fields("collected_at"),
    }
}

func (BronzeGCPComputeInstance) Annotations() []schema.Annotation {
    return []schema.Annotation{
        // NO entsql.Schema() - set at runtime
        entsql.Annotation{Table: "gcp_compute_instances"},
    }
}

// BronzeGCPComputeInstanceLabel - child entity
type BronzeGCPComputeInstanceLabel struct {
    ent.Schema
}

func (BronzeGCPComputeInstanceLabel) Fields() []ent.Field {
    return []ent.Field{
        // NO explicit foreign key field - edge handles it
        field.String("key").NotEmpty(),
        field.String("value").Optional(),
    }
}

func (BronzeGCPComputeInstanceLabel) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("instance", BronzeGCPComputeInstance.Type).
            Ref("labels").
            Unique().
            Required(),
    }
}

func (BronzeGCPComputeInstanceLabel) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{Table: "gcp_compute_instance_labels"},
    }
}
```

## Bronze History Schema Pattern

```go
package compute

import (
    "encoding/json"

    "entgo.io/ent"
    "entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

// BronzeHistoryGCPComputeInstance - historical snapshots
type BronzeHistoryGCPComputeInstance struct {
    ent.Schema
}

// NO Mixin() - history has explicit collected_at field

func (BronzeHistoryGCPComputeInstance) Fields() []ent.Field {
    return []ent.Field{
        // History tracking fields
        field.Uint("history_id").Unique().Immutable(),
        field.String("resource_id").NotEmpty(),
        field.Time("valid_from").Immutable(),
        field.Time("valid_to").Optional().Nillable(),

        // All fields from bronze (copy them all)
        field.String("name").NotEmpty(),
        field.String("status").Optional(),
        field.Bool("deletion_protection").Default(false),
        field.Int32("cpu_count").Default(0),
        field.JSON("metadata_json", json.RawMessage{}).Optional(),

        // Collection metadata
        field.String("project_id").NotEmpty(),
        field.Time("collected_at"),  // Explicit in history
    }
}

// NO Edges() - history tables don't have relationships

func (BronzeHistoryGCPComputeInstance) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("resource_id", "valid_from"),
        index.Fields("valid_to"),
        index.Fields("collected_at"),
    }
}

func (BronzeHistoryGCPComputeInstance) Annotations() []schema.Annotation {
    return []schema.Annotation{
        // _history suffix in TABLE NAME
        entsql.Annotation{Table: "gcp_compute_instances_history"},
    }
}

// BronzeHistoryGCPComputeInstanceLabel - child history
type BronzeHistoryGCPComputeInstanceLabel struct {
    ent.Schema
}

func (BronzeHistoryGCPComputeInstanceLabel) Fields() []ent.Field {
    return []ent.Field{
        field.Uint("history_id").Unique().Immutable(),
        field.Uint("instance_history_id"),  // Link to parent
        field.Time("valid_from").Immutable(),
        field.Time("valid_to").Optional().Nillable(),

        field.String("key").Optional(),
        field.String("value").Optional(),
    }
}

// NO Edges() - history tables don't have relationships

func (BronzeHistoryGCPComputeInstanceLabel) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("instance_history_id"),
        index.Fields("valid_from"),
        index.Fields("valid_to"),
    }
}

func (BronzeHistoryGCPComputeInstanceLabel) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{Table: "gcp_compute_instance_labels_history"},
    }
}
```

## Common Patterns

### Field Types

```go
// String fields
field.String("name").NotEmpty()          // Required
field.String("description").Optional()   // Nullable
field.Text("long_text").Optional()       // TEXT type

// Numeric fields
field.Int("count").Default(0)
field.Int32("small_number").Default(0)
field.Int64("big_number").Default(0)
field.Float("decimal").Optional()

// Boolean fields
field.Bool("enabled").Default(false)

// Time fields
field.Time("created_at")
field.Time("deleted_at").Optional().Nillable()

// JSON fields - ALWAYS use json.RawMessage{}
field.JSON("settings_json", json.RawMessage{}).Optional()
```

### Parent-Child Relationships

```go
// Parent defines edge TO children
func (Parent) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("children", Child.Type),  // CASCADE DELETE automatic
    }
}

// Child defines edge FROM parent
func (Child) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("parent", Parent.Type).
            Ref("children").  // Must match parent edge name
            Unique().         // One parent
            Required(),       // NOT NULL
    }
}
```

## Common Mistakes

### ❌ Adding explicit foreign key fields

```go
// WRONG - Don't do this
func (BronzeGCPComputeInstanceLabel) Fields() []ent.Field {
    return []ent.Field{
        field.String("instance_resource_id").NotEmpty(),  // ❌ DON'T
        field.String("key").NotEmpty(),
    }
}
```

Let ent manage foreign keys through edges automatically.

### ❌ Using mixin in history schemas

```go
// WRONG
type BronzeHistoryGCPComputeInstance struct {
    ent.Schema
}

func (BronzeHistoryGCPComputeInstance) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.Timestamp{},  // ❌ DON'T - causes duplicate collected_at
    }
}
```

History schemas should have explicit `collected_at` field.

### ❌ Adding entsql.Schema() annotation

```go
// WRONG
func (BronzeGCPComputeInstance) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Schema("bronze"),  // ❌ DON'T - triggers Atlas Pro requirement
        entsql.Annotation{Table: "gcp_compute_instances"},
    }
}
```

Schema is set at runtime via `AlternateSchema()`.

### ❌ Adding _history to type name

```go
// WRONG
type BronzeGCPComputeInstance_history struct {  // ❌ DON'T
    ent.Schema
}
```

Use `BronzeHistoryGCPComputeInstance` instead. The `_history` suffix goes in the table name only.

### ❌ Adding edges to history schemas

```go
// WRONG
func (BronzeHistoryGCPComputeInstance) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("labels", BronzeHistoryGCPComputeInstanceLabel.Type),  // ❌ DON'T
    }
}
```

History tables are denormalized - no relationships.

## Generating Code

After creating or modifying schemas:

```bash
cd pkg/storage
go generate
```

This runs `entc.go` which:
1. Discovers all schemas in `pkg/schema/`
2. Generates wrapper structs
3. Runs ent code generation
4. Creates `pkg/storage/ent/` with all client code

## Workaround: Ent Issue #2330

Due to [ent issue #2330](https://github.com/ent/ent/issues/2330), we must use `_history` suffix on history table names to avoid duplicate constant collisions.

**See**: `docs/decisions/002-ent-issue-2330-workaround.md`

When ent fixes this issue, we'll rename history tables to remove the suffix.
