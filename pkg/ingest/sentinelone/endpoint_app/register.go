package endpoint_app

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
)

// Register registers endpoint app activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ents1.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.ListAgentIDs)
	w.RegisterActivity(activities.FetchAndSaveAgentApps)
	w.RegisterActivity(activities.DeleteStaleEndpointApps)

	w.RegisterWorkflow(S1EndpointAppWorkflow)
}
