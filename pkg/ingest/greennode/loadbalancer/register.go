package loadbalancer

import (
	"entgo.io/ent/dialect"
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer/certificate"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer/lb"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer/lbpackage"
	entlb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// Register registers all GreenNode load balancer activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entlb.NewClient(entlb.Driver(driver), entlb.AlternateSchema(entlb.DefaultSchemaConfig()))

	lb.Register(w, configService, entClient, iamAuth, limiter)
	certificate.Register(w, configService, entClient, iamAuth, limiter)
	lbpackage.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeLoadBalancerWorkflow)
}
