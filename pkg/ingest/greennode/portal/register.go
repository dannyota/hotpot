package portal

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/quota"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/region"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/zone"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode portal activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	region.Register(w, configService, entClient, iamAuth, limiter)
	quota.Register(w, configService, entClient, iamAuth, limiter)
	zone.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodePortalWorkflow)
}
