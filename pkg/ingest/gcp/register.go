package gcp

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/ingest/gcp/compute"
	"hotpot/pkg/ingest/gcp/resourcemanager"
)

// Register registers all GCP activities and workflows with the Temporal worker.
// No cleanup needed - clients are managed per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
	// Register resource manager (project discovery)
	resourcemanager.Register(w, configService, db)

	// Register compute (instances, disks, networks)
	compute.Register(w, configService, db)

	// Register GKE (future)
	// gke.Register(w, configService, db)

	// Register IAM (future)
	// iam.Register(w, configService, db)

	// Register GCP inventory workflow
	w.RegisterWorkflow(GCPInventoryWorkflow)
}
