package cluster

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entalloydb "github.com/dannyota/hotpot/pkg/storage/ent/gcp/alloydb"
)

// Register registers AlloyDB cluster workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entalloydb.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestAlloyDBClusters)
	w.RegisterWorkflow(GCPAlloyDBClusterWorkflow)
}
