package connector

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
	entvpcaccess "danny.vn/hotpot/pkg/storage/ent/gcp/vpcaccess"
)

// Register registers connector activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entvpcaccess.Client, computeClient *entcompute.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, computeClient, limiter)
	w.RegisterActivity(activities.IngestVpcAccessConnectors)
	w.RegisterWorkflow(GCPVpcAccessConnectorWorkflow)
}
