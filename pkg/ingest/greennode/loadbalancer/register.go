package loadbalancer

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer/certificate"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer/lb"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer/lbpackage"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode load balancer activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	lb.Register(w, configService, entClient, iamAuth, limiter)
	certificate.Register(w, configService, entClient, iamAuth, limiter)
	lbpackage.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeLoadBalancerWorkflow)
}
