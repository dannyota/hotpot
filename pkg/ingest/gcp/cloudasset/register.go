package cloudasset

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/asset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/iampolicysearch"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/resourcesearch"
	"entgo.io/ent/dialect"
	entcloudasset "github.com/dannyota/hotpot/pkg/storage/ent/gcp/cloudasset"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all Cloud Asset Inventory activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entcloudasset.NewClient(entcloudasset.Driver(driver), entcloudasset.AlternateSchema(entcloudasset.DefaultSchemaConfig()))
	rmClient := entresourcemanager.NewClient(entresourcemanager.Driver(driver), entresourcemanager.AlternateSchema(entresourcemanager.DefaultSchemaConfig()))
	asset.Register(w, configService, entClient, rmClient, limiter)
	iampolicysearch.Register(w, configService, entClient, rmClient, limiter)
	resourcesearch.Register(w, configService, entClient, rmClient, limiter)

	w.RegisterWorkflow(GCPCloudAssetWorkflow)
}
