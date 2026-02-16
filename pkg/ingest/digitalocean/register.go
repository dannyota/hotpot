package digitalocean

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/account"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/database"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/domain"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/droplet"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/firewall"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/key"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/kubernetes"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/loadbalancer"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/project"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/volume"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/vpc"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all DigitalOcean activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:do",
		ReqPerMin:   configService.DORateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	account.Register(w, configService, entClient, limiter)
	database.Register(w, configService, entClient, limiter)
	domain.Register(w, configService, entClient, limiter)
	droplet.Register(w, configService, entClient, limiter)
	firewall.Register(w, configService, entClient, limiter)
	key.Register(w, configService, entClient, limiter)
	kubernetes.Register(w, configService, entClient, limiter)
	loadbalancer.Register(w, configService, entClient, limiter)
	project.Register(w, configService, entClient, limiter)
	volume.Register(w, configService, entClient, limiter)
	vpc.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(DOInventoryWorkflow)

	return rateLimitSvc
}
