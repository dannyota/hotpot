package resourcemanager

import (
	"go.temporal.io/sdk/worker"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/ingest/gcp/resourcemanager/project"
)

// Register registers all Resource Manager activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter *rate.Limiter) {
	// Register sub-packages with config service
	project.Register(w, configService, db, limiter)
	// folder.Register(w, configService, db, limiter)        // future
	// organization.Register(w, configService, db, limiter)  // future

	// Register resource manager workflow
	w.RegisterWorkflow(GCPResourceManagerWorkflow)
}
