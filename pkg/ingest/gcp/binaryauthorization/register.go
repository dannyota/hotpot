package binaryauthorization

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/binaryauthorization/attestor"
	"danny.vn/hotpot/pkg/ingest/gcp/binaryauthorization/policy"
	"entgo.io/ent/dialect"
	entbinaryauthorization "danny.vn/hotpot/pkg/storage/ent/gcp/binaryauthorization"
)

// Register registers all Binary Authorization activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entbinaryauthorization.NewClient(entbinaryauthorization.Driver(driver), entbinaryauthorization.AlternateSchema(entbinaryauthorization.DefaultSchemaConfig()))
	policy.Register(w, configService, entClient, limiter)
	attestor.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBinaryAuthorizationWorkflow)
}
