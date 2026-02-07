package globalforwardingrule

import (
	"go.temporal.io/sdk/worker"
	"hotpot/pkg/base/ratelimit"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// Register registers global forwarding rule workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, db, limiter)
	w.RegisterActivity(activities.IngestComputeGlobalForwardingRules)
	w.RegisterWorkflow(GCPComputeGlobalForwardingRuleWorkflow)
}
