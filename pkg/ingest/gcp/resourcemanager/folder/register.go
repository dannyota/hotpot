package folder

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers folder activities and workflows with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, entClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, entClient, limiter)

	// Register activities
	w.RegisterActivity(activities.IngestFolders)

	// Register workflows
	w.RegisterWorkflow(GCPResourceManagerFolderWorkflow)
}
