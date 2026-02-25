package serviceaccountkey

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entiam "github.com/dannyota/hotpot/pkg/storage/ent/gcp/iam"
)

func Register(w worker.Worker, configService *config.Service, entClient *entiam.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestIAMServiceAccountKeys)
	w.RegisterWorkflow(GCPIAMServiceAccountKeyWorkflow)
}
