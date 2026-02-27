package reference

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest"
	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
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
