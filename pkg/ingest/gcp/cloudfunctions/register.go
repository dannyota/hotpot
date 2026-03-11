package cloudfunctions

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/cloudfunctions/function"
	"entgo.io/ent/dialect"
	entcloudfunctions "danny.vn/hotpot/pkg/storage/ent/gcp/cloudfunctions"
)

// Register registers all Cloud Functions activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entcloudfunctions.NewClient(entcloudfunctions.Driver(driver), entcloudfunctions.AlternateSchema(entcloudfunctions.DefaultSchemaConfig()))
	function.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPCloudFunctionsWorkflow)
}
