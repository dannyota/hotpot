package storage

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/storage/bucket"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/storage/bucketiam"
	"entgo.io/ent/dialect"
	entstorage "github.com/dannyota/hotpot/pkg/storage/ent/gcp/storage"
)

// Register registers all Storage activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entstorage.NewClient(entstorage.Driver(driver), entstorage.AlternateSchema(entstorage.DefaultSchemaConfig()))
	bucket.Register(w, configService, entClient, limiter)
	bucketiam.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPStorageWorkflow)
}
