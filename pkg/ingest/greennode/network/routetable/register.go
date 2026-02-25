package routetable

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
)

// Register registers route table workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entnet.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestNetworkRouteTables)
	w.RegisterWorkflow(GreenNodeNetworkRouteTableWorkflow)
}
