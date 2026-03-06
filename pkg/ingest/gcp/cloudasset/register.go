package cloudasset

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/cloudasset/asset"
	"danny.vn/hotpot/pkg/ingest/gcp/cloudasset/iampolicysearch"
	"danny.vn/hotpot/pkg/ingest/gcp/cloudasset/resourcesearch"
	"entgo.io/ent/dialect"
	entcloudasset "danny.vn/hotpot/pkg/storage/ent/gcp/cloudasset"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
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
