package storage

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/storage/bucket"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Storage activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	bucket.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPStorageWorkflow)
}
