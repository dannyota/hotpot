package hostedzone

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entdns "github.com/dannyota/hotpot/pkg/storage/ent/greennode/dns"
)

// Register registers hosted zone workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entdns.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestDNSHostedZones)
	w.RegisterWorkflow(GreenNodeDNSHostedZoneWorkflow)
}
