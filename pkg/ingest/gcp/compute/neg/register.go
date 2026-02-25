package neg

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
)

func Register(w worker.Worker, configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestComputeNegs)
	w.RegisterWorkflow(GCPComputeNegWorkflow)
}
