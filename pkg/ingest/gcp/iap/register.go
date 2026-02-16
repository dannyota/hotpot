package iap

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap/iampolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap/settings"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Identity-Aware Proxy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	settings.Register(w, configService, entClient, limiter)
	iampolicy.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPIAPWorkflow)
}
