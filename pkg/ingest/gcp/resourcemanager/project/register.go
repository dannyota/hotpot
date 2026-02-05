package project

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers project activities and workflows with the Temporal worker.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
	// Create activities with config service (client created per session)
	activities := NewActivities(configService, db)

	// Register activities
	w.RegisterActivity(activities.IngestProjects)
	w.RegisterActivity(activities.CloseSessionClient)

	// Register workflows
	w.RegisterWorkflow(GCPResourceManagerProjectWorkflow)
}
