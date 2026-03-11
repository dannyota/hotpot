package lb

import (
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entlb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// Register registers load balancer workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entlb.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestLoadBalancerLBs)
	w.RegisterWorkflow(GreenNodeLoadBalancerLBWorkflow)
}
