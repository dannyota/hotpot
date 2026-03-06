package accesspolicy

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entaccesscontextmanager "danny.vn/hotpot/pkg/storage/ent/gcp/accesscontextmanager"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all access policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entaccesscontextmanager.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestAccessPolicies)
	w.RegisterWorkflow(GCPAccessContextManagerAccessPolicyWorkflow)
}
