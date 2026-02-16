package cloudasset

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/asset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/iampolicysearch"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/resourcesearch"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Cloud Asset Inventory activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	asset.Register(w, configService, entClient, limiter)
	iampolicysearch.Register(w, configService, entClient, limiter)
	resourcesearch.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPCloudAssetWorkflow)
}
