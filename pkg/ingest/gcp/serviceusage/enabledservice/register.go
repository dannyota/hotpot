package enabledservice

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entserviceusage "github.com/dannyota/hotpot/pkg/storage/ent/gcp/serviceusage"
)

// Register registers enabled service workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entserviceusage.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestEnabledServices)
	w.RegisterWorkflow(GCPServiceUsageEnabledServiceWorkflow)
}
