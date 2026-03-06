package iampolicysearch

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcloudasset "danny.vn/hotpot/pkg/storage/ent/gcp/cloudasset"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all IAM policy search activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entcloudasset.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestIAMPolicySearch)
	w.RegisterWorkflow(GCPCloudAssetIAMPolicySearchWorkflow)
}
