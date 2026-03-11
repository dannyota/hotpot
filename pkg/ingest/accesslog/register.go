package accesslog

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
	entaccesslog "danny.vn/hotpot/pkg/storage/ent/accesslog"
)

// serviceRegFunc is the function signature for accesslog service registration.
type serviceRegFunc = func(worker.Worker, *config.Service, *entaccesslog.Client)

// Register registers all accesslog activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) {
	entClient := entaccesslog.NewClient(
		entaccesslog.Driver(driver),
		entaccesslog.AlternateSchema(entaccesslog.DefaultSchemaConfig()),
	)

	// Register provider-level activities.
	activities := NewActivities(configService, entClient)
	w.RegisterActivity(activities.DiscoverLogSources)
	w.RegisterActivity(activities.CleanupStaleBronze)

	// Register all services for this provider.
	for _, svc := range ingest.Services("accesslog") {
		svc.Register.(serviceRegFunc)(w, configService, entClient)
	}

	w.RegisterWorkflow(AccessLogWorkflow)
}
