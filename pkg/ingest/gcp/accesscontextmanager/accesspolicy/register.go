package accesspolicy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entaccesscontextmanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/accesscontextmanager"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all access policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entaccesscontextmanager.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestAccessPolicies)
	w.RegisterWorkflow(GCPAccessContextManagerAccessPolicyWorkflow)
}
