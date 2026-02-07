package disk

import (
	"go.temporal.io/sdk/worker"
	"hotpot/pkg/base/ratelimit"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers disk workflows and activities with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, db, limiter)
	w.RegisterActivity(activities.IngestComputeDisks)
	w.RegisterWorkflow(GCPComputeDiskWorkflow)
}
