package run

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/run/revision"
	"danny.vn/hotpot/pkg/ingest/gcp/run/service"
	"entgo.io/ent/dialect"
	entrun "danny.vn/hotpot/pkg/storage/ent/gcp/run"
)

// Register registers all Cloud Run activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entrun.NewClient(entrun.Driver(driver), entrun.AlternateSchema(entrun.DefaultSchemaConfig()))
	service.Register(w, configService, entClient, limiter)
	revision.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPRunWorkflow)
}
