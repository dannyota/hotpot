package vpcaccess

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/vpcaccess/connector"
	"entgo.io/ent/dialect"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
	entvpcaccess "danny.vn/hotpot/pkg/storage/ent/gcp/vpcaccess"
)

// Register registers all VPC Access activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entvpcaccess.NewClient(entvpcaccess.Driver(driver), entvpcaccess.AlternateSchema(entvpcaccess.DefaultSchemaConfig()))
	computeClient := entcompute.NewClient(entcompute.Driver(driver), entcompute.AlternateSchema(entcompute.DefaultSchemaConfig()))
	connector.Register(w, configService, entClient, computeClient, limiter)

	w.RegisterWorkflow(GCPVpcAccessWorkflow)
}
