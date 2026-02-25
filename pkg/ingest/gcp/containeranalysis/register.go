package containeranalysis

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis/note"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis/occurrence"
	"entgo.io/ent/dialect"
	entcontaineranalysis "github.com/dannyota/hotpot/pkg/storage/ent/gcp/containeranalysis"
)

// Register registers all Container Analysis activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entcontaineranalysis.NewClient(entcontaineranalysis.Driver(driver), entcontaineranalysis.AlternateSchema(entcontaineranalysis.DefaultSchemaConfig()))
	note.Register(w, configService, entClient, limiter)
	occurrence.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPContainerAnalysisWorkflow)
}
