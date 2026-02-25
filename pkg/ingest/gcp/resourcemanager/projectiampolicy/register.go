package projectiampolicy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all project IAM policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestProjectIamPolicy)
	w.RegisterWorkflow(GCPResourceManagerProjectIamPolicyWorkflow)
}
