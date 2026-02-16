package filestore

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/filestore/instance"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Filestore activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	instance.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPFilestoreWorkflow)
}
