package ec2

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/aws/ec2/instance"
	entec2 "github.com/dannyota/hotpot/pkg/storage/ent/aws/ec2"
)

// Register registers all EC2 activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entec2.NewClient(entec2.Driver(driver), entec2.AlternateSchema(entec2.DefaultSchemaConfig()))

	// Register instance sub-package
	instance.Register(w, configService, entClient, limiter)

	// Register EC2 workflow
	w.RegisterWorkflow(AWSEC2Workflow)
}
