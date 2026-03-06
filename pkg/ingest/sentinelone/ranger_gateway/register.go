package ranger_gateway

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
)

// Register registers ranger gateway activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ents1.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestS1RangerGateways)

	w.RegisterWorkflow(S1RangerGatewayWorkflow)
}
