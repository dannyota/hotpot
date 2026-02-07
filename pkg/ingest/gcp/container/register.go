package container

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/container/cluster"
)

// Register registers all Container Engine activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	// Register sub-packages with config service
	cluster.Register(w, configService, db, limiter)

	// Register container workflow
	w.RegisterWorkflow(GCPContainerWorkflow)
}
