package lifecycle

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	entlifecycle "github.com/dannyota/hotpot/pkg/storage/ent/lifecycle"
)

// Register wires lifecycle detection activities and workflow to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB) {
	entClient := entlifecycle.NewClient(
		entlifecycle.Driver(driver),
		entlifecycle.AlternateSchema(entlifecycle.DefaultSchemaConfig()),
	)

	activities := NewActivities(configService, entClient, db)
	w.RegisterActivity(activities.MatchProducts)
	w.RegisterActivity(activities.ClassifyOSCore)
	w.RegisterActivity(activities.MarkUnmatched)
	w.RegisterActivity(activities.CleanupStale)
	w.RegisterWorkflow(SoftwareLifecycleWorkflow)
}
