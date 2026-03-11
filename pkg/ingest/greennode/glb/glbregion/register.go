package glbregion

import (
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entglb "danny.vn/hotpot/pkg/storage/ent/greennode/glb"
)

// Register registers global region workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entglb.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestGLBGlobalRegions)
	w.RegisterWorkflow(GreenNodeGLBGlobalRegionWorkflow)
}
