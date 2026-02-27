package sentinelone

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest"
	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
)

// serviceRegFunc is the function signature for SentinelOne service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, *ents1.Client, ratelimit.Limiter)

// Register registers all SentinelOne activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:s1",
		ReqPerMin:   configService.S1RateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	entClient := ents1.NewClient(ents1.Driver(driver), ents1.AlternateSchema(ents1.DefaultSchemaConfig()))

	for _, svc := range ingest.Services("sentinelone") {
		svc.Register.(serviceRegFunc)(w, configService, entClient, limiter)
	}

	w.RegisterWorkflow(S1InventoryWorkflow)

	return rateLimitSvc
}
