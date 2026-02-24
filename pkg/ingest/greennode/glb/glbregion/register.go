package glbregion

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers global region workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestGLBGlobalRegions)
	w.RegisterWorkflow(GreenNodeGLBGlobalRegionWorkflow)
}
