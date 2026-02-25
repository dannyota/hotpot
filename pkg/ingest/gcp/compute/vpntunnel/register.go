package vpntunnel

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entvpn "github.com/dannyota/hotpot/pkg/storage/ent/gcp/vpn"
)

// Register registers VPN tunnel activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entvpn.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestComputeVpnTunnels)
	w.RegisterWorkflow(GCPComputeVpnTunnelWorkflow)
}
