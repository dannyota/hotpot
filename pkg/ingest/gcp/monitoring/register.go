package monitoring

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring/alertpolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring/uptimecheck"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Monitoring activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	alertpolicy.Register(w, configService, entClient, limiter)
	uptimecheck.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPMonitoringWorkflow)
}
