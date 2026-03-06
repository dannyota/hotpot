package spanner

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/spanner/database"
	"danny.vn/hotpot/pkg/ingest/gcp/spanner/instance"
	"entgo.io/ent/dialect"
	entspanner "danny.vn/hotpot/pkg/storage/ent/gcp/spanner"
)

// Register registers all Spanner activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entspanner.NewClient(entspanner.Driver(driver), entspanner.AlternateSchema(entspanner.DefaultSchemaConfig()))
	instance.Register(w, configService, entClient, limiter)
	database.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPSpannerWorkflow)
}
