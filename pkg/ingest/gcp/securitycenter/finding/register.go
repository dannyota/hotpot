package finding

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
	entsecuritycenter "github.com/dannyota/hotpot/pkg/storage/ent/gcp/securitycenter"
)

// Register registers all SCC finding activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entsecuritycenter.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestFindings)
	w.RegisterWorkflow(GCPSecurityCenterFindingWorkflow)
}
