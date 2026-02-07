package cluster

import (
	"go.temporal.io/sdk/worker"
	"hotpot/pkg/base/ratelimit"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers cluster activities and workflows with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, db, limiter)

	// Register activities
	w.RegisterActivity(activities.IngestContainerClusters)

	// Register workflows
	w.RegisterWorkflow(GCPContainerClusterWorkflow)
}
