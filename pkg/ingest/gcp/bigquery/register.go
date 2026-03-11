package bigquery

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/bigquery/dataset"
	"danny.vn/hotpot/pkg/ingest/gcp/bigquery/table"
	"entgo.io/ent/dialect"
	entbigquery "danny.vn/hotpot/pkg/storage/ent/gcp/bigquery"
)

// Register registers all BigQuery activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entbigquery.NewClient(entbigquery.Driver(driver), entbigquery.AlternateSchema(entbigquery.DefaultSchemaConfig()))
	dataset.Register(w, configService, entClient, limiter)
	table.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBigQueryWorkflow)
}
