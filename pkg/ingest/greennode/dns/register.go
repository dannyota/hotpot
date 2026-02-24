package dns

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/dns/hostedzone"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode DNS activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	hostedzone.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeDNSWorkflow)
}
