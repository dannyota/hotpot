package vpcaccess

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/vpcaccess/connector"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all VPC Access activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	connector.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPVpcAccessWorkflow)
}
