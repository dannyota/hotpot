package logbucket

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entlogging "github.com/dannyota/hotpot/pkg/storage/ent/gcp/logging"
)

// Register registers log bucket workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entlogging.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestLoggingBuckets)
	w.RegisterWorkflow(GCPLoggingBucketWorkflow)
}
