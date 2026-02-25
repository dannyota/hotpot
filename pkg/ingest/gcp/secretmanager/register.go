package secretmanager

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/secretmanager/secret"
	"entgo.io/ent/dialect"
	entsecretmanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/secretmanager"
)

// Register registers all Secret Manager activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entsecretmanager.NewClient(entsecretmanager.Driver(driver), entsecretmanager.AlternateSchema(entsecretmanager.DefaultSchemaConfig()))
	secret.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPSecretManagerWorkflow)
}
