package droplet

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entdo "github.com/dannyota/hotpot/pkg/storage/ent/do"
)

// Register registers Droplet activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entdo.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestDODroplets)

	w.RegisterWorkflow(DODropletWorkflow)
}
