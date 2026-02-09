package digitalocean

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/vpc"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all DigitalOcean activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:do",
		ReqPerMin:   configService.DORateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	vpc.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(DOInventoryWorkflow)

	return rateLimitSvc
}
