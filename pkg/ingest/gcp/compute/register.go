package compute

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/ingest/gcp/compute/disk"
	"hotpot/pkg/ingest/gcp/compute/instance"
	"hotpot/pkg/ingest/gcp/compute/network"
	"hotpot/pkg/ingest/gcp/compute/subnetwork"
)

// Register registers all Compute Engine activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
	// Register sub-packages with config service
	instance.Register(w, configService, db)
	disk.Register(w, configService, db)
	network.Register(w, configService, db)
	subnetwork.Register(w, configService, db)

	// Register compute workflow
	w.RegisterWorkflow(GCPComputeWorkflow)
}
