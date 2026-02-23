package ingest

import (
	"io"
	"sync"

	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ProviderRegistration describes a provider that can be started by the ingest runner.
// Providers self-register via init() using RegisterProvider.
type ProviderRegistration struct {
	Name               string
	TaskQueue          string
	Enabled            func(*config.Service) bool
	RateLimitPerMinute func(*config.Service) int
	Register           func(worker.Worker, *config.Service, *ent.Client) io.Closer

	// Workflow is the top-level inventory workflow function for scheduling.
	Workflow interface{}

	// WorkflowArgs are the default arguments passed to the workflow.
	// Nil for workflows that only take workflow.Context.
	WorkflowArgs []interface{}
}

var (
	mu        sync.Mutex
	providers []ProviderRegistration
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
