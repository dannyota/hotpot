package run

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/run/revision"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/run/service"
	"entgo.io/ent/dialect"
	entrun "github.com/dannyota/hotpot/pkg/storage/ent/gcp/run"
)

// Register registers all Cloud Run activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entrun.NewClient(entrun.Driver(driver), entrun.AlternateSchema(entrun.DefaultSchemaConfig()))
	service.Register(w, configService, entClient, limiter)
	revision.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPRunWorkflow)
}
