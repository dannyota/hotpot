package binaryauthorization

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization/attestor"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization/policy"
	"entgo.io/ent/dialect"
	entbinaryauthorization "github.com/dannyota/hotpot/pkg/storage/ent/gcp/binaryauthorization"
)

// Register registers all Binary Authorization activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entbinaryauthorization.NewClient(entbinaryauthorization.Driver(driver), entbinaryauthorization.AlternateSchema(entbinaryauthorization.DefaultSchemaConfig()))
	policy.Register(w, configService, entClient, limiter)
	attestor.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBinaryAuthorizationWorkflow)
}
