package instance

import (
	"go.temporal.io/sdk/worker"
	"hotpot/pkg/base/ratelimit"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers instance activities and workflows with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, db, limiter)

	// Register activities
	w.RegisterActivity(activities.IngestComputeInstances)

	// Register workflows
	w.RegisterWorkflow(GCPComputeInstanceWorkflow)
}
