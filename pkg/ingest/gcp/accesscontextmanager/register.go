package accesscontextmanager

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/accesslevel"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/accesspolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/serviceperimeter"
	"entgo.io/ent/dialect"
	entaccesscontextmanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/accesscontextmanager"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
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
