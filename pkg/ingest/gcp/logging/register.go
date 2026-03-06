package logging

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/logbucket"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/logexclusion"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/logmetric"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/sink"
	"entgo.io/ent/dialect"
	entlogging "danny.vn/hotpot/pkg/storage/ent/gcp/logging"
)

// Register registers all Logging activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entlogging.NewClient(entlogging.Driver(driver), entlogging.AlternateSchema(entlogging.DefaultSchemaConfig()))
	sink.Register(w, configService, entClient, limiter)
	logbucket.Register(w, configService, entClient, limiter)
	logmetric.Register(w, configService, entClient, limiter)
	logexclusion.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPLoggingWorkflow)
}
