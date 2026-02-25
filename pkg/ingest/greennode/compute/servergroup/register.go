package servergroup

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
)

// Register registers server group workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestComputeServerGroups)
	w.RegisterWorkflow(GreenNodeComputeServerGroupWorkflow)
}
