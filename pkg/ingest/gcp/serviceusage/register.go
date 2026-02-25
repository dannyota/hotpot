package serviceusage

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/serviceusage/enabledservice"
	"entgo.io/ent/dialect"
	entserviceusage "github.com/dannyota/hotpot/pkg/storage/ent/gcp/serviceusage"
)

// Register registers all Service Usage activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entserviceusage.NewClient(entserviceusage.Driver(driver), entserviceusage.AlternateSchema(entserviceusage.DefaultSchemaConfig()))
	enabledservice.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPServiceUsageWorkflow)
}
