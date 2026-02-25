package uptimecheck

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entmonitoring "github.com/dannyota/hotpot/pkg/storage/ent/gcp/monitoring"
)

// Register registers all Monitoring uptime check config activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entmonitoring.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestUptimeChecks)
	w.RegisterWorkflow(GCPMonitoringUptimeCheckWorkflow)
}
