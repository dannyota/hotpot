package greennode

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:greennode",
		ReqPerMin:   configService.GreenNodeRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	activities := NewActivities(configService)
	w.RegisterActivity(activities.DiscoverRegions)

	portal.Register(w, configService, entClient, limiter)
	compute.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GreenNodeInventoryWorkflow)

	return rateLimitSvc
}
