package appengine

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/appengine/application"
	"danny.vn/hotpot/pkg/ingest/gcp/appengine/appservice"
	"entgo.io/ent/dialect"
	entappengine "danny.vn/hotpot/pkg/storage/ent/gcp/appengine"
)

// Register registers all App Engine activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entappengine.NewClient(entappengine.Driver(driver), entappengine.AlternateSchema(entappengine.DefaultSchemaConfig()))
	application.Register(w, configService, entClient, limiter)
	appservice.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPAppEngineWorkflow)
}
