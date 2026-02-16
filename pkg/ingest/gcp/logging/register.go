package logging

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logbucket"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logexclusion"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logmetric"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/sink"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Logging activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	sink.Register(w, configService, entClient, limiter)
	logbucket.Register(w, configService, entClient, limiter)
	logmetric.Register(w, configService, entClient, limiter)
	logexclusion.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPLoggingWorkflow)
}
