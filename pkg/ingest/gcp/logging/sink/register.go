package sink

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entlogging "danny.vn/hotpot/pkg/storage/ent/gcp/logging"
)

// Register registers sink workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entlogging.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestLoggingSinks)
	w.RegisterWorkflow(GCPLoggingSinkWorkflow)
}
