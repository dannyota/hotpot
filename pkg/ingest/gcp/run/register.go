package run

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/run/revision"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/run/service"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Cloud Run activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	service.Register(w, configService, entClient, limiter)
	revision.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPRunWorkflow)
}
