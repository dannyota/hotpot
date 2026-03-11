package constraint

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entorgpolicy "danny.vn/hotpot/pkg/storage/ent/gcp/orgpolicy"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

func Register(w worker.Worker, configService *config.Service, entClient *entorgpolicy.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestConstraints)
	w.RegisterWorkflow(GCPOrgPolicyConstraintWorkflow)
}
