package targettcpproxy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers target TCP proxy workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestComputeTargetTcpProxies)
	w.RegisterWorkflow(GCPComputeTargetTcpProxyWorkflow)
}
