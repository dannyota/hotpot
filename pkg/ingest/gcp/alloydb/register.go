package alloydb

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/alloydb/cluster"
	"entgo.io/ent/dialect"
	entalloydb "danny.vn/hotpot/pkg/storage/ent/gcp/alloydb"
)

// Register registers all AlloyDB activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entalloydb.NewClient(entalloydb.Driver(driver), entalloydb.AlternateSchema(entalloydb.DefaultSchemaConfig()))
	cluster.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPAlloyDBWorkflow)
}
