package targetvpngateway

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

// Register registers target VPN gateway activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, db, limiter)
	w.RegisterActivity(activities.IngestComputeTargetVpnGateways)
	w.RegisterWorkflow(GCPComputeTargetVpnGatewayWorkflow)
}
