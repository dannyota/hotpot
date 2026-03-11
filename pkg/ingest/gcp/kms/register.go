package kms

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/kms/cryptokey"
	"danny.vn/hotpot/pkg/ingest/gcp/kms/keyring"
	"entgo.io/ent/dialect"
	entkms "danny.vn/hotpot/pkg/storage/ent/gcp/kms"
)

// Register registers all KMS activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entkms.NewClient(entkms.Driver(driver), entkms.AlternateSchema(entkms.DefaultSchemaConfig()))
	keyring.Register(w, configService, entClient, limiter)
	cryptokey.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPKMSWorkflow)
}
