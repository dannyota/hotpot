package cluster

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entbigtable "danny.vn/hotpot/pkg/storage/ent/gcp/bigtable"
)

// Register registers all Bigtable cluster activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entbigtable.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestBigtableClusters)
	w.RegisterWorkflow(GCPBigtableClusterWorkflow)
}
