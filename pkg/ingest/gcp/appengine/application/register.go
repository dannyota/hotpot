package application

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entappengine "github.com/dannyota/hotpot/pkg/storage/ent/gcp/appengine"
)

// Register registers all App Engine application activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entappengine.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestAppEngineApplications)
	w.RegisterWorkflow(GCPAppEngineApplicationWorkflow)
}
