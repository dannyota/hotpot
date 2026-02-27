package digitalocean

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest"
	entdo "github.com/dannyota/hotpot/pkg/storage/ent/do"
)

// serviceRegFunc is the function signature for DigitalOcean service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, *entdo.Client, ratelimit.Limiter)

// Register registers all DigitalOcean activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:do",
		ReqPerMin:   configService.DORateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	entClient := entdo.NewClient(entdo.Driver(driver), entdo.AlternateSchema(entdo.DefaultSchemaConfig()))

	for _, svc := range ingest.Services("digitalocean") {
		svc.Register.(serviceRegFunc)(w, configService, entClient, limiter)
	}

	w.RegisterWorkflow(DOInventoryWorkflow)

	return rateLimitSvc
}
