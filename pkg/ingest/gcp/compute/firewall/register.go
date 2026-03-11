package firewall

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
)

// Register registers firewall workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) {
	temporalClient := configService.TemporalClient().(client.Client)

	activities := NewActivities(configService, entClient, limiter, temporalClient)
	w.RegisterActivity(activities.FetchAndSaveFirewallsPage)
	w.RegisterActivity(activities.DeleteStaleFirewalls)
	w.RegisterWorkflow(GCPComputeFirewallWorkflow)
}
