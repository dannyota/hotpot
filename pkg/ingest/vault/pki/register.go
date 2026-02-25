package pki

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/vault/pki/certificate"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers PKI activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Register certificate resource
	certificate.Register(w, configService, entClient, limiter)

	// Register PKI-level activities
	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverMounts)

	// Register PKI workflow
	w.RegisterWorkflow(VaultPKIWorkflow)
}
