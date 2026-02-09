package sentinelone

import (
	"go.temporal.io/sdk/worker"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/sentinelone/account"
	"hotpot/pkg/ingest/sentinelone/agent"
	"hotpot/pkg/ingest/sentinelone/threat"
	"hotpot/pkg/storage/ent"
)

// Register registers all SentinelOne activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:s1",
		ReqPerMin:   configService.S1RateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	account.Register(w, configService, entClient, limiter)
	agent.Register(w, configService, entClient, limiter)
	threat.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(S1InventoryWorkflow)

	return rateLimitSvc
}
