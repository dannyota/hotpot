package instance

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entspanner "danny.vn/hotpot/pkg/storage/ent/gcp/spanner"
)

// Register registers all Spanner instance activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entspanner.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestSpannerInstances)
	w.RegisterWorkflow(GCPSpannerInstanceWorkflow)
}
