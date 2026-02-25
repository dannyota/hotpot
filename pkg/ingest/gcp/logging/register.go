package logging

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logbucket"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logexclusion"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logmetric"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/sink"
	"entgo.io/ent/dialect"
	entlogging "github.com/dannyota/hotpot/pkg/storage/ent/gcp/logging"
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
