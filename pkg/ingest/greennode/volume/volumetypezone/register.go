package volumetypezone

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entvol "danny.vn/hotpot/pkg/storage/ent/greennode/volume"
)

// Register registers volume type zone workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entvol.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestVolumeVolumeTypeZones)
	w.RegisterWorkflow(GreenNodeVolumeVolumeTypeZoneWorkflow)
}
