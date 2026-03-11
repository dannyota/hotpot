package meec

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest"
	entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"
)

// serviceRegFunc is the function signature for MEEC service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, *entinventory.Client, ratelimit.Limiter, *TokenSource)

// Register registers all MEEC activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:meec",
		ReqPerMin:   configService.MEECRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	entClient := entinventory.NewClient(entinventory.Driver(driver), entinventory.AlternateSchema(entinventory.DefaultSchemaConfig()))

	tokenSource := NewTokenSource(
		configService.MEECBaseURL(),
		configService.MEECAPIVersion(),
		configService.MEECUsername(),
		configService.MEECPassword(),
		configService.MEECAuthType(),
		configService.MEECTOTPSecret(),
		configService.MEECVerifySSL(),
	)

	for _, svc := range ingest.Services("meec") {
		svc.Register.(serviceRegFunc)(w, configService, entClient, limiter, tokenSource)
	}

	w.RegisterWorkflow(MEECInventoryWorkflow)

	return rateLimitSvc
}
