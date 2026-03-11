package bigtable

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/bigtable/cluster"
	"danny.vn/hotpot/pkg/ingest/gcp/bigtable/instance"
	"entgo.io/ent/dialect"
	entbigtable "danny.vn/hotpot/pkg/storage/ent/gcp/bigtable"
)

// Register registers all Bigtable activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entbigtable.NewClient(entbigtable.Driver(driver), entbigtable.AlternateSchema(entbigtable.DefaultSchemaConfig()))
	instance.Register(w, configService, entClient, limiter)
	cluster.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBigtableWorkflow)
}
