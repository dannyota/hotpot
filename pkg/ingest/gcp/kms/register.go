package kms

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms/cryptokey"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms/keyring"
	"entgo.io/ent/dialect"
	entkms "github.com/dannyota/hotpot/pkg/storage/ent/gcp/kms"
)

// Register registers all KMS activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entkms.NewClient(entkms.Driver(driver), entkms.AlternateSchema(entkms.DefaultSchemaConfig()))
	keyring.Register(w, configService, entClient, limiter)
	cryptokey.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPKMSWorkflow)
}
