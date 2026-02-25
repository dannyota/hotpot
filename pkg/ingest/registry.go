package ingest

import (
	"io"
	"sync"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
)

// ProviderRegistration describes a provider that can be started by the ingest runner.
// Providers self-register via init() using RegisterProvider.
type ProviderRegistration struct {
	Name               string
	TaskQueue          string
	Enabled            func(*config.Service) bool
	RateLimitPerMinute func(*config.Service) int
	Register           func(worker.Worker, *config.Service, dialect.Driver) io.Closer

	// Workflow is the top-level inventory workflow function for scheduling.
	Workflow interface{}

	// WorkflowArgs are the default arguments passed to the workflow.
	// Nil for workflows that only take workflow.Context.
	WorkflowArgs []interface{}
}

// ServiceScope indicates whether a service runs once globally or per-region.
type ServiceScope int

const (
	// ScopeRegional means the service runs per-region per-project.
	ScopeRegional ServiceScope = iota
	// ScopeGlobal means the service runs once using the first project/region.
	ScopeGlobal
)

// ServiceRegistration describes a service within a provider.
// Services self-register via init() using RegisterService.
type ServiceRegistration struct {
	Provider string
	Name     string
	Scope    ServiceScope

	// Register is the provider-specific registration function.
	// Called via type assertion by the provider's register.go.
	Register any

	// Workflow is the Temporal workflow function for this service.
	Workflow any

	// NewParams creates the workflow params struct from projectID and region.
	NewParams func(projectID, region string) any

	// NewResult creates a zero-value result pointer for the workflow.
	NewResult func() any

	// Aggregate merges a service result into the provider-level result.
	// Called via type assertion by the provider's workflows.go.
	Aggregate any
}

var (
	mu        sync.Mutex
	providers []ProviderRegistration
	services  []ServiceRegistration
)

// RegisterProvider adds a provider to the global registry.
// It is intended to be called from provider init() functions.
func RegisterProvider(p ProviderRegistration) {
	mu.Lock()
	defer mu.Unlock()
	providers = append(providers, p)
}

// Providers returns a copy of all registered providers.
func Providers() []ProviderRegistration {
	mu.Lock()
	defer mu.Unlock()
	out := make([]ProviderRegistration, len(providers))
	copy(out, providers)
	return out
}

// ResetProviders clears the registry. Intended for tests only.
func ResetProviders() {
	mu.Lock()
	defer mu.Unlock()
	providers = nil
}

// RegisterService adds a service to the global registry.
// It is intended to be called from service init() functions.
func RegisterService(s ServiceRegistration) {
	mu.Lock()
	defer mu.Unlock()
	services = append(services, s)
}

// Services returns all registered services for the given provider.
func Services(provider string) []ServiceRegistration {
	mu.Lock()
	defer mu.Unlock()
	var out []ServiceRegistration
	for _, s := range services {
		if s.Provider == provider {
			out = append(out, s)
		}
	}
	return out
}

// ResetServices clears the service registry. Intended for tests only.
func ResetServices() {
	mu.Lock()
	defer mu.Unlock()
	services = nil
}
