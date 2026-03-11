package instance

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entgcpsql "danny.vn/hotpot/pkg/storage/ent/gcp/sql"
)

// Register registers SQL instance activities and workflows with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, entClient *entgcpsql.Client, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, entClient, limiter)

	// Register activities
	w.RegisterActivity(activities.IngestSQLInstances)

	// Register workflows
	w.RegisterWorkflow(GCPSQLInstanceWorkflow)
}
