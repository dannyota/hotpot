package vpcaccess

import (
	"go.temporal.io/sdk/worker"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/vpcaccess/connector"
	"hotpot/pkg/storage/ent"
)

// Register registers all VPC Access activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	connector.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPVpcAccessWorkflow)
}
