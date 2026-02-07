package instancegroup

import (
	"go.temporal.io/sdk/worker"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers instance group workflows and activities with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter *rate.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, db, limiter)
	w.RegisterActivity(activities.IngestComputeInstanceGroups)
	w.RegisterWorkflow(GCPComputeInstanceGroupWorkflow)
}
