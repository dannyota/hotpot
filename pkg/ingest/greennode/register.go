package greennode

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/dns"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:greennode",
		ReqPerMin:   configService.GreenNodeRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverRegions)
	w.RegisterActivity(activities.DiscoverProjects)

	portal.Register(w, configService, entClient, activities.iamAuth, limiter)
	compute.Register(w, configService, entClient, activities.iamAuth, limiter)
	network.Register(w, configService, entClient, activities.iamAuth, limiter)
	volume.Register(w, configService, entClient, activities.iamAuth, limiter)
	loadbalancer.Register(w, configService, entClient, activities.iamAuth, limiter)
	glb.Register(w, configService, entClient, activities.iamAuth, limiter)
	dns.Register(w, configService, entClient, activities.iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeInventoryWorkflow)

	return rateLimitSvc
}
