package lifecycle

import (
	"database/sql"

	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
)

// Register wires lifecycle detection activities and workflow to the worker.
func Register(w worker.Worker, configService *config.Service, db *sql.DB) {
	activities := NewActivities(configService, db)
	w.RegisterActivity(activities.MatchProducts)
	w.RegisterActivity(activities.ClassifyOSCore)
	w.RegisterActivity(activities.MarkUnmatched)
	w.RegisterActivity(activities.CleanupStale)
	w.RegisterWorkflow(SoftwareLifecycleWorkflow)

	w.RegisterActivity(activities.MatchOSLifecycle)
	w.RegisterActivity(activities.CleanupStaleOS)
	w.RegisterWorkflow(OSLifecycleWorkflow)
}
