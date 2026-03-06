package storage

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/storage/bucket"
	"danny.vn/hotpot/pkg/ingest/gcp/storage/bucketiam"
	"entgo.io/ent/dialect"
	entstorage "danny.vn/hotpot/pkg/storage/ent/gcp/storage"
)

// Register registers all Storage activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entstorage.NewClient(entstorage.Driver(driver), entstorage.AlternateSchema(entstorage.DefaultSchemaConfig()))
	bucket.Register(w, configService, entClient, limiter)
	bucketiam.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPStorageWorkflow)
}
