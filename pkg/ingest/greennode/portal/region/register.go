package region

import (
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entportal "danny.vn/hotpot/pkg/storage/ent/greennode/portal"
)

// Register registers region workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entportal.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestPortalRegions)
	w.RegisterWorkflow(GreenNodePortalRegionWorkflow)
}
