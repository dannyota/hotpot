package targetvpngateway

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entvpn "danny.vn/hotpot/pkg/storage/ent/gcp/vpn"
)

// Register registers target VPN gateway activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entvpn.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestComputeTargetVpnGateways)
	w.RegisterWorkflow(GCPComputeTargetVpnGatewayWorkflow)
}
