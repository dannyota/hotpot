package endpoint_app

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
)

// Register registers endpoint app activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ents1.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.ListAgentIDs)
	w.RegisterActivity(activities.FetchAndSaveBatch)
	w.RegisterActivity(activities.DeleteOrphanEndpointApps)

	w.RegisterWorkflow(S1EndpointAppWorkflow)
}
