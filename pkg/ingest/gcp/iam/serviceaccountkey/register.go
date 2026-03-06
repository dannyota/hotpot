package serviceaccountkey

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entiam "danny.vn/hotpot/pkg/storage/ent/gcp/iam"
)

func Register(w worker.Worker, configService *config.Service, entClient *entiam.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestIAMServiceAccountKeys)
	w.RegisterWorkflow(GCPIAMServiceAccountKeyWorkflow)
}
