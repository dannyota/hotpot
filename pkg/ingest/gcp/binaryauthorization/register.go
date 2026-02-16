package binaryauthorization

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization/attestor"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization/policy"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Binary Authorization activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	policy.Register(w, configService, entClient, limiter)
	attestor.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBinaryAuthorizationWorkflow)
}
