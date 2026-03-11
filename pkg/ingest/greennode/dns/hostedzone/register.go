package hostedzone

import (
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entdns "danny.vn/hotpot/pkg/storage/ent/greennode/dns"
)

// Register registers hosted zone workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entdns.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestDNSHostedZones)
	w.RegisterWorkflow(GreenNodeDNSHostedZoneWorkflow)
}
