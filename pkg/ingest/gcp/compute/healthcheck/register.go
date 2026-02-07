package healthcheck

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

// Register registers health check activities and workflows with a Temporal worker.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, db, limiter)
	w.RegisterActivity(activities.IngestComputeHealthChecks)
	w.RegisterWorkflow(GCPComputeHealthCheckWorkflow)
}
