package vault

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/vault/pki"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Vault activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
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

	// Register PKI service
	pki.Register(w, configService, entClient, limiter)

	// Register inventory workflow
	w.RegisterWorkflow(VaultInventoryWorkflow)

	return rateLimitSvc
}
