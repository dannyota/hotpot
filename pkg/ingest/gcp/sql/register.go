package sql

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/sql/instance"
	"entgo.io/ent/dialect"
	entgcpsql "github.com/dannyota/hotpot/pkg/storage/ent/gcp/sql"
)

// Register registers all Cloud SQL activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entgcpsql.NewClient(entgcpsql.Driver(driver), entgcpsql.AlternateSchema(entgcpsql.DefaultSchemaConfig()))
	// Register sub-packages with config service
	instance.Register(w, configService, entClient, limiter)

	// Register SQL workflow
	w.RegisterWorkflow(GCPSQLWorkflow)
}
