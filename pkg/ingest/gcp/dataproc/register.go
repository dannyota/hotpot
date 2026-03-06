package dataproc

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/dataproc/cluster"
	"entgo.io/ent/dialect"
	entdataproc "danny.vn/hotpot/pkg/storage/ent/gcp/dataproc"
)

// Register registers all Dataproc activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entdataproc.NewClient(entdataproc.Driver(driver), entdataproc.AlternateSchema(entdataproc.DefaultSchemaConfig()))
	cluster.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPDataprocWorkflow)
}
