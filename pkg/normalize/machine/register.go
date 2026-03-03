package machine

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	entmachine "github.com/dannyota/hotpot/pkg/storage/ent/machine"
)

// Register wires machine normalize activities and workflow to the worker.
// Providers are passed in to avoid import cycles (sub-packages import machine).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB, providers []Provider) {
	entClient := entmachine.NewClient(
		entmachine.Driver(driver),
		entmachine.AlternateSchema(entmachine.DefaultSchemaConfig()),
	)

	activities := NewActivities(configService, entClient, db, providers)
	w.RegisterActivity(activities.NormalizeProvider)
	w.RegisterActivity(activities.MergeMachines)
	w.RegisterWorkflow(NormalizeMachinesWorkflow)
}
