package containeranalysis

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis/note"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis/occurrence"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Container Analysis activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	note.Register(w, configService, entClient, limiter)
	occurrence.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPContainerAnalysisWorkflow)
}
