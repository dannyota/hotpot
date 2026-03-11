package serviceusage

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/serviceusage/enabledservice"
	"entgo.io/ent/dialect"
	entserviceusage "danny.vn/hotpot/pkg/storage/ent/gcp/serviceusage"
)

// Register registers all Service Usage activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entserviceusage.NewClient(entserviceusage.Driver(driver), entserviceusage.AlternateSchema(entserviceusage.DefaultSchemaConfig()))
	enabledservice.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPServiceUsageWorkflow)
}
