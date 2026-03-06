package instance

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entec2 "danny.vn/hotpot/pkg/storage/ent/aws/ec2"
)

// Register registers instance activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entec2.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestEC2Instances)

	w.RegisterWorkflow(AWSEC2InstanceWorkflow)
}
