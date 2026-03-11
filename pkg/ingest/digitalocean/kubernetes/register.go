package kubernetes

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entdo "danny.vn/hotpot/pkg/storage/ent/do"
)

// Register registers Kubernetes activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entdo.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestDOKubernetesClusters)
	w.RegisterActivity(activities.IngestDOKubernetesNodePools)

	w.RegisterWorkflow(DOKubernetesWorkflow)
}
