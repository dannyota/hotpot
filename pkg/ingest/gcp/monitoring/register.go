package monitoring

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/monitoring/alertpolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/monitoring/uptimecheck"
	"entgo.io/ent/dialect"
	entmonitoring "danny.vn/hotpot/pkg/storage/ent/gcp/monitoring"
)

// Register registers all Monitoring activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entmonitoring.NewClient(entmonitoring.Driver(driver), entmonitoring.AlternateSchema(entmonitoring.DefaultSchemaConfig()))
	alertpolicy.Register(w, configService, entClient, limiter)
	uptimecheck.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPMonitoringWorkflow)
}
