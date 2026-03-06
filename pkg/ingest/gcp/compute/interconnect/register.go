package interconnect

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
)

// Register registers interconnect activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestComputeInterconnects)
	w.RegisterWorkflow(GCPComputeInterconnectWorkflow)
}
