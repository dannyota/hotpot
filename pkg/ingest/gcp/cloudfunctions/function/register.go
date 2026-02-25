package function

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entcloudfunctions "github.com/dannyota/hotpot/pkg/storage/ent/gcp/cloudfunctions"
)

// Register registers Cloud Functions function workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcloudfunctions.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestCloudFunctionsFunctions)
	w.RegisterWorkflow(GCPCloudFunctionsFunctionWorkflow)
}
