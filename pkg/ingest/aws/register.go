package aws

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/aws/ec2"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all AWS activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	// Create shared rate limiter for all AWS API calls
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:aws",
		ReqPerMin:   configService.AWSRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	// Register region discovery activities
	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverRegions)

	// Register EC2 (instances)
	ec2.Register(w, configService, entClient, limiter)

	// Register AWS inventory workflow
	w.RegisterWorkflow(AWSInventoryWorkflow)

	return rateLimitSvc
}
