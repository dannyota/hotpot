package secretmanager

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/secretmanager/secret"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Secret Manager activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	secret.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPSecretManagerWorkflow)
}
