package resourcemanager

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/folder"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/folderiampolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/orgiampolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/organization"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/project"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/projectiampolicy"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Resource Manager activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Register sub-packages with config service
	project.Register(w, configService, entClient, limiter)
	organization.Register(w, configService, entClient, limiter)
	folder.Register(w, configService, entClient, limiter)
	orgiampolicy.Register(w, configService, entClient, limiter)
	folderiampolicy.Register(w, configService, entClient, limiter)
	projectiampolicy.Register(w, configService, entClient, limiter)

	// Register resource manager workflow
	w.RegisterWorkflow(GCPResourceManagerWorkflow)
}
