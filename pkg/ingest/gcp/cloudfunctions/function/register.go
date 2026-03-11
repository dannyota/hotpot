package function

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcloudfunctions "danny.vn/hotpot/pkg/storage/ent/gcp/cloudfunctions"
)

// Register registers Cloud Functions function workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcloudfunctions.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestCloudFunctionsFunctions)
	w.RegisterWorkflow(GCPCloudFunctionsFunctionWorkflow)
}
