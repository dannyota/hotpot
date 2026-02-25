package iampolicy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entiap "github.com/dannyota/hotpot/pkg/storage/ent/gcp/iap"
)

// Register registers all IAP IAM policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entiap.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestIAPIAMPolicy)
	w.RegisterWorkflow(GCPIAPIAMPolicyWorkflow)
}
