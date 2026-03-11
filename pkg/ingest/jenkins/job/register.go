package job

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entjenkins "danny.vn/hotpot/pkg/storage/ent/jenkins"
)

// Register registers job activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entjenkins.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestJenkinsJobs)

	w.RegisterWorkflow(JenkinsJobWorkflow)
}
