package secgroup

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
)

// Register registers security group workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entnet.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestNetworkSecgroups)
	w.RegisterWorkflow(GreenNodeNetworkSecgroupWorkflow)
}
