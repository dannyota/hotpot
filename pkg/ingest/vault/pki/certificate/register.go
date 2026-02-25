package certificate

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entpki "github.com/dannyota/hotpot/pkg/storage/ent/vault/pki"
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
