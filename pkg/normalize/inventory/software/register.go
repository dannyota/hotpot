package software

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	entsoftware "danny.vn/hotpot/pkg/storage/ent/inventory/software"
)

// Register wires software normalize activities and workflow to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB, providers []Provider) {
	entClient := entsoftware.NewClient(
		entsoftware.Driver(driver),
		entsoftware.AlternateSchema(entsoftware.DefaultSchemaConfig()),
	)

	activities := NewActivities(configService, entClient, db, providers)
	w.RegisterActivity(activities.NormalizeSoftwareProvider)
	w.RegisterActivity(activities.MergeSoftware)
	w.RegisterWorkflow(NormalizeSoftwareWorkflow)
}
