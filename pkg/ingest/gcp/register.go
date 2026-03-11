package gcp

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest"
)

// serviceRegFunc is the function signature for GCP service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, dialect.Driver, ratelimit.Limiter)

// Register registers all GCP activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	// Create shared rate limiter for all GCP API calls
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:gcp",
		ReqPerMin:   configService.GCPRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	// Register provider-level activities (project discovery, API discovery)
	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverProjects)
	w.RegisterActivity(activities.DiscoverEnabledAPIs)
	w.RegisterActivity(activities.GetConfigQuotaProject)

	for _, svc := range ingest.Services("gcp") {
		svc.Register.(serviceRegFunc)(w, configService, driver, limiter)
	}

	// Register GCP inventory workflow
	w.RegisterWorkflow(GCPInventoryWorkflow)

	return rateLimitSvc
}
