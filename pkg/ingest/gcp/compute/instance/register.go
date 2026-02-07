package instance

import (
	"go.temporal.io/sdk/worker"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers instance activities and workflows with the Temporal worker.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter *rate.Limiter) {
	// Create activities with config service (client created per session)
	activities := NewActivities(configService, db, limiter)

	// Register activities
	w.RegisterActivity(activities.IngestComputeInstances)
	w.RegisterActivity(activities.CloseSessionClient)

	// Register workflows
	w.RegisterWorkflow(GCPComputeInstanceWorkflow)
}
