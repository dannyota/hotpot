package reference

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest"
	entreference "danny.vn/hotpot/pkg/storage/ent/reference"
)

// serviceRegFunc is the function signature for reference service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, *entreference.Client, ratelimit.Limiter)

// Register registers all reference data activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:reference",
		ReqPerMin:   configService.ReferenceRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	entClient := entreference.NewClient(entreference.Driver(driver), entreference.AlternateSchema(entreference.DefaultSchemaConfig()))

	for _, svc := range ingest.Services("reference") {
		svc.Register.(serviceRegFunc)(w, configService, entClient, limiter)
	}

	w.RegisterWorkflow(ReferenceInventoryWorkflow)

	return rateLimitSvc
}
