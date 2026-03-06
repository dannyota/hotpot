package container

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/container/cluster"
	"entgo.io/ent/dialect"
	entcontainer "danny.vn/hotpot/pkg/storage/ent/gcp/container"
)

// Register registers all Container Engine activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entcontainer.NewClient(entcontainer.Driver(driver), entcontainer.AlternateSchema(entcontainer.DefaultSchemaConfig()))
	// Register sub-packages with config service
	cluster.Register(w, configService, entClient, limiter)

	// Register container workflow
	w.RegisterWorkflow(GCPContainerWorkflow)
}
