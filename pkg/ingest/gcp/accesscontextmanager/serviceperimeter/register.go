package serviceperimeter

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entaccesscontextmanager "danny.vn/hotpot/pkg/storage/ent/gcp/accesscontextmanager"
)

// Register registers all service perimeter activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entaccesscontextmanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestServicePerimeters)
	w.RegisterWorkflow(GCPAccessContextManagerServicePerimeterWorkflow)
}
