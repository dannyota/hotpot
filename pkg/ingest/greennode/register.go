package greennode

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// serviceRegFunc is the function signature for GreenNode service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, *ent.Client, *auth.IAMUserAuth, ratelimit.Limiter)

// Register registers all GreenNode activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:greennode",
		ReqPerMin:   configService.GreenNodeRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverRegions)
	w.RegisterActivity(activities.DiscoverProjects)

	for _, svc := range ingest.Services("greennode") {
		svc.Register.(serviceRegFunc)(w, configService, entClient, activities.iamAuth, limiter)
	}

	w.RegisterWorkflow(GreenNodeInventoryWorkflow)

	return rateLimitSvc
}
