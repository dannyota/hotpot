package resourcemanager

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/project"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Resource Manager activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Register sub-packages with config service
	project.Register(w, configService, entClient, limiter)
	// folder.Register(w, configService, db, limiter)        // future
	// organization.Register(w, configService, db, limiter)  // future

	// Register resource manager workflow
	w.RegisterWorkflow(GCPResourceManagerWorkflow)
}
