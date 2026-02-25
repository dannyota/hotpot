package aws

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/aws/ec2"
	entec2 "github.com/dannyota/hotpot/pkg/storage/ent/aws/ec2"
)

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

	// Register region discovery activities
	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverRegions)

	// Create per-service ent client
	entClient := entec2.NewClient(entec2.Driver(driver), entec2.AlternateSchema(entec2.DefaultSchemaConfig()))

	// Register EC2 (instances)
	ec2.Register(w, configService, entClient, limiter)

	// Register AWS inventory workflow
	w.RegisterWorkflow(AWSInventoryWorkflow)

	return rateLimitSvc
}
