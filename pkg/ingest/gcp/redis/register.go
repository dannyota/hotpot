package redis

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/redis/instance"
	"entgo.io/ent/dialect"
	entredis "danny.vn/hotpot/pkg/storage/ent/gcp/redis"
)

// Register registers all Memorystore Redis activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entredis.NewClient(entredis.Driver(driver), entredis.AlternateSchema(entredis.DefaultSchemaConfig()))
	instance.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPRedisWorkflow)
}
