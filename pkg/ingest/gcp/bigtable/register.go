package bigtable

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigtable/cluster"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigtable/instance"
	"entgo.io/ent/dialect"
	entbigtable "github.com/dannyota/hotpot/pkg/storage/ent/gcp/bigtable"
)

// Register registers all Bigtable activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entbigtable.NewClient(entbigtable.Driver(driver), entbigtable.AlternateSchema(entbigtable.DefaultSchemaConfig()))
	instance.Register(w, configService, entClient, limiter)
	cluster.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBigtableWorkflow)
}
