package glbresource

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
)

// Register registers GLB workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entglb.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestGLBGlobalLoadBalancers)
	w.RegisterWorkflow(GreenNodeGLBGlobalLoadBalancerWorkflow)
}
