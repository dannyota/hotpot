package software

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/meec"
	entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"
)

// Register registers software activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entinventory.Client, limiter ratelimit.Limiter, tokenSource *meec.TokenSource) {
	activities := NewActivities(configService, entClient, limiter, tokenSource)

	w.RegisterActivity(activities.IngestSoftware)

	w.RegisterWorkflow(MEECSoftwareWorkflow)
}
