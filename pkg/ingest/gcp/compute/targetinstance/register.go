package targetinstance

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers target instance workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
	activities := NewActivities(configService, db)
	w.RegisterActivity(activities.IngestComputeTargetInstances)
	w.RegisterActivity(activities.CloseSessionClient)
	w.RegisterWorkflow(GCPComputeTargetInstanceWorkflow)
}
