package monitoring

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring/alertpolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring/uptimecheck"
	"entgo.io/ent/dialect"
	entmonitoring "github.com/dannyota/hotpot/pkg/storage/ent/gcp/monitoring"
)

// Register registers all Monitoring activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entmonitoring.NewClient(entmonitoring.Driver(driver), entmonitoring.AlternateSchema(entmonitoring.DefaultSchemaConfig()))
	alertpolicy.Register(w, configService, entClient, limiter)
	uptimecheck.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPMonitoringWorkflow)
}
