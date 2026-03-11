package loadbalancer

import (
	"entgo.io/ent/dialect"
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/greennode/loadbalancer/certificate"
	"danny.vn/hotpot/pkg/ingest/greennode/loadbalancer/lb"
	"danny.vn/hotpot/pkg/ingest/greennode/loadbalancer/lbpackage"
	entlb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// Register registers all GreenNode load balancer activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entlb.NewClient(entlb.Driver(driver), entlb.AlternateSchema(entlb.DefaultSchemaConfig()))

	lb.Register(w, configService, entClient, iamAuth, limiter)
	certificate.Register(w, configService, entClient, iamAuth, limiter)
	lbpackage.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeLoadBalancerWorkflow)
}
