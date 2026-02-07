package gcp

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/compute"
	"hotpot/pkg/ingest/gcp/container"
	"hotpot/pkg/ingest/gcp/resourcemanager"
)

// Register registers all GCP activities and workflows with the Temporal worker.
// No cleanup needed - clients are managed per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB) {
	// Create shared rate limiter for all GCP API calls
	limiter := ratelimit.NewLimiter(configService.GCPRateLimitPerMinute())

	// Register resource manager (project discovery)
	resourcemanager.Register(w, configService, db, limiter)

	// Register compute (instances, disks, networks)
	compute.Register(w, configService, db, limiter)

	// Register container (GKE clusters)
	container.Register(w, configService, db, limiter)

	// Register IAM (future)
	// iam.Register(w, configService, db, limiter)

	// Register GCP inventory workflow
	w.RegisterWorkflow(GCPInventoryWorkflow)
}
