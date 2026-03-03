package software

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/meec"
	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
)

// Register registers software activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entinventory.Client, limiter ratelimit.Limiter, tokenSource *meec.TokenSource) {
	activities := NewActivities(configService, entClient, limiter, tokenSource)

	w.RegisterActivity(activities.IngestSoftware)

	w.RegisterWorkflow(MEECSoftwareWorkflow)
}
