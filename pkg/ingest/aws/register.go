package aws

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest"
)

// serviceRegFunc is the function signature for AWS service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, dialect.Driver, ratelimit.Limiter)

// Register registers all AWS activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	// Create shared rate limiter for all AWS API calls
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:aws",
		ReqPerMin:   configService.AWSRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	// Register provider-level activities (region discovery)
	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverRegions)

	for _, svc := range ingest.Services("aws") {
		svc.Register.(serviceRegFunc)(w, configService, driver, limiter)
	}

	// Register AWS inventory workflow
	w.RegisterWorkflow(AWSInventoryWorkflow)

	return rateLimitSvc
}
