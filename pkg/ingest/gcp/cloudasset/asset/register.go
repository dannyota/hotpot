package asset

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entcloudasset "github.com/dannyota/hotpot/pkg/storage/ent/gcp/cloudasset"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all Cloud Asset asset activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entcloudasset.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestAssets)
	w.RegisterWorkflow(GCPCloudAssetAssetWorkflow)
}
