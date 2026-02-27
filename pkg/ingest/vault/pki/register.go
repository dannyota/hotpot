package pki

import (
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/vault/pki/certificate"
	entpki "github.com/dannyota/hotpot/pkg/storage/ent/vault/pki"
)

// Register registers PKI activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entpki.NewClient(entpki.Driver(driver), entpki.AlternateSchema(entpki.DefaultSchemaConfig()))

	// Register certificate resource
	certificate.Register(w, configService, entClient, limiter)

	// Register PKI-level activities
	activities := NewActivities(configService, limiter)
	w.RegisterActivity(activities.DiscoverMounts)

	// Register PKI workflow
	w.RegisterWorkflow(VaultPKIWorkflow)
}
