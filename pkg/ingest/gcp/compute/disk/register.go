package disk

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers disk workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
	activities := NewActivities(configService, db)
	w.RegisterActivity(activities.IngestComputeDisks)
	w.RegisterActivity(activities.CloseSessionClient)
	w.RegisterWorkflow(GCPComputeDiskWorkflow)
}
