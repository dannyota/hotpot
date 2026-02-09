package account

import (
	"go.temporal.io/sdk/worker"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/storage/ent"
)

// Register registers account activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestS1Accounts)

	w.RegisterWorkflow(S1AccountWorkflow)
}
