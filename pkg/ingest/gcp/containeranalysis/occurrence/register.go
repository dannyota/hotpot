package occurrence

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcontaineranalysis "danny.vn/hotpot/pkg/storage/ent/gcp/containeranalysis"
)

// Register registers all Container Analysis occurrence activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entcontaineranalysis.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestOccurrences)
	w.RegisterWorkflow(GCPContainerAnalysisOccurrenceWorkflow)
}
