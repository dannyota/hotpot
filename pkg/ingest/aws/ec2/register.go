package ec2

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/aws/ec2/instance"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all EC2 activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Register instance sub-package
	instance.Register(w, configService, entClient, limiter)

	// Register EC2 workflow
	w.RegisterWorkflow(AWSEC2Workflow)
}
