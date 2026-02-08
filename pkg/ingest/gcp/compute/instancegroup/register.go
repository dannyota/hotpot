package instancegroup

import (
	"go.temporal.io/sdk/worker"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/storage/ent"
)

// Register registers instance group workflows and activities with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestComputeInstanceGroups)
	w.RegisterWorkflow(GCPComputeInstanceGroupWorkflow)
}
