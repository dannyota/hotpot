package key

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entdo "danny.vn/hotpot/pkg/storage/ent/do"
)

// Register registers SSH key activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entdo.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestDOKeys)

	w.RegisterWorkflow(DOKeyWorkflow)
}
