package container

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/container/cluster"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Container Engine activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Register sub-packages with config service
	cluster.Register(w, configService, entClient, limiter)

	// Register container workflow
	w.RegisterWorkflow(GCPContainerWorkflow)
}
