package cluster

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entcontainer "danny.vn/hotpot/pkg/storage/ent/gcp/container"
)

// Register registers cluster activities and workflows with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, entClient *entcontainer.Client, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, entClient, limiter)

	// Register activities
	w.RegisterActivity(activities.IngestContainerClusters)

	// Register workflows
	w.RegisterWorkflow(GCPContainerClusterWorkflow)
}
