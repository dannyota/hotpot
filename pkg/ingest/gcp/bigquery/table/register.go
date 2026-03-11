package table

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entbigquery "danny.vn/hotpot/pkg/storage/ent/gcp/bigquery"
)

// Register registers all BigQuery table activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entbigquery.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestBigQueryTables)
	w.RegisterWorkflow(GCPBigQueryTableWorkflow)
}
