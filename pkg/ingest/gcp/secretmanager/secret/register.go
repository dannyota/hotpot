package secret

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entsecretmanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/secretmanager"
)

// Register registers secret workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entsecretmanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestSecretManagerSecrets)
	w.RegisterWorkflow(GCPSecretManagerSecretWorkflow)
}
