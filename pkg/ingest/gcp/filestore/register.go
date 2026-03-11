package filestore

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/filestore/instance"
	"entgo.io/ent/dialect"
	entfilestore "danny.vn/hotpot/pkg/storage/ent/gcp/filestore"
)

// Register registers all Filestore activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entfilestore.NewClient(entfilestore.Driver(driver), entfilestore.AlternateSchema(entfilestore.DefaultSchemaConfig()))
	instance.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPFilestoreWorkflow)
}
