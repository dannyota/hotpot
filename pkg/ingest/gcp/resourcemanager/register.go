package resourcemanager

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/ingest/gcp/resourcemanager/project"
)

// Register registers all Resource Manager activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
	// Register sub-packages with config service
	project.Register(w, configService, db)
	// folder.Register(w, configService, db)        // future
	// organization.Register(w, configService, db)  // future

	// Register resource manager workflow
	w.RegisterWorkflow(GCPResourceManagerWorkflow)
}
