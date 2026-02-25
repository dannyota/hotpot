package connector

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
	entvpcaccess "github.com/dannyota/hotpot/pkg/storage/ent/gcp/vpcaccess"
)

// Register registers connector activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entvpcaccess.Client, computeClient *entcompute.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, computeClient, limiter)
	w.RegisterActivity(activities.IngestVpcAccessConnectors)
	w.RegisterWorkflow(GCPVpcAccessConnectorWorkflow)
}
