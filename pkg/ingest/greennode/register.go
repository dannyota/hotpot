package greennode

import (
	"entgo.io/ent/dialect"
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest"
)

// serviceRegFunc is the function signature for GreenNode service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, dialect.Driver, *auth.IAMUserAuth, ratelimit.Limiter)

// Register registers all GreenNode activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
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
		svc.Register.(serviceRegFunc)(w, configService, driver, activities.iamAuth, limiter)
	}

	w.RegisterWorkflow(GreenNodeInventoryWorkflow)

	return rateLimitSvc
}
