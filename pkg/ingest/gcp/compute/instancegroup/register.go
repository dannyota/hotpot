package instancegroup

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
)

// Register registers instance group workflows and activities with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestComputeInstanceGroups)
	w.RegisterWorkflow(GCPComputeInstanceGroupWorkflow)
}
