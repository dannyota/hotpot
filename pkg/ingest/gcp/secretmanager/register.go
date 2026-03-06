package secretmanager

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/secretmanager/secret"
	"entgo.io/ent/dialect"
	entsecretmanager "danny.vn/hotpot/pkg/storage/ent/gcp/secretmanager"
)

// Register registers all Secret Manager activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entsecretmanager.NewClient(entsecretmanager.Driver(driver), entsecretmanager.AlternateSchema(entsecretmanager.DefaultSchemaConfig()))
	secret.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPSecretManagerWorkflow)
}
