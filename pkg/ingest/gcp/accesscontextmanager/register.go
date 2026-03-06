package accesscontextmanager

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/accesscontextmanager/accesslevel"
	"danny.vn/hotpot/pkg/ingest/gcp/accesscontextmanager/accesspolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/accesscontextmanager/serviceperimeter"
	"entgo.io/ent/dialect"
	entaccesscontextmanager "danny.vn/hotpot/pkg/storage/ent/gcp/accesscontextmanager"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all Access Context Manager activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entaccesscontextmanager.NewClient(entaccesscontextmanager.Driver(driver), entaccesscontextmanager.AlternateSchema(entaccesscontextmanager.DefaultSchemaConfig()))
	rmClient := entresourcemanager.NewClient(entresourcemanager.Driver(driver), entresourcemanager.AlternateSchema(entresourcemanager.DefaultSchemaConfig()))
	accesspolicy.Register(w, configService, entClient, rmClient, limiter)
	accesslevel.Register(w, configService, entClient, limiter)
	serviceperimeter.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPAccessContextManagerWorkflow)
}
