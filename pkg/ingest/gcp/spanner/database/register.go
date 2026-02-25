package database

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entspanner "github.com/dannyota/hotpot/pkg/storage/ent/gcp/spanner"
)

// Register registers all Spanner database activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entspanner.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestSpannerDatabases)
	w.RegisterWorkflow(GCPSpannerDatabaseWorkflow)
}
