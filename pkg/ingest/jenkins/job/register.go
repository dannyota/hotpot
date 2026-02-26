package job

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entjenkins "github.com/dannyota/hotpot/pkg/storage/ent/jenkins"
)

// Register registers job activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entjenkins.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestJenkinsJobs)

	w.RegisterWorkflow(JenkinsJobWorkflow)
}
