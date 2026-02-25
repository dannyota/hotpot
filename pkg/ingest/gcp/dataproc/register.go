package dataproc

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dataproc/cluster"
	"entgo.io/ent/dialect"
	entdataproc "github.com/dannyota/hotpot/pkg/storage/ent/gcp/dataproc"
)

// Register registers all Dataproc activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entdataproc.NewClient(entdataproc.Driver(driver), entdataproc.AlternateSchema(entdataproc.DefaultSchemaConfig()))
	cluster.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPDataprocWorkflow)
}
