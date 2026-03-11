package apicatalog

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	entapicatalog "danny.vn/hotpot/pkg/storage/ent/apicatalog"
)

// Register registers API catalog activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver) {
	entClient := entapicatalog.NewClient(
		entapicatalog.Driver(driver),
		entapicatalog.AlternateSchema(entapicatalog.DefaultSchemaConfig()),
	)

	activities := NewActivities(configService, entClient)
	w.RegisterActivity(activities.ImportCSV)
	w.RegisterWorkflow(ApiCatalogWorkflow)
}
