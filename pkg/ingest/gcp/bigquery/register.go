package bigquery

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery/dataset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery/table"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all BigQuery activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	dataset.Register(w, configService, entClient, limiter)
	table.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPBigQueryWorkflow)
}
