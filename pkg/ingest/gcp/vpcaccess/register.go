package vpcaccess

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/vpcaccess/connector"
	"entgo.io/ent/dialect"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
	entvpcaccess "github.com/dannyota/hotpot/pkg/storage/ent/gcp/vpcaccess"
)

// Register registers all VPC Access activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entvpcaccess.NewClient(entvpcaccess.Driver(driver), entvpcaccess.AlternateSchema(entvpcaccess.DefaultSchemaConfig()))
	computeClient := entcompute.NewClient(entcompute.Driver(driver), entcompute.AlternateSchema(entcompute.DefaultSchemaConfig()))
	connector.Register(w, configService, entClient, computeClient, limiter)

	w.RegisterWorkflow(GCPVpcAccessWorkflow)
}
