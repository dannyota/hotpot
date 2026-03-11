package containeranalysis

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/containeranalysis/note"
	"danny.vn/hotpot/pkg/ingest/gcp/containeranalysis/occurrence"
	"entgo.io/ent/dialect"
	entcontaineranalysis "danny.vn/hotpot/pkg/storage/ent/gcp/containeranalysis"
)

// Register registers all Container Analysis activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entcontaineranalysis.NewClient(entcontaineranalysis.Driver(driver), entcontaineranalysis.AlternateSchema(entcontaineranalysis.DefaultSchemaConfig()))
	note.Register(w, configService, entClient, limiter)
	occurrence.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPContainerAnalysisWorkflow)
}
