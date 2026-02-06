# Code Style

Code style and testing conventions for Hotpot.

## File Organization

Keep files focused on a single responsibility. Split when:
- A file covers multiple unrelated concerns
- Navigation becomes difficult

No hard line limit—prefer logical grouping.

**Model files** — group parent and child models together:

```go
// gcp_compute_instance.go - all instance-related models
type GCPComputeInstance struct { ... }
type GCPComputeInstanceDisk struct { ... }
type GCPComputeInstanceNIC struct { ... }
// ... other children
```

See [PRINCIPLES.md](../architecture/PRINCIPLES.md#8-model-conventions) for details.

## Naming

Follow [Google Go Style Guide](https://google.github.io/styleguide/go/):
- Short names for small scopes (`i`, `n`, `err`)
- Descriptive names for package-level or wide scope
- MixedCaps, not underscores
- Acronyms: consistent case (`HTTPClient` or `httpClient`)

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

## Model Tags

**Bronze models** — `json` tag = original API field name (for traceability):

```go
type GCPComputeInstance struct {
    // gorm = our name + type, json = original API field
    ResourceID  string `gorm:"column:resource_id;type:varchar(255);uniqueIndex" json:"id"`
    Name        string `gorm:"column:name;type:varchar(255);not null" json:"name"`
    MachineType string `gorm:"column:machine_type;type:text" json:"machineType"`
    Status      string `gorm:"column:status;type:varchar(50);index" json:"status"`

    // Collection metadata (not from API)
    ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
    CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`
}
```

**Silver/Gold models** — `json` tag matches column name (for API responses):

```go
type Asset struct {
    ID       string `gorm:"column:id" json:"id"`
    Name     string `gorm:"column:name" json:"name"`
    SourceID string `gorm:"column:source_id" json:"source_id"`
}
```

| Layer | `json` tag purpose |
|-------|-------------------|
| Bronze | Original external API field name |
| Silver/Gold | API response field (matches column) |

## JSONB Fields

Use `jsonb.JSON` (from `hotpot/pkg/base/jsonb`) with `type:jsonb` GORM tag for JSONB columns. This is `[]byte` under the hood — nil maps to SQL NULL, non-nil maps to valid JSONB. Do NOT use `gorm.io/datatypes` — it pulls in MySQL/SQLite as transitive dependencies.

**Model:**
```go
import "hotpot/pkg/base/jsonb"

UsersJSON jsonb.JSON `gorm:"column:users_json;type:jsonb" json:"users"`
```

**Converter** — nil-check + `json.Marshal`:
```go
if obj.Users != nil {
    model.UsersJSON, err = json.Marshal(obj.Users)
    if err != nil {
        return Model{}, fmt.Errorf("failed to marshal users: %w", err)
    }
}
// nil field → stays nil → SQL NULL
// non-nil field → JSON bytes → valid JSONB
```

**Diff** — use `jsonb.Changed` for comparison (`[]byte` cannot use `!=`):
```go
import "hotpot/pkg/base/jsonb"

jsonb.Changed(old.UsersJSON, new.UsersJSON)
```

## Imports

Group and separate with blank lines:

```go
import (
    "context"
    "fmt"

    "go.temporal.io/sdk/workflow"

    "hotpot/pkg/base"
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
