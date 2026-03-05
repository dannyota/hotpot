package installedsoftware

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	entinstalledsoftware "github.com/dannyota/hotpot/pkg/storage/ent/installedsoftware"
)

// Register wires installed software normalize activities and workflow to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB, providers []Provider) {
	entClient := entinstalledsoftware.NewClient(
		entinstalledsoftware.Driver(driver),
		entinstalledsoftware.AlternateSchema(entinstalledsoftware.DefaultSchemaConfig()),
	)

	activities := NewActivities(configService, entClient, db, providers)
	w.RegisterActivity(activities.NormalizeInstalledSoftwareProvider)
	w.RegisterActivity(activities.MergeInstalledSoftware)
	w.RegisterWorkflow(NormalizeInstalledSoftwareWorkflow)
}
