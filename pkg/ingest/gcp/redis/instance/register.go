package instance

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entredis "danny.vn/hotpot/pkg/storage/ent/gcp/redis"
)

// Register registers Redis instance workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entredis.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestRedisInstances)
	w.RegisterWorkflow(GCPRedisInstanceWorkflow)
}
