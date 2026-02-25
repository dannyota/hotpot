package cryptokey

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entkms "github.com/dannyota/hotpot/pkg/storage/ent/gcp/kms"
)

// Register registers crypto key workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entkms.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestKMSCryptoKeys)
	w.RegisterWorkflow(GCPKMSCryptoKeyWorkflow)
}
