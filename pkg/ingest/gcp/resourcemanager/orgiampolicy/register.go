package orgiampolicy

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all organization IAM policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestOrgIamPolicies)
	w.RegisterWorkflow(GCPResourceManagerOrgIamPolicyWorkflow)
}
