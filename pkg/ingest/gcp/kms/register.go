package kms

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms/cryptokey"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms/keyring"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all KMS activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	keyring.Register(w, configService, entClient, limiter)
	cryptokey.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPKMSWorkflow)
}
