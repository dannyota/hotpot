package snapshot

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
)

// Register registers snapshot activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) {
	temporalClient := configService.TemporalClient().(client.Client)

	activities := NewActivities(configService, entClient, limiter, temporalClient)
	w.RegisterActivity(activities.FetchAndSaveSnapshotsPage)
	w.RegisterActivity(activities.DeleteStaleSnapshots)
	w.RegisterWorkflow(GCPComputeSnapshotWorkflow)
}
