package keyring

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entkms "danny.vn/hotpot/pkg/storage/ent/gcp/kms"
)

// Register registers key ring workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entkms.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestKMSKeyRings)
	w.RegisterWorkflow(GCPKMSKeyRingWorkflow)
}
