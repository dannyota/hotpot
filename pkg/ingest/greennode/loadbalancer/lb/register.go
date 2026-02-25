package lb

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entlb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// Register registers load balancer workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entlb.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestLoadBalancerLBs)
	w.RegisterWorkflow(GreenNodeLoadBalancerLBWorkflow)
}
