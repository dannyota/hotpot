package osimage

import (
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcompute "danny.vn/hotpot/pkg/storage/ent/greennode/compute"
)

// Register registers OS image workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestComputeOSImages)
	w.RegisterWorkflow(GreenNodeComputeOSImageWorkflow)
}
