package appservice

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entappengine "danny.vn/hotpot/pkg/storage/ent/gcp/appengine"
)

// Register registers all App Engine service activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entappengine.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestAppEngineServices)
	w.RegisterWorkflow(GCPAppEngineServiceWorkflow)
}
