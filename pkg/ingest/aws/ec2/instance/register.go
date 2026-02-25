package instance

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entec2 "github.com/dannyota/hotpot/pkg/storage/ent/aws/ec2"
)

// Register registers instance activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entec2.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestEC2Instances)

	w.RegisterWorkflow(AWSEC2InstanceWorkflow)
}
