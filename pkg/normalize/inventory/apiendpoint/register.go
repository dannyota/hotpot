package apiendpoint

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	entapiendpoint "danny.vn/hotpot/pkg/storage/ent/inventory/apiendpoint"
)

// Register wires API endpoint normalize activities and workflow to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB, providers []Provider) {
	entClient := entapiendpoint.NewClient(
		entapiendpoint.Driver(driver),
		entapiendpoint.AlternateSchema(entapiendpoint.DefaultSchemaConfig()),
	)

	activities := NewActivities(configService, entClient, db, providers)
	w.RegisterActivity(activities.NormalizeApiEndpoints)
	w.RegisterWorkflow(NormalizeApiEndpointsWorkflow)
}
