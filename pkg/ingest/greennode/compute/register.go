package compute

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/server"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/servergroup"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/sshkey"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode compute activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	server.Register(w, configService, entClient, limiter)
	sshkey.Register(w, configService, entClient, limiter)
	servergroup.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GreenNodeComputeWorkflow)
}
