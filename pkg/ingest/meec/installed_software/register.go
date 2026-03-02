package installed_software

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/meec"
	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
)

// Register registers installed software activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entinventory.Client, limiter ratelimit.Limiter, tokenSource *meec.TokenSource) {
	activities := NewActivities(configService, entClient, limiter, tokenSource)

	w.RegisterActivity(activities.ListComputerIDs)
	w.RegisterActivity(activities.FetchAndSaveBatch)
	w.RegisterActivity(activities.DeleteStaleInstalledSoftware)

	w.RegisterWorkflow(MEECInstalledSoftwareWorkflow)
}
