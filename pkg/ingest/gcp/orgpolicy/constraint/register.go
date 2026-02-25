package constraint

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entorgpolicy "github.com/dannyota/hotpot/pkg/storage/ent/gcp/orgpolicy"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

func Register(w worker.Worker, configService *config.Service, entClient *entorgpolicy.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestConstraints)
	w.RegisterWorkflow(GCPOrgPolicyConstraintWorkflow)
}
