package serviceusage

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/serviceusage/enabledservice"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Service Usage activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	enabledservice.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPServiceUsageWorkflow)
}
