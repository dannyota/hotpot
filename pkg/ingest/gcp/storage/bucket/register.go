package bucket

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entstorage "github.com/dannyota/hotpot/pkg/storage/ent/gcp/storage"
)

// Register registers bucket workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entstorage.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestStorageBuckets)
	w.RegisterWorkflow(GCPStorageBucketWorkflow)
}
