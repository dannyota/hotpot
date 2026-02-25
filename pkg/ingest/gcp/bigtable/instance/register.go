package instance

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entbigtable "github.com/dannyota/hotpot/pkg/storage/ent/gcp/bigtable"
)

// Register registers all Bigtable instance activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entbigtable.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestBigtableInstances)
	w.RegisterWorkflow(GCPBigtableInstanceWorkflow)
}
