package cluster

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entdataproc "github.com/dannyota/hotpot/pkg/storage/ent/gcp/dataproc"
)

// Register registers Dataproc cluster workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entdataproc.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestDataprocClusters)
	w.RegisterWorkflow(GCPDataprocClusterWorkflow)
}
