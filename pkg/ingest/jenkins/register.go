package jenkins

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/jenkins/job"
	entjenkins "github.com/dannyota/hotpot/pkg/storage/ent/jenkins"
)

// Register registers all Jenkins activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:jenkins",
		ReqPerMin:   configService.JenkinsRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	entClient := entjenkins.NewClient(entjenkins.Driver(driver), entjenkins.AlternateSchema(entjenkins.DefaultSchemaConfig()))

	job.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(JenkinsInventoryWorkflow)

	return rateLimitSvc
}
