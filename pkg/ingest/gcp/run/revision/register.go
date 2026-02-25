package revision

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entrun "github.com/dannyota/hotpot/pkg/storage/ent/gcp/run"
)

// Register registers all Cloud Run revision activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entrun.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestRunRevisions)
	w.RegisterWorkflow(GCPRunRevisionWorkflow)
}
