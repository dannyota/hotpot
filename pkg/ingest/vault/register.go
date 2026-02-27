package vault

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest"
)

// serviceRegFunc is the function signature for Vault service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, dialect.Driver, ratelimit.Limiter)

// Register registers all Vault activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	// Create shared rate limiter for all Vault API calls
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:vault",
		ReqPerMin:   configService.VaultRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	// Register provider-level activities
	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.ListVaultInstances)

	for _, svc := range ingest.Services("vault") {
		svc.Register.(serviceRegFunc)(w, configService, driver, limiter)
	}

	// Register inventory workflow
	w.RegisterWorkflow(VaultInventoryWorkflow)

	return rateLimitSvc
}
