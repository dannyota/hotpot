package bigquery

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery/dataset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery/table"
	"entgo.io/ent/dialect"
	entbigquery "github.com/dannyota/hotpot/pkg/storage/ent/gcp/bigquery"
)

// Register registers all BigQuery activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entbigquery.NewClient(entbigquery.Driver(driver), entbigquery.AlternateSchema(entbigquery.DefaultSchemaConfig()))
	dataset.Register(w, configService, entClient, limiter)
	table.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBigQueryWorkflow)
}
