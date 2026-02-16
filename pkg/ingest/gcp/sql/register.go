package sql

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/sql/instance"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Cloud SQL activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Register sub-packages with config service
	instance.Register(w, configService, entClient, limiter)

	// Register SQL workflow
	w.RegisterWorkflow(GCPSQLWorkflow)
}
