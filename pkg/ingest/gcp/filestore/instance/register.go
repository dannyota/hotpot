package instance

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entfilestore "danny.vn/hotpot/pkg/storage/ent/gcp/filestore"
)

// Register registers Filestore instance workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entfilestore.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestFilestoreInstances)
	w.RegisterWorkflow(GCPFilestoreInstanceWorkflow)
}
