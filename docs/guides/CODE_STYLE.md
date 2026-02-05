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
