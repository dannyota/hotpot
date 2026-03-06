package certificate

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entpki "danny.vn/hotpot/pkg/storage/ent/vault/pki"
)

// Register registers certificate activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entpki.Client, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, entClient, limiter)

	// Register activities
	w.RegisterActivity(activities.IngestCertificates)

	// Register workflows
	w.RegisterWorkflow(VaultPKICertificateWorkflow)
}
