package sentinelone

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/account"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/agent"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/app"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/group"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/site"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/threat"
	"github.com/dannyota/hotpot/pkg/storage/ent"
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
	app.Register(w, configService, entClient, limiter)
	group.Register(w, configService, entClient, limiter)
	site.Register(w, configService, entClient, limiter)
	threat.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(S1InventoryWorkflow)

	return rateLimitSvc
}
