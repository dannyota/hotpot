package resourcemanager

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/resourcemanager/folder"
	"danny.vn/hotpot/pkg/ingest/gcp/resourcemanager/folderiampolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/resourcemanager/orgiampolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/resourcemanager/organization"
	"danny.vn/hotpot/pkg/ingest/gcp/resourcemanager/project"
	"danny.vn/hotpot/pkg/ingest/gcp/resourcemanager/projectiampolicy"
	"entgo.io/ent/dialect"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all Resource Manager activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entresourcemanager.NewClient(entresourcemanager.Driver(driver), entresourcemanager.AlternateSchema(entresourcemanager.DefaultSchemaConfig()))
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
