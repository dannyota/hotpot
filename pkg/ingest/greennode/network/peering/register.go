package peering

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entnet "danny.vn/hotpot/pkg/storage/ent/greennode/network"
)

// Register registers peering workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entnet.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestNetworkPeerings)
	w.RegisterWorkflow(GreenNodeNetworkPeeringWorkflow)
}
